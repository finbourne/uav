# UAV (the golang version)
> This is a blant rip off of @dan's work; see the c++ directory
only in golang.

Dan's original code was brilliant, and did the job of merging pipelines together in a way that I don't think had been considered before.  So Dan did all the mental heavy lifting.

## Example usage
The following command will take the pipeline defined in `my.pipeline.yaml` and output the result to `stdout`.  If you wanted to output to a file, supply a filename in place of the `-` to the `-o` flag.
```yaml
$ uav merge -p my.pipeline.yaml -o-
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
* Note that the pipeline is located in the current directory `.` and it refereneces files in a subdirectory.

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
And the output from that should be:
```yaml
$ uav merge -p my.pipeline.yaml -o-
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

* subsequent merges still use the working directory as the source for templates.  In this case, the template for the resource was not in `jobs/resources/`, but instead was in `resources/`

# Restrictions & Known Bugs
* There is currently no way of using the templating engines `template` function.  Arguments would need to be added to pull in additional named template in order to fulfil this condition were it to become a requirement.
* I'm sure there are quirks.  Find them, and we can fix them.
* There are occassions that the ranging doesn't work.  Haven't figured out why.  It's deterministic, it's just my lack of understanding around something there.