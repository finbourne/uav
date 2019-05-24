//go:generate ./version.sh
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/finbourne/uav/golang/pkg/pipeline"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
}

var (
	app          = kingpin.New("uav", "A commandline app for composing Concourse-CI pipelines.")
	merge        = app.Command("merge", "Take the pipeline and merge all the templates into it.")
	pipelineFile = merge.Flag("pipeline", "Name of file containing the pipeline to process.").Required().Short('p').File()
	templateDirs = merge.Flag("directory", "A directory containing additional Go templates to parse and make available to pipelines.").Short('d').ExistingDirs()
	templates    = merge.Arg("template", "An additional Go template to parse and make available to pipelines.").ExistingFiles()
	verbose      = app.Flag("verbose", "Verbose output.").Short('v').Bool()
	jsonVerbose  = app.Flag("json", "Verbose output in JSON format - use in combination with '--verbose'.").Short('j').Bool()

	outputFile = merge.Flag("output", "The file to save the output to.").Short('o').String()
)

func main() {
	kingpin.CommandLine.Help = "UAV - the Concourse Pipeline generator"
	kingpin.Version(version)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		logrus.SetLevel(logrus.InfoLevel)

		if *jsonVerbose {
			logrus.SetFormatter(new(logrus.JSONFormatter))
		}
	}

	switch command {
	case merge.FullCommand():
		pipeline, err := ioutil.ReadFile((*pipelineFile).Name())
		if err != nil {
			logrus.Fatalf("Error reading pipeline file: %v", err)
		}

		output, err := performMerge(string(pipeline), *templates, *templateDirs)
		if err != nil {
			logrus.Fatalf("Error creating new pipeline: %v", err)
		}

		if *outputFile == "-" || *outputFile == "" {
			_, err = os.Stdout.WriteString(output)
		} else {
			err = ioutil.WriteFile(*outputFile, []byte(output), 0644)
		}

		if err != nil {
			logrus.Fatalf("Error writing output: %v", err)
		}

	default:
		os.Exit(1)
	}
}

func performMerge(inputPipeline string, templates []string, templateDirs []string) (string, error) {
	var err error

	if len(templateDirs) > 0 {
		//If the user has specified directories containing template files,
		//these are combined with the template files individually specified (if any)
		templates, err = combineTemplates(templates, templateDirs)
		if err != nil {
			return "", fmt.Errorf("combining template files and template directories: %v", err)
		}
	}

	pl, err := pipeline.NewPipeline(inputPipeline, nil, templates)
	if err != nil {
		return "", fmt.Errorf("transforming pipeline file: %v", err)
	}

	pl, err = pl.Transform()
	if err != nil {
		return "", err
	}

	return pl.String(), nil
}

// combineTemplates returns a slice which is the superset of the templates slice and the file paths of
// all files contained in directories rooted at the directories specified in the templateDirs slice
func combineTemplates(templates []string, templateDirs []string) ([]string, error) {
	directoryTemplates, err := getDirectoryTemplates(templateDirs)
	if err != nil {
		return nil, err
	}

	return append(templates, directoryTemplates...), nil
}

// getDirectoryTemplates recurses through the directory tree rooted at each element of slice templateDirs
// and adds each file path to the returned slice
func getDirectoryTemplates(templateDirs []string) ([]string, error) {
	// At this point, kingpin has already ensured that the templateDirs are
	// actually directories and not any other kind of file

	var templateFiles []string

	// Callback function called when each directory entry is visited by the walker
	walkFunc := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("reading file %s: %v", currentPath, err)
		}

		if !info.IsDir() {
			templateFiles = append(templateFiles, currentPath)
		}

		return nil
	}

	for _, templateDir := range templateDirs {
		err := filepath.Walk(templateDir, walkFunc)
		if err != nil {
			return nil, fmt.Errorf("recursing directory tree rooted at %s: %v", templateDir, err)
		}
	}

	return templateFiles, nil
}
