package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

const makefileTemplate = `
.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
{{if .shadow}}	shadow ./...{{end}}
.PHONY:vet

build: vet
	go build
.PHONY:build
`

func main() {
	t := flag.Bool("test", false, "Adds test to makefile")
	b := flag.Bool("bench", false, "Adds bench to makefile")
	s := flag.Bool("shadow", false, "Adds shadow to makefile")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Expected use: maker DIRNAME")
		os.Exit(1)
	}
	dirName := flag.Arg(0)

	templ := template.Must(template.New("makefile").Parse(makefileTemplate))

	var buffer bytes.Buffer
	err := templ.Execute(&buffer, map[string]interface{}{
		"test":   *t,
		"bench":  *b,
		"shadow": *s,
	})
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"Makefile", buffer.Bytes(), 0744)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"main.go", []byte(`
package main

func main() {
}
`), 0744)
	if err != nil {
		panic(err)
	}
}
