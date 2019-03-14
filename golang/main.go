//go:generate ./version.sh
package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/finbourne/uav/golang/pkg/pipeline"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	app          = kingpin.New("uav", "A commandline app for composing Concourse-ci pipelines.")
	merge        = app.Command("merge", "Take the pipeline and merge all the templates into it.")
	pipelineFile = merge.Flag("pipeline", "Name of file containing the pipeline to process.").Required().Short('p').File()
	templates    = merge.Flag("template", "An additional golang text/template to parse and make available to pipelines.").Short('t').ExistingFiles()

	outputFile = merge.Flag("output", "The file to save the output to.").Short('o').String()
)

func main() {
	kingpin.CommandLine.Help = "UAV - the Concourse Pipeline generator"
	kingpin.Version(version)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case merge.FullCommand():
		p, err := ioutil.ReadFile((**pipelineFile).Name())
		if err != nil {
			log.Fatalf("Error reading pipeline file: %v", err)
		}

		pl, err := pipeline.NewPipeline(string(p), nil, *templates)
		if err != nil {
			log.Fatalf("Error transforming pipeline file: %v", err)
		}

		output := pl.Transform().String()

		if *outputFile == "-" {
			os.Stdout.WriteString(output)
		} else {
			err = ioutil.WriteFile(*outputFile, []byte(output), 0644)
		}
		if err != nil {
			log.Fatalf("Error writing new pipeline: %v", err)
		}
	default:
		os.Exit(1)
	}
}
