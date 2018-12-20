# uav
Unmanned Air Vehicle - the self-flying machine.

# What is `uav`?
We use concourse to manage our build and deployment pipeline. As our pipeline has grown larger and more complex, we've found a few points of frustration that have led us to propose some changes to the way `fly` reads and configures pipelines - mainly revolving around ahead-of-time plan and task substitution, combined with parameterisation.

This repository contains the source to a proof-of-concept tool, `uav`, which has been developed to demonstrate what we would like to see in `fly`. These are by no means demands, we'd just like to contribute our part to concourse!

Although this tool is currently written in C++, we're happy to rewrite this in go, and submit a pull-request. We're currently stuck on an older version of concourse (3.9.2), so the decision was made to write this as a stand-alone pre-processor.

## Using `uav`
Please refer to [uav.md](./uav.md) to see how to build and use `uav`.

## Ahead of time templating
Our biggest issue with operating "by the book" with concourse revolves around three main issues:

1. There is a disparity between what pipeline has currently been deployed, and what the git repository contains for pipeline utilities.
1. When deploying to different environments, a lot of code needs to be repeated, even when trying to utilise best practices with out-of-band tasks (i.e, tasks that are store in a resource)
1. Out-of-bound tasks have no automatic support for determining whether or not all of the arguments/required parameters have been set. The only way of knowing is by deploying a pipeline, and seeing if it runs successfully - even this is not guarantee, either.

For example, our main pipeline file is currently 3777 lines long on it's own - and whilst this could probably be reduced somewhat, I would be genuinely surprised if we could get it below 2500 lines. This does not include any task definitions, either - summing all of the yaml in our pipeline definition repository brings this figure to 69830. Again, there is a lot of oppertunity to reduce this figure. 

We feel that this task would be made easier (and seems to be beneficial for us, so far) by using an ahead-of-time instantiation of tasks that are defined in different files, along with allowing support for parameterisation of jobs.

## Task parameterisation
Currently, if you wish to parameterisation/template a task, the task has to live in a separate location. This file is retrieved just-in-time, and allows for parameters to specified, which are then passed through to environment of the container that runs the task.

We propose that this gets injected into a generated pipeline, following a syntax like this:

```yaml
- get: "product-src"
  passed:
    - "product.unit-test"
- task:
  name: "product.build"
  template: "some-template.yml"
  arguments:
    src_path: "product-src"
    bin_path: "product-bin"
    build_configuration: "Release"
- put: "product-bin"
```

Where `some-template.yml` looks similar to an out-of-band task;

```yaml
platform: linux
image_resource:
  type: docker-image
  source:
    repository: some-docker-image-source
inputs:
- name: "{{product_src}}"
  path: "build-src"
caches:
- name: "{{product_build}}-cache"
  path: "build-cache"
outputs:
- name: "{{product_bin}}"
  path: "build-bin"
```

The only significant difference between the current scheme and this, is that the variables in enclosed in double-curly-braces would be substituted immediately, at the command line, before the final pipeline configuration is submitted to ATC. 

Additionally, we think it would be beneficial to substitute the build script at this point, too:

```yaml
platform: "linux"
image_resource:
  # as above
template: "build.sh"
interpreter: "/bin/bash"
arguments:
  build_configuration: "{{build-configuration}}"
```

Where `build.sh` looks something like:

```bash
#!/usr/bin/env bash
set -ce

dotnet restore "build-src" \
  --packages "build-cache"

dotnet build "build-src"   \
  --packages "build-cache" \
  --configuration "{{build-configuration}}"
```

When calling `set-pipeline` from fly, the final pipeline submitted to ATC would look something like this:

```yaml
- get: "product-src"
- task: "product.build"
  config:
    image_resource:
      type: docker-image
      source:
        repository: some-docker-image-source
    inputs:
    - name: "product-src"
      path: "build-src"
    caches:
    - name: "product-src-cache"
      path: "build-cache"
    outputs:
    - name: "product-bin"
      path: "build-bin"
    run:
      path: /bin/bash
      args:
      - "-c"
      - |
        #!/usr/bin/env bash
        set -ce

        dotnet restore "build-src" \
          --packages "build-cache"

        dotnet build "build-src"   \
          --packages "build-cache" \
          --configuration "Release"
- put: "product-bin"
```

## Job parameterisation
In a similar fashion, we think it would be beneficial to be able to move jobs into their own files, following the same approach as tasks. If we look at the above example, we could additionally parameterise our product - in the cases where they have the same build process:

```yaml
- name: "product-1.build"
  serial: true
  template: jobs/product-build.yaml
  arguments:
    product_name: "product-1"
    build_configuration: "Release"

- name: "product-2.build"
  serial: true
  template: jobs/product-build.yaml
  arguments:
    product_name: "product-2"
    build_configuration: "Release"
```

If we change our initial task definition to:

```yaml
- get: "{{product_name}}-src"
  passed:
    - "{{product_name}}.unit-test"
- task: "{{product_name}}.build"
  template: "tasks/some-template.yml"
  arguments:
    src_path: "{{product_name}}-src"
    bin_path: "{{product_name}}-bin"
    build_configuration: "{{build_configuration}}"
- put: "{{product_name}}-bin"
```

This allows us to share the whole job group between two pipelines. This is a very common use-case for us, and is proving to be useful.

Note that variables are not passed down the next template-layer - this is to reduce the cognitive load of tracking of these variables.

### Early failure in parameter substitution
Additionally, it is considered an error to define a variable in a template that has not been set. This can be determined at pipeline generation time. In a more general sense, failing early when possible has always been a desirable feature for us.

## Merging of multiple pipelines
Finally, we also find it useful to have pipelines defined in potentially multiple files, but ultimately get merged into a single pipeline definition. Our biggest use-case for this is being able to define the resources for a particular portion of jobs, that are related to the jobs at hand.

The proposal for this is to simply merge each pipeline file's `groups`, `jobs`, `resources`, and `resource_types`. In the case of `resources` and `resources_types`, allowing for duplicate definitions of these items only if their definitions are identical.

For example, with two pipeline files:

```yaml
#pipeline1.yml
- groups:
  - name: group-1
    jobs:
      - job-1
      - job-2

- jobs:
  #definition of job1, job2

- resources:
  - name: resource-1
    #etc.
```

And, 

```yaml
#pipeline2.yaml
- groups:
  - name: group-2
    jobs:
      - job-3
      - job-4

- jobs:
  #definition of job3, job4

- resources:
  - name: resource-1
    #etc.
```

Will be merged into a single pipeline:

```yaml
- groups:
  - name: group-1
    jobs:
      - job-1
      - job-2
  - name: group-2
    jobs:
      - job-3
      - job-4
      
- jobs:
  #jobs 1-4

- resources:
  - name: resource-1
    #etc.
```

## Job group parameterisation
Finally, we can use this tool to parameterise whole pipeline sections. We allow for `uav` to take command-line arguments for its variable substitutions. By doing this, we can generate sections of our pipeline at a time, and then use a final invocation of `uav` to merge these pipelines together.

The resulting pipeline can be quite.. large. However, we have a lot of moving pieces in our pipeline, and they all interact in various ways with each other.

## Backward compatibility
None of the changes proposed should impact any pipeline configurations as they stand - it is possible to use a mix of the two styles without any issue. The additional configuration syntax should be entirely orthogonal to the current scheme.

# Feedback
We'd love to see this get into `master`, and we're absolutely open to suggestions, if there is something the concourse team aren't particularly happy with. Feel free to leave issues on this repository, contact myself  - <dan.curran@finbourne.com>, or our [engineering address](mailto:engineering@finbourne.com).

We look forward to hearing your feedback :)
