// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates version.go. It can be invoked by running:
// go generate
package main

import (
	"log"
	"os"
	"text/template"
	"time"
)

func main() {
	appVersion, isSet := os.LookupEnv("version")
	if !isSet {
		log.Fatalf("Environment variable 'version' is not set")
	}

	f, err := os.Create("version.go")
	die(err)
	defer f.Close()

	err = packageTemplate.Execute(f, struct {
		Timestamp time.Time
		Version   string
	}{
		Timestamp: time.Now(),
		Version:   appVersion,
	})
	die(err)
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var packageTemplate = template.Must(template.New("").Parse(`/*
* CODE GENERATED AUTOMATICALLY @ {{ .Timestamp }}
* THIS FILE SHOULD NOT BE EDITED BY HAND
* 
* ====================================================================
* NEVER COMMIT THIS FILE TO A SOURCE REPOSITORY - it contains secrets
* ====================================================================
 */

package main

var (
	version      = "{{ .Version }}"
)
`))
