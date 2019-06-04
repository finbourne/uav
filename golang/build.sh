#!/bin/bash

# set the gopath - we should be compiling in /var/app/okta
GOPATH=$GOPATH:/var/app

export version=${version:-"0.0.1"}

set -e

echo Getting dependencies.
dep ensure --vendor-only

echo Generating secrets file
go generate

echo "Linter...."
set +e
gometalinter --debug --deadline=60s > metalinter.out
set -e

echo "Tests...."
go test -v ./... -coverprofile=test.coverage.out > tests.out
echo "send analytic results to sonar"

# echo Creating Linux binary
GOOS=linux GOARCH=386 go build -o uav
zip uav-linux-$version.zip uav

# echo Creating MacOS binary
GOOS=darwin GOARCH=386 go build -o uav
zip uav-mac-$version.zip uav

# echo Creating Windows binary
env GOOS=windows GOARCH=386 go build -o uav.exe
zip uav-windows-$version.zip uav.exe

rm uav
rm uav.exe
