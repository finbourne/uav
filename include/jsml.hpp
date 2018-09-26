#pragma once
#include <yaml-cpp/yaml.h>
#include <nlohmann/json.hpp>
#include <iostream>
using json = nlohmann::basic_json<>;

class jsml {
  static json
  from_yaml_scalar
  (const YAML::Node &node)
  {
    try {
      static const std::string kYamlIntTag    = "!!int";
      static const std::string kYamlFloatTag  = "!!float";
      static const std::string kYamlDoubleTag = "!!double";
      static const std::string kYamlBoolTag   = "!!bool";
      
      const std::string tag = node.Tag();
      if (kYamlIntTag == tag)
        return node.as<int>();

      if (kYamlFloatTag == tag)
        return node.as<float>();

      if (kYamlDoubleTag == tag)
        return node.as<double>();

      if (kYamlBoolTag == tag)
        return node.as<bool>();

      const std::string str = node.as<std::string>();
      
      if (str == "true")
        return true;
      if ( str == "false")
        return false;

      //at this point, we can only have a string. probably.
      return str;
    }
    catch(const std::exception *e)
    {
      throw e;
    }
  }

  static json
  from_yaml_sequence
  (const YAML::Node &node)
  {
    try {
      std::vector<json> result;
      for (const auto &elt: node)
        result.emplace_back(from_yaml(elt));
      return result;
    }
    catch (const std::exception &e)
    {
      throw e;
    }

  }

  static json
  from_yaml_map
  (const YAML::Node &node)
  {
      std::map<std::string, json> result;
      for (const auto &elt: node)
      {
        try {
            result.emplace(elt.first.as<std::string>(), from_yaml(elt.second));
        }
        catch (const std::exception &e)
        {
          std::cout  << elt.first;
        }
      }
      return result;
  }

  public:
  static json
  from_yaml
  (const YAML::Node &node)
  {
    try {
      switch(node.Type())
      {
        case YAML::NodeType::Scalar:
          return from_yaml_scalar(node);

        case YAML::NodeType::Sequence:
          return from_yaml_sequence(node);

        case YAML::NodeType::Map:
          return from_yaml_map(node);

        case YAML::NodeType::Undefined:
          return { };

        case YAML::NodeType::Null:
          return NULL;
      }
    }
    catch (const std::exception &e)
    {
      throw e;
    }
  }
};
