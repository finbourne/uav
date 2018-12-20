package pipeline

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestTransformSimpleJob(t *testing.T) {
	p := `
merge:
- template: test.d/job_simple.yaml
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy
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
          echo Hello World!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTemplateTransform(t *testing.T) {
	p := `
merge:
{{ range .env}}
- template: test.d/job_template.yaml
  args:
    env: {{ . }}
{{ end }}
`

	args := make(map[string]interface{})
	args["env"] = []string{"ci", "qa"}

	expectedPipeline := `
jobs:
- name: deploy-ci
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
          echo Hello ci!
- name: deploy-qa
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
          echo Hello qa!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeWithNoTemplate(t *testing.T) {
	p := `
jobs:
- name: deploy
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
          echo Hello World!
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy
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
          echo Hello World!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeSimpleJobWithParamsWorksWithSecrets(t *testing.T) {
	p := `
merge:
- template: test.d/job_with_secret.yaml
  args:
    param1: Hello
    param2: World!
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy
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
          echo Hello World!
          echo ((secret))
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeSimpleJobWithGroups(t *testing.T) {
	p := `
groups:
- name: All
  jobs:
  - deploy
- name: Test
  jobs:
  - deploy
merge:
- template: test.d/job_simple.yaml
  args:
    param1: Hello
    param2: World!
  groups:
  - All
  - Test
`
	args := make(map[string]interface{})
	expectedPipeline := `
groups:
- name: All
  jobs:
  - deploy
- name: Test
  jobs:
  - deploy
jobs:
- name: deploy
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
          echo Hello World!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeSimpleJobWithResources(t *testing.T) {
	p := `
merge:
- template: test.d/job_simple.yaml
  args:
    param1: Hello
    param2: World!
- template: test.d/resource_simple.yaml
  args:
    param1: github
`
	args := make(map[string]interface{})
	expectedPipeline := `
resources:
- name: test
  type: git
  source:
    uri: git@github.com:concourse/concourse.git
    branch: master
    private_key: ((github.privatekey))

jobs:
- name: deploy
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
          echo Hello World!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTransformSubTemplate(t *testing.T) {
	p := `
merge:
- template: test.d/sub_template.yaml
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy
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
          echo Hello World!
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTransformBackslashQuote(t *testing.T) {
	p := `
merge:
- template: test.d/job_backslash_quote.yaml
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy-sonar
  plan:
  - aggregate:
    - get: thing
  - config:
      image_resource:
        source:
          repository: alpine
        type: docker-image
      inputs:
      - name: thing
      platform: linux
      run:
        args:
        - -cel
        - |
          cat >.extra-values.yaml <<EOF
          postgresql:
            postgresPassword: ${password}
          EOF
        path: /bin/bash
    task: Do Something
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTransformWithOneExtraTemplate(t *testing.T) {
	p := `
merge:
- template: test.d/job_use_template.yaml
  args:
    pass_counter:
    - deploy-all
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: deploy-sonar
  plan:
  - aggregate:
    - get: counter-fbn-prod
      trigger: true
      passed:
      - deploy-all
  - config:
      image_resource:
        source:
          repository: alpine
        type: docker-image
      inputs:
      - name: thing
      platform: linux
      run:
        args:
        - -cel
        - |
          cat >.extra-values.yaml <<EOF
          postgresql:
            postgresPassword: ${password}
          EOF
        path: /bin/bash
    task: Do Something
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, []string{"test.d/t1.tpl"})
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTransformSprigFunctionWorks(t *testing.T) {
	p := `
jobs:
  - name: test
{{ print "body:" | indent 4}}
{{ print "- blah" | indent 6}}
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: test
  body: 
    - blah
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, nil)
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestTransformIndentOfTemplateWorks(t *testing.T) {
	p := `
jobs:
  - name: test
    plan:
{{ include "test-indent" "" | indent 4 }}
`
	args := make(map[string]interface{})
	expectedPipeline := `
jobs:
- name: test
  plan:
  - get: test
    trigger: true
`
	var pipeline Pipeline

	yaml.Unmarshal([]byte(expectedPipeline), &pipeline)
	expected, _ := yaml.Marshal(&pipeline)
	merger, _ := NewPipeline(p, args, []string{"test.d/t2.tpl"})
	result := merger.Transform().String()
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}
