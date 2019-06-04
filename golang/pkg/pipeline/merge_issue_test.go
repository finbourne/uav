package pipeline

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestTransformIssue(t *testing.T) {
	p := `
merge:
- template: test.d/group1.yaml
- template: test.d/group2.yaml
- template: test.d/group3.yaml
`
	args := make(map[string]interface{})
	expectedPipeline := `
groups:
- jobs:
  - job1
  - job2
  - job3
  name: blah
- jobs:
  - job1
  name: blah_de_blah
- jobs:
  - job3
  name: All
jobs:
- name: deploy1
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
          echo Hello World!
        path: /bin/bash
    task: task1
  serial: true
- name: deploy2
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
          echo Hello World!
        path: /bin/bash
    task: task1
  serial: true
- name: deploy3
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
          echo Hello World!
        path: /bin/bash
    task: task1
  serial: true
`
	pipeline := new(Pipeline)

	yaml.Unmarshal([]byte(expectedPipeline), pipeline)
	expected, _ := yaml.Marshal(pipeline)
	merger, _ := NewPipeline(p, args, nil)

	var err error
	transformedPipeline, err := merger.Transform()
	if err != nil {
		t.Errorf("Error transforming %v: %v", p, err)
	}

	result := transformedPipeline.String()

	fmt.Println(result)
	if result != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}
