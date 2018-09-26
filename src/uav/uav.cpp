#include <uav.hpp>
#include <jsml.hpp>
#include "string.hpp"
#include "failure.hpp"

using json = nlohmann::json;

namespace uav {
  static json
  add_all_except
  (const json &initial_map, const json &remaining_map, const uav::string::list except = {})
  {
    json result = initial_map;
    const std::set<std::string> exclude(except.begin(), except.end());
    for (auto i = remaining_map.begin(); i != remaining_map.end(); i++)
    if (!exclude.count(i.key()))
      result.emplace(i.key(), i.value());

    return result;
  }

  static std::string
  read_templated_file
  (const std::string &path, const std::map<std::string, std::string> &substitutions = {})
  {
    std::vector<char> v;

    FILE *fp = fopen(path.data(), "r");
    if (NULL == fp)
      throw uav::failure::with("could not open file '" + path + "' for templating.");

    char buf[1024];
    while (size_t len = fread(buf, 1, sizeof(buf), fp))
      v.insert(v.end(), buf, buf + len);
    fclose(fp);

    auto result = uav::string(v.begin(), v.end());
    for (const auto &elt: substitutions)
        result = result.replace("{{" + elt.first + "}}", elt.second);

    return result;
  }

  static const std::string
  get_directory
  (const uav::string &path)
  {
    const auto components = path.split("/");
    std::string result = "";
    for (auto i = components.cbegin(); i != (components.cend() - 1); i++)
      result += *i + "/";

    return result;
  }

  static const std::string
  get_path_relative_to
  (const std::string &path, const uav::string::list &lineage = {})
  {
    std::string result = path;
    for (auto i = lineage.crbegin(); i != lineage.crend(); i++)
    {
      const auto directory = get_directory(*i);
      if (directory == "")
        continue;

      result = directory + result;
    }

    return result;
  }

  static json
  read_input
  (const std::string &path, const std::map<std::string, std::string> &substitutions = {})
  { 
    try {
      return jsml::from_yaml(YAML::Load(read_templated_file(path, substitutions)));
    }
    catch (const std::exception &e) {
      throw uav::failure::with("whilst processing pipeline '" + path + "',");
    }
  }

  static std::map<std::string, json>
  join_groups
  (std::map<std::string, json> groups, const json &pipeline)
  {
    if (0 == pipeline.count("groups"))
      return groups;

    const json pipeline_groups = pipeline.at("groups");
    if (!pipeline_groups.is_array())
      throw uav::failure::with("group definitions is not an array");

    for (const auto &group: pipeline_groups)
    {
      if (0 == group.count("name"))
        throw uav::failure::with("group has no name");

      const std::string name = group.require<std::string>("name");
      if (0 == group.count("jobs"))
        throw uav::failure::with("group with name '" + name + "' contains no jobs");

      if (groups.count(name))
        throw uav::failure::with("group with name '" + name + "' already defined");

      groups.emplace(name, group);
    }

    return groups;
  }

  static json
  join_config_run
  (const json &config, const std::string &base_path, const std::string &plan_path, const std::string &task_path)
  {
    if (config.count("run"))
      return add_all_except(
        { },
        config,
        { "template", "arguments", "interpreter" }
      );

    if (0 == config.count("template"))
      throw uav::failure::with("task configuration contains no run command or template");

    const std::string script_template_path  = config.require<std::string>("template");
    const std::string interpreter           = config.get_alt<std::string>("interpreter", "/bin/bash");
    const std::string templated_script      = read_templated_file(
      get_path_relative_to(
        script_template_path,
        { base_path, plan_path, task_path }
      ),
      config.get_alt<std::map<std::string, std::string>>("arguments")
    );

    const json run = {
      { "path", interpreter },
      { "args", { "-c", templated_script } },
    };

    return add_all_except(
      { { "run", run } },
      config,
      { "template", "arguments", "interpreter", "run" }
    );
  }

  static json
  join_task_config
  (const json &task, const std::string &base_path, const std::string &plan_path)
  {
    if (task.count("config"))
      return join_config_run(task.at("config"), base_path, plan_path, "");

    const std::string template_path = task.require<std::string>("template");
    return join_config_run(
      read_input(
        get_path_relative_to(template_path, { base_path, plan_path }),
        task.get_alt<std::map<std::string, std::string>>("arguments")
      ),
      base_path,
      plan_path,
      template_path
    );
  }

  static json
  join_task
  (const json &task, const std::string &base_path, const std::string &plan_path)
  {
    try {
      if (task.count("config"))
        return task;

      if (task.count("file"))
        return task;

      //add every other parameter into the task
      return add_all_except(
        { { "config", join_task_config(task, base_path, plan_path) } },
        task,
        {"template", "arguments", "config"}
      );
    }
    catch (const std::exception &e)
    {
      const std::string name = task.require<std::string>("task");
      throw uav::failure::with(e, "whilst processing task '" + name + "',");
    }
  }

  static json
  join_job_steps
  (const json &steps, const std::string &base_path, const std::string &plan_path);

  static json
  join_job_step
  (const json &step, const std::string &base_path, const std::string &plan_path)
  {
    if (step.count("task"))
      return join_task(step, base_path, plan_path);
    
    if (step.count("get"))
      return step;
    
    if (step.count("put"))
      return step;
    
    if (step.count("aggregate"))
      return join_job_steps(step.at("aggregate"), base_path, plan_path);
    
    if (step.count("do"))
      return join_job_steps(step.at("do"), base_path, plan_path);
    
    if (step.count("try"))
      return join_job_step(step.at("try"), base_path, plan_path);

    throw uav::failure::with("read an unexpected plan step:\n\n" + step.dump(4) + "\n\n");
  }

  static json
  join_job_steps
  (const json &steps, const std::string &base_path, const std::string &plan_path)
  {
    if (!steps.is_array())
      throw uav::failure::with("job steps definition is not an array");

    std::vector<json> result;
    for (const auto &step: steps)
      result.emplace_back(join_job_step(step, base_path, plan_path));

    return result;
  }

  static json
  join_plan
  (const json &job, const std::string &base_path)
  {
    //if it defines a plan, just return that.
    if (job.count("plan"))
      return join_job_steps(job.at("plan"), base_path, "");

    if (0 == job.count("template"))
      throw uav::failure::with("job does not define a job plan or template");

    const std::string template_path = job.require<std::string>("template");
    return join_job_steps(
      read_input(
        get_path_relative_to(template_path, { base_path }),
        job.get_alt<std::map<std::string, std::string>>("arguments")
      ),
      base_path,
      template_path
    );
  }

  static const json
  join_job
  (const json &job, const std::string &base_path)
  {
    return add_all_except(
      { { "plan", join_plan(job, base_path) } },
      job,
      {"template", "arguments", "plan"}
    );
  }

  static std::map<std::string, json>
  join_jobs
  (std::map<std::string, json> jobs, const json &pipeline, const std::string &base_path)
  {
    if (0 == pipeline.count("jobs"))
      return jobs;

    const json pipeline_jobs = pipeline.at("jobs");
    if (!pipeline_jobs.is_array())
      throw uav::failure::with("jobs section is not a array");

    for (const auto &j: pipeline_jobs)
    {
      if (0 == j.count("name"))
        throw uav::failure::with("job has no name");

      const std::string name = j.require<std::string>("name");
      if (jobs.count(name))
        throw uav::failure::with("job with name '" + name + "' already defined");

      try {
        jobs.emplace(name, join_job(j, base_path));
      }
      catch (const std::exception &e) {
        throw uav::failure::with(e, "whilst joining plan in job '" + name + "',");
      }
    }

    return jobs;
  }

  static std::vector<json>
  values
  (const std::map<std::string, json> &map)
  {
    std::vector<json> result;
    for (const auto &elt: map)
      result.emplace_back(elt.second);

    return result;
  }

  static json
  construct_pipeline(
    const std::map<std::string, json> &groups,
    const std::map<std::string, json> &jobs,
    const std::map<std::string, json> &resources,
    const std::map<std::string, json> &resource_types
  )
  {
    return {
      { "groups",         values(groups)         },
      { "jobs",           values(jobs)           },
      { "resources",      values(resources)      },
      { "resource_types", values(resource_types) },
    };
  }

  static std::map<std::string, json>
  join_resources
  (std::map<std::string, json> resources, const json &pipeline)
  {
    if (0 == pipeline.count("resources"))
      return resources;
    
    const json pipeline_resources = pipeline.at("resources");
    if (!pipeline_resources.is_array())
      throw uav::failure::with("resources is not an array");

    for (const auto &resource: pipeline_resources)
    {
      if (0 == resource.count("name"))
        throw uav::failure::with("resource has no name");

      const auto name = resource.require<std::string>("name");

      //TODO: assuming uniqueness
      if (0 == resources.count(name))
        resources.emplace(name, resource);
    }

    return resources;
  }

  static std::map<std::string, json>
  join_resource_types
  (std::map<std::string, json> resource_types, const json &pipeline)
  {
    if (0 == pipeline.count("resource_types"))
      return resource_types;
    
    const json pipeline_resource_types = pipeline.at("resource_types");
    if (!pipeline_resource_types.is_array())
      throw uav::failure::with("resources is not an array");

    for (const auto &resource_type: pipeline_resource_types)
    {
      if (0 == resource_type.count("name"))
        throw uav::failure::with("resource_type has no name");

      const auto name = resource_type.require<std::string>("name");

      //TODO: assuming uniqueness
      if (0 == resource_types.count(name))
        resource_types.emplace(name, resource_type);
    }

    return resource_types;
  }

  static std::string
  validate_output
  (const std::string &input)
  {
    //search for any instances of '{{' - i.e, unresolved variables.
    const size_t vbegin = input.find("{{");
    if (vbegin == std::string::npos)
      return input;

    const size_t vend = input.find("}}", vbegin);
    if (vend == std::string::npos)
      return input;

    throw uav::failure::with(
      "Unresolved variable in pipeline definition: '" +
      input.substr(vbegin + 2, vend - vbegin - 2) +
      "'"
    );
  }

  std::string
  generate_pipeline
  (const uav::arguments &arguments)
  {
    //using map's here to ensure element-name uniqueness
    std::map<std::string, json> groups;
    std::map<std::string, json> jobs;
    std::map<std::string, json> resources;
    std::map<std::string, json> resource_types;
    for (const auto &path: arguments.input_paths)
    {
      try {
        const json input    = read_input(path, arguments.substitutions);
        groups              = join_groups(groups, input);
        jobs                = join_jobs(jobs, input, path);
        resources           = join_resources(resources, input);
        resource_types      = join_resource_types(resource_types, input);
      }
      catch (const std::exception &e) {
        throw uav::failure::with(e, "whilst processing pipeline '" + path + "',");
      }
    }

    const auto pipeline = construct_pipeline(groups, jobs, resources, resource_types);
    return validate_output(pipeline.dump(2));
  }
}
