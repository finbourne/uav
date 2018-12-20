package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/Masterminds/sprig"
	yaml "gopkg.in/yaml.v2"
)

// Pipeline is the piepline definition.  Added `merge` directive.
type Pipeline struct {
	Merge          []interface{} `yaml:"merge,omitempty"`
	Groups         []interface{} `yaml:"groups,omitempty"`
	Resources      []interface{} `yaml:"resources,omitempty"`
	ResourceTypes  []interface{} `yaml:"resource_types,omitempty"`
	Jobs           []interface{} `yaml:"jobs,omitempty"`
	extraTemplates []string
}

type mergeConfig struct {
	FilePath   string                 `yaml:"template"`
	Parameters map[string]interface{} `yaml:"args,omitempty"`
}

// NewPipeline constructs a merger object for merging pipelines.
func NewPipeline(pipeline string, args map[string]interface{}, templates []string) (*Pipeline, error) {
	out := transformTemplateWithParams(args, pipeline, templates)

	var p Pipeline
	err := yaml.Unmarshal([]byte(out), &p)
	if err != nil {
		log.Printf("error unmarshalling pipeline: %v\npipeline: %v", err, out)

		return nil, err
	}
	p.extraTemplates = templates

	return &p, nil
}

// Transform takes the current pipeline and begins recursive transformation to produce the finished pipeline.
func (p *Pipeline) Transform() *Pipeline {
	var err error
	pipeline := Pipeline{
		Groups:         p.Groups,
		Resources:      p.Resources,
		ResourceTypes:  p.ResourceTypes,
		Jobs:           p.Jobs,
		extraTemplates: p.extraTemplates,
	}

	if len(p.Merge) > 0 {
		for _, v := range p.Merge {
			c := mapInterfaceInterfaceToMapStringInterface(v.(map[interface{}]interface{}))
			if mc, ok := mergeConfigFromTemplateWithParams(c); ok {
				cp := mapInterfaceInterfaceToPipeline(stringToMapInterfaceInterface(transformTemplateWithParams(mc.Parameters, getYamlMap(mc.FilePath), pipeline.extraTemplates)))
				pipeline, err = merge(pipeline, cp)
				if err != nil {
					fmt.Printf("Unable to merge pipelines %v\n", err)
				}
			}
		}

		return pipeline.Transform()
	}

	return &pipeline
}

func (p *Pipeline) String() string {
	text, err := yaml.Marshal(&p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return string(text)
}

func mergeConfigFromTemplateWithParams(data map[string]interface{}) (mergeConfig, bool) {
	if data["template"] == nil {
		return mergeConfig{}, false
	}

	var m mergeConfig
	m.FilePath = data["template"].(string)
	if data["args"] != nil {
		m.Parameters = mapInterfaceInterfaceToMapStringInterface(data["args"].(map[interface{}]interface{}))
	} else {
		m.Parameters = make(map[string]interface{})
	}
	return m, true
}

func mapInterfaceInterfaceToPipeline(data map[interface{}]interface{}) Pipeline {
	m := mapInterfaceInterfaceToMapStringInterface(data)
	pipeline := Pipeline{}
	if m["groups"] != nil {
		pipeline.Groups = m["groups"].([]interface{})
	}
	if m["jobs"] != nil {
		pipeline.Jobs = m["jobs"].([]interface{})
	}
	if m["merge"] != nil {
		pipeline.Merge = m["merge"].([]interface{})
	}
	if m["resource_types"] != nil {
		pipeline.ResourceTypes = m["resource_types"].([]interface{})
	}
	if m["resources"] != nil {
		pipeline.Resources = m["resources"].([]interface{})
	}
	return pipeline
}

func mapInterfaceInterfaceToMapStringInterface(data map[interface{}]interface{}) map[string]interface{} {
	m := make(map[string]interface{})

	for key, value := range data {
		switch key := key.(type) {
		case string:
			m[key] = value
		default:
			log.Fatal("key should be string")
		}
	}
	return m
}

func getYamlMap(filename string) string {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(yamlFile)
}

func transformTemplateWithParams(params map[string]interface{}, t string, ts []string) string {
	templates := template.New("pipeline")
	var err error
	if len(ts) > 0 {
		templates, err = templates.Funcs(funcMap(templates)).ParseFiles(ts...)
		if err != nil {
			log.Fatalf("%v\n\nTemplate: %v", err, t)
		}
	} else {
		templates = templates.Funcs(funcMap(templates))
	}
	_, err = templates.Parse(t)
	if err != nil {
		log.Fatalf("%v\n\nTemplate: %v", err, t)
	}

	buf := bytes.NewBufferString("")
	err = templates.Execute(buf, params)
	if err != nil {
		log.Fatalf("%v\n\nTemplate: %v", err, t)
	}

	return buf.String()
}

func stringToMapInterfaceInterface(data string) map[interface{}]interface{} {
	buf := bytes.NewBufferString(data)
	var snippet map[interface{}]interface{}
	err := yaml.Unmarshal(buf.Bytes(), &snippet)
	if err != nil {
		log.Fatalf("Unmarshal: %v\n%v", err, data)
	}
	return snippet

}

// ToYaml takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

func funcMap(t *template.Template) template.FuncMap {
	f := sprig.TxtFuncMap()

	// Add some extra functionality
	extra := template.FuncMap{
		"toYaml":   toYaml,
		"fromYaml": fromYaml,
		"toJson":   toJson,
		"fromJson": fromJson,
		"include": func(name string, data interface{}) (string, error) {
			buf := bytes.NewBuffer(nil)
			if err := t.ExecuteTemplate(buf, name, data); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}

	for k, v := range extra {
		f[k] = v
	}

	return f
}

func fromYaml(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

func toJson(v interface{}) string { // nolint
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

func fromJson(str string) map[string]interface{} { // nolint
	m := map[string]interface{}{}

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}
