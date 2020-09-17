package pipeline

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestMergeSimple(t *testing.T) {
	y1 := `
merge:
- template: test.d/job_simple.yaml
`
	y2 := `
merge:
- template: test.d/another_file.yaml
`
	e1 := `
merge:
- template: test.d/job_simple.yaml
- template: test.d/another_file.yaml
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeDifferentResources(t *testing.T) {
	y1 := `
resources:
- name: blah
  other: blah
`
	y2 := `
resources:
- name: blah2
  other: blah
`
	e1 := `
resources:
- name: blah2
  other: blah
- name: blah
  other: blah
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", string(merged), string(expected))
	}
}

func TestMergeSameResourcesSuccessful(t *testing.T) {
	y1 := `
resources:
- name: blah
  other: blah
`
	y2 := `
resources:
- name: blah
  other: blah
`
	e1 := `
resources:
- name: blah
  other: blah
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeSameResourcesFails(t *testing.T) {
	y1 := `
resources:
- name: blah
  other: blah
`
	y2 := `
resources:
- name: blah
  other: blah2
`
	var p1 Pipeline
	var p2 Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)

	_, ok := merge(p1, p2)
	if ok == nil {
		t.Errorf("Merging 2 resources with same name that are different should fail; %v", ok)
	}
}

func TestMergeGroupsNoOverlap(t *testing.T) {
	y1 := `
groups:
- name: blah
  jobs: 
  - job1
`
	y2 := `
groups:
- name: blah_de_blah
  jobs: 
  - job2
`
	e1 := `
groups:
- name: blah
  jobs: 
  - job1
- name: blah_de_blah
  jobs: 
  - job2
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result, string(expected))
	}
}

func TestMergeGroupsWithOverlappingGroupNames(t *testing.T) {
	y1 := `
groups:
- name: blah
  jobs: 
  - job1
`
	y2 := `
groups:
- name: blah
  jobs: 
  - job2
`
	e1 := `
groups:
- name: blah
  jobs: 
  - job1
  - job2
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result.String(), string(expected))
	}
}

func TestMergeGroupsWithOverlappingGroupNamesAndJobs(t *testing.T) {
	y1 := `
groups:
- name: blah
  jobs: 
  - job1
- name: blah_de_blah
  jobs: 
  - job1
`
	y2 := `
groups:
- name: blah
  jobs: 
  - job2
`
	e1 := `
groups:
- name: blah
  jobs: 
  - job1
  - job2
- name: blah_de_blah
  jobs: 
  - job1
`
	var p1 Pipeline
	var p2 Pipeline
	var ep Pipeline

	yaml.Unmarshal([]byte(y1), &p1)
	yaml.Unmarshal([]byte(y2), &p2)
	yaml.Unmarshal([]byte(e1), &ep)
	expected, _ := yaml.Marshal(&ep)

	result, _ := merge(p1, p2)
	merged, _ := yaml.Marshal(&result)
	if string(merged) != string(expected) {
		t.Errorf("[%v] is not equal to [%v]\n", result.String(), string(expected))
	}
}
