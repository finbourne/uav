#!/usr/bin/env -S bash -e

version=${version:-0.0.5}

cat >version.go <<- EOF
package main

var (
	version      = "${version}"
)
EOF