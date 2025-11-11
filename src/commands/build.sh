#!/bin/bash
# Build script for src-commands

# Collect all .go files in package main
FILES="main.go custom_renderers.go"
FILES="$FILES $(find impl -name '*.go' ! -name '*_test.go' ! -path '*/design/design/*' ! -path '*/docs/docs/*' ! -path '*/get/get/*' ! -path '*/templates/templates/*' ! -path '*/commit-message/*' ! -path '*/pipelinerunner/*')"

# Build 
go build -o commands.exe $FILES
