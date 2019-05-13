package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const (
	input = `merge:
- template: jobs/test.yml
  args:
    env: qa
    repo_master: github	
`
	expectedOutput = `resources:
- name: test
  source:
    branch: master
    private_key: ((github.privatekey))
    uri: git@github.com:concourse/concourse.git
  type: git
jobs:
- name: deploy-qa
  plan:
  - get: repo
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
          cd repo
          echo Hello qa!
        path: /bin/bash
    task: task1
  serial: true
`
)

type testCase struct {
	workingDir string
	dirs       []string
	files      []string
}

func (tc testCase) String() string {
	return fmt.Sprintf("Working dir: %s; Directories: %v; Files: %v", tc.workingDir, tc.dirs, tc.files)
}

func TestPerformMerge(t *testing.T) {
	err := os.Chdir("testdata")
	if err != nil {
		t.Fatalf("Unable to change directory for tests: %v", err)
	}

	tests := []testCase{
		{
			// Test case for when no directories are provided
			workingDir: "tpl_files_only",
			files:      []string{"jobs/test.yml", "repo.yml"},
		},
		{
			// Test case for when no files are provided
			workingDir: "dirs_only",
			dirs:       []string{"jobs"},
		},
		{
			// Test case for when both files and directories are provided
			workingDir: "tpl_files_and_dirs",
			files:      []string{"repo.yml"},
			dirs:       []string{"jobs"},
		},
		{
			// Test case for when multiple directories are provided
			workingDir: "multi_dirs",
			dirs:       []string{"jobs", "resources"},
		},
		{
			// Test case for when an empty directory is provided
			workingDir: "empty_dir",
			dirs:       []string{"jobs", "empty"},
		},
		{
			// Test case for when a directory containing nested directories is provided
			// and a file contained in the nested directory is referenced in a template
			workingDir: "nested_dir",
			dirs:       []string{"jobs"}, //repo.yml is contained in jobs/resources
		},
	}

	for _, test := range tests {
		doTest(t, test)
	}
}

func doTest(t *testing.T, test testCase) {
	err := os.Chdir(test.workingDir)
	if err != nil {
		t.Errorf("Unable to change directory for %v: %v", test, err)
		return
	}

	defer os.Chdir("..")

	currentWorkingDir, _ := os.Getwd()
	log.Printf("Running test [%v] in directory [%v]", test, currentWorkingDir)

	output, err := performMerge(input, test.files, test.dirs)
	if err != nil {
		t.Errorf("Error encountered for %v: %v", test, err)
		return
	}

	if output != expectedOutput {
		t.Errorf("Incorrect output:\n%s\nfor %v", output, test)

		info, err := os.Stat("incorrect_output")
		if err != nil {
			t.Fatalf("Unable to stat incorrect_output directory for %v: %v", test, err)
		}

		if info != nil && !info.IsDir() {
			t.Fatalf("File incorrect_output exists, must be directory for %v: %v", test, err)
		}

		if info == nil {
			err = os.Mkdir("incorrect_output", 0755)
			if err != nil {
				t.Fatalf("Unable to create incorrect_output directory for %v: %v", test, err)
			}
		}

		err = ioutil.WriteFile("incorrect_output"+string(os.PathSeparator)+"pipeline.yml", []byte(output), 0644)
		if err != nil {
			t.Fatalf("Unable to create incorrect_output file for %v: %v", test, err)
		}
	}
}
