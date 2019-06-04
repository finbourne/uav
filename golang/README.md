# UAV (the golang version)
> This is a blatant rip off of @dan's work; see the c++ directory
only in golang.

Dan's original code was brilliant, and did the job of merging pipelines together in a way that I don't think had been considered before.  So Dan did all the mental heavy lifting.

## Example usage
The following command will take the pipeline defined in `my.pipeline.yaml` and output the result to `stdout`.  If you wanted to output to a file, redirect the output or use the `-o` flag.

`uav merge -p my.pipeline.yaml`

```yaml
jobs:
- name: deploy-ci
  plan:
  - config:
      image_resource:
        source:
          repository: test/docker-container
        type: docker-image
      platform: linux
      run:
        args:
        - -cel
        - |
          echo Hello ci!
        path: /bin/bash
    task: task1
  serial: true
- name: deploy-qa
  plan:
  - config:
      image_resource:
        source:
          repository: test/docker-container
        type: docker-image
      platform: linux
      run:
        args:
        - -cel
        - |
          echo Hello qa!
        path: /bin/bash
    task: task1
  serial: true

```
`my.pipeline.yaml`
```yaml
merge:
- template: jobs/test.yaml
  args:
    envs:
    -
      env: ci
    -
      env: qa
```
`jobs/test.yaml`
```yaml
jobs:
{{ range .envs }}

- name: deploy-{{ .env }}
  serial: true
  plan:
  - task: task1
    config:
      platform: linux

      image_resource:
        type: docker-image
        source:
          repository: test/docker-container
      run:
        path: /bin/bash
        args:
        - -cel
        - |
          echo Hello {{ .env }}!

{{- end }}
```
To explain all that:
* One new pipeline construct has been added; `merge`.  
This will be evaluated recursively, so you can add a `merge` that references your resources from each of your pipelines.
* Arguments can be passed to the templates in `merge` via the `args` map.  
`args` is a yaml map.  In the example above, it has one entry `envs` which is itself an array of maps.
* All the power of golang text/template is at your fingertips.  
So things like loops `{{ range }}` and if's `{{ if }}` are available to use in the pipelines.  
See https://golang.org/pkg/text/template/ for further information.
* It is worth noting that the yaml that goes in may not resemble the yaml that comes out.  This is because yaml maps aren't order specific, so when processed, end up being printed out in alphabetical order.
* Note that the pipeline is located in the current directory `.` and it references files in a subdirectory.

## Another example
`my.pipeline.yaml`
```yaml
merge:
- template: jobs/test.yaml
  args:
    env: qa
    repo_master: github
```
`jobs/test.yaml`
```yaml
jobs:
- name: deploy-{{ .env }}
  serial: true
  plan:
  - get: repo
  - task: task1
    config:
      platform: linux

      image_resource:
        type: docker-image
        source:
          repository: test/docker-container
      run:
        path: /bin/bash
        args:
        - -cel
        - |
          cd repo
          echo Hello {{ .env }}!
merge:
- template: resources/repo.yaml
  args:
    repo_master: {{ .repo_master }}
```
`resources/repo.yaml`
```yaml
resources:
- name: test
  type: git
  source:
    uri: git@{{ .repo_master }}.com:concourse/concourse.git
    branch: master
    private_key: ((github.privatekey))
```
`uav merge -p my.pipeline.yaml`

And the output from that should be:
```yaml
jobs:
- name: deploy-qa
  serial: true
  plan:
  - get: repo
  - task: task1
    config:
      platform: linux

      image_resource:
        type: docker-image
        source:
          repository: test/docker-container
      run:
        path: /bin/bash
        args:
        - -cel
        - |
          cd repo
          echo Hello qa!
resources:
- name: test
  type: git
  source:
    uri: git@github.com:concourse/concourse.git
    branch: master
    private_key: ((github.privatekey))
```

* Subsequent merges still use the working directory as the source for templates.  In this case, the template for the resource was not in `jobs/resources/`, but instead was in `resources/`

# Templates
As well as the 'top-level' Concourse pipeline objects specified by the `merge` clause, snippets may be provided as Go templates. These are imported into the template using the `include` function:

`include "<template name>" [<arg1>...]`

Per standard Go templating rules, the `template name` may either be a named template (if multiple templates reside in one file) or the basename of a file containing a single unnamed template. As the basename of the file is used, it is therefore not possible to have multiple template files with the same name, even if they reside in different directories.

UAV provides two mechanisms for making these templates available for inclusion into other templates:
* Directories containing templates or nested subdirectories containing templates may be provided using the `--directory` or `-d` flag:
`--directory <dir1> [<dir2>...]`
* Individual template file(s) may be provided as arguments.

# Template Functions
In addition to the standard functions from the Go text/template package, the functions from the Sprig library (http://masterminds.github.io/sprig/) are available.

UAV also provides a few functions of its own:
* `indentSub n "text"` - where `text` is some text to be indented (often piped from an `include`) and `n` is the number of spaces all lines but the first in the `text` will be indented by. This is useful for including snippets at the correct level of indentation to ensure the result is valid YAML.
* `toYaml object` - marshall an arbitrary object (Go `struct`, `map` or `slice`) into YAML.
* `fromYaml` - unmarshall YAML into a Go `map[string]interface{}` (a map of string keys to arbitrary objects).
* `toJson` - marshall an arbitrary object (Go `struct`, `map` or `slice`) into JSON.
* `fromJson` - unmarshall JSON into a Go `map[string]interface{}` (a map of string keys to arbitrary objects).
* `skipLines n "text"` - where `text` is some text (often piped from another function) and `n` is the number of lines from the input to skip in the output.

# Example Project Layout

A typical project layout showing how UAV is used at [Finbourne](https://www.finbourne.com):

```
├── pipelines
│   ├── build
│   │   └── components.tpl
│   ├── deploy
│   │   └── components.tpl
│   └── verify
│       └── verify.tpl
├── resource_types
│   ├── semver.tpl
│   └── slack-notification.tpl
├── resources
│   ├── git
│   │   └── pipelines.tpl
│   ├── semver
│   │   └── versions.tpl
│   └── slack
│       └── slack-alert.tpl
├── pipeline.tpl
└── templates
    └── on_failure.tpl
```

This UAV project is "built" using the following invocation:

`uav merge --pipeline telemetry.pipeline.tpl --directory templates > .pipeline.yml`

* The parent directory contains the file `pipeline.ypl` which contains a single `merge` containing an array of `template`s:

```YAML
{{$COMPONENTS := `["prometheus", "grafana"]`}}
merge:
  - template: pipelines/build/components.tpl
    args:
      components: {{$COMPONENTS}}
  - template: pipelines/verify/verify.tpl
    args:
      envs:
        - name: CI
          dependencies:
            - build-telemetry-deployables
        - name: QA
          dependent_on_env: CI
        - name: PROD
          dependent_on_env: QA        
      components: {{$COMPONENTS}}
  - template: pipelines/deploy/components.tpl
    args:
      envs:
        - name: CI
        - name: QA
        - name: PROD
      components: {{$COMPONENTS}}  
```

* The `pipelines/build/components.tpl` file contains the Concourse `job` which builds the component deployables and uploads to cloud storage:

```YAML
merge:
  - template: resources/semver/versions.tpl
    args:
      semvers:
        - name: telemetry-version
          file: telemetry.version  
  - template: resources/git/pipelines.tpl

  # Used by the "on_failure.tpl" include
  - template: resources/slack/slack-alert.tpl
jobs:
  - name: build-telemetry-deployables
    plan:
    - aggregate:
      - get: pipelines
      - get: telemetry-version
        trigger: true
    - aggregate:
      {{range .components}}
      - task: Create {{. | title}} deploy zip
        config:
          platform: linux
          image_resource:
            type: docker-image
            source:
              repository: ...          
          inputs:
            - name: pipelines
            - name: telemetry-version
          params:
            component: {{.}}
          run:
            ...
      {{end}}

    {{include "on_failure.tpl" | indentSub 4}}
```

* The `resources/semver/versions.tpl` file contains the Concourse `resource` definition for a `semver` resource:

```YAML
merge:
  - template: resource_types/semver.tpl

resources:
  {{range .semvers}}
  - name: {{.name}}
    type: semver    
    source:
      driver: git
      branch: master
      uri: ...
      file: {{.file}}
      private_key: ((gitlab-private-key))
  {{end}}
```

* The `resource_types/semver.tpl` file contains the Concourse `resource_type` definition for a custom `semver` resource (this would not be required when using the built-in resource of the same name):

```YAML
resource_types:
- name: semver  
  type: docker-image
  source:
    repository: ...
```

* The `resources/git/pipelines.tpl` file contains the Concourse `resource` definition for a `git` resource:

```YAML
resources:
- name: pipelines
  type: git  
  source:
    uri: ...
    branch: master
    private_key: ((gitlab-private-key))
```

* The `resources/slack/slack-alert.tpl` file contains the Concourse `resource` definition for a `slack-notification` resource:

```YAML
merge:
  - template: resource_types/slack-notification.tpl

resources:
  - name: slack-alert    
    type: slack-notification
    source:
      url: ((slack.url))
```

* The `resource_types/slack-notification.tpl` file contains the Concourse `resource_type` definition for a third-party `slack-notification` resource:

```yaml
resource_types:
- name: slack-notification
  type: docker-image
  source:
    repository: cfcommunity/slack-notification-resource
    platform: latest
```

* The `on_failure.tpl` template is contained in file `templates/on_failure.tpl` (see discussion of Go template naming conventions above) and contains a snippet of Concourse `job` configuration - specifically, the `on_failure` clause. The same boilerplate is used for every job in the pipeline, hence it is moved into its own file:

```YAML
on_failure:
  put: slack-alert
  params:
    channel: '#build'
    icon_emoji: ':warning:'
    text: |
      HEY <!channel>! Something went wrong with the build ($BUILD_PIPELINE_NAME/$BUILD_JOB_NAME).      
```
The key thing to notice here is that a "snippet" (i.e. some re-usable boilerplate which is not a "top-level" Concourse type) is `include`d into a larger template rather than `merge`d. This necessitates the "forward-declaring" of the `resource` named `slack-alert` in the template files which `include` the `on_failure.tpl` template.

* The `pipelines/verify/verify.tpl` file contains the Concourse `job` which builds the verifies the environments are ready for the installation of the deployables:

```YAML
merge:  
  - template: resources/semver/versions.tpl
    args:
      semvers:
        - name: telemetry-version
          file: telemetry.version
  - template: resources/git/pipelines.tpl

  # Used by the "on_failure.tpl" include
  - template: resources/slack/slack-alert.tpl  
jobs:
{{range .envs}}
{{$env := .}}
  - name: verify environment [{{$env.name}}]
    plan:
      - aggregate:
        - get: pipelines    
        - get: telemetry-version
          passed:
          {{if $env.dependencies}}
          {{range $env.dependencies}}          
          - {{.}}
          {{end}}
          {{else}}          
          {{range $.components}}
          - {{printf "telemetry-%s-deploy [%s]" . $env.dependent_on_env}}
          {{end}}
          {{end}}
          trigger: true
      - aggregate:
        - task: Run verifications
          config:
            platform: linux
            image_resource:
              type: docker-image
              source:
                repository: ...                
                username: ((docker-username))
                password: ((docker-password))
            run:
              path: /var/app/start.sh
              dir: /var/app

    {{include "on_failure.tpl" | indentSub 4}}
{{end}}
```

Notice that all `resource`s are declared again, even though they were declared previously in the `pipelines/build/components.tpl` file. This does not cause an issue in the final YAML which is created as UAV will merge repeated identical definitions. The advantage of this approach is that it declares the resources needed by a job "locally", allowing larger pipelines to be composed from individual job templates - each job will "bundle" its dependencies with it.

* The `pipelines/deploy/components.tpl` file contains the Concourse `job` which installs the deployables onto the environments:

```YAML
merge:  
  - template: resources/semver/versions.tpl
    args:
      semvers:
        - name: telemetry-version
          file: telemetry.version  
  - template: resources/git/pipelines.tpl

  # Used by the "on_failure.tpl" include
  - template: resources/slack/slack-alert.tpl
jobs:
  {{range .envs}}
  {{$env := .}}
  {{range $.components}}
  - name: "telemetry-{{.}}-deploy [{{$env.name}}]"
    plan:
      - aggregate:
        - get: pipelines
        - get: telemetry-version
          trigger: true
          passed:
            - "verify environment [{{$env.name}}]"      
      - task: Deploy    
        config:
          platform: linux
          image_resource:
            type: docker-image
            source:
              repository: ...
          inputs:
            - name: telemetry-version
            - name: pipelines
          params:            
            component: {{.}}
          run:
            path: /bin/bash
            args: ...
      - task: Check application health
        file: pipelines/tasks/...
        params:        
          envName: "telemetry-{{.}}"        

    {{include "on_failure.tpl" | indentSub 4}}
{{end}}
{{end}}
```

# Restrictions & Known Bugs
* There is currently no way of using the templating engine's `template` function.  Arguments would need to be added to pull in additional named template in order to fulfil this condition were it to become a requirement.
* I'm sure there are quirks.  Find them, and we can fix them.
* There are occasions that the ranging doesn't work.  Haven't figured out why.  It's deterministic, it's just my lack of understanding around something there.
