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
	@go fmt ./...
.PHONY:fmt

lint: fmt
	@golint ./...
.PHONY:lint

vet: fmt
	@go vet ./...
{{- if .shadow}}	@shadow ./...{{end}}
.PHONY:vet

{{ if not .library}}
build: vet
	@go build
.PHONY:build

run: vet
	@go run main.go
.PHONY:run
{{ else}}
build: vet
	@go build ./...
.PHONY:build
{{end}}

{{- if .test}}
test: vet
	@go test -v ./...
.PHONY:test
{{ end }}

{{- if .bench}}
bench: vet
	@go test -v -bench=. -benchmem ./...
.PHONY:bench
{{ end }}

{{- if and .test .cover}}
test-cover: vet
	@go test -v -cover ./...
.PHONY:test-cover
{{ end }}

{{- if and .test .coverHTML}}
test-cover-html: vet
	@go test -v -cover -coverprofile=c.out ./...
	@go tool cover -html=c.out
.PHONY:test-cover-html
{{ end }}

{{- if .testRace}}
test-race: vet
	@go test -race ./...
.PHONY:test-race
{{ end }}

{{- if .race}}
build-race: vet
	@go build -race
.PHONY:build-race
{{ end }}

{{- if .cpuProfile}}
test-cpu: vet
	@go test {{if .bench}}-bench=. -benchmem{{end}} -cpuprofile cpu.out ./...
	@go tool pprof cpu.out
.PHONY:test-cpu
{{ end }}

{{- if .memProfile}}
test-mem: vet
	@go test {{if .bench}}-bench=. -benchmem{{end}} -memprofile mem.out ./...
	@go tool pprof mem.out
.PHONY:test-mem
{{ end }}
`

func main() {
	t := flag.Bool("test", false, "Adds test to makefile")
	b := flag.Bool("bench", false, "Adds bench to makefile")
	s := flag.Bool("shadow", false, "Adds shadow to makefile")
	c := flag.Bool("cover", false, "Adds cover to makefile")
	ch := flag.Bool("coverHTML", false, "Adds cover HTML to makefile")
	cp := flag.Bool("cpuProfile", false, "Adds CPU profiling to makefile")
	mp := flag.Bool("memProfile", false, "Adds Memory profiling to makefile")
	r := flag.Bool("race", false, "Adds race checking to makefile")
	tr := flag.Bool("testRace", false, "Adds race checking tests to makefile")
	l := flag.Bool("library", false, "Creates a library makefile")
	m := flag.String("mod", "", "Creates a mod file. Specify the source control path (github.com/user/project).")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Expected use: maker DIRNAME")
		os.Exit(1)
	}
	dirName := flag.Arg(0)

	templ := template.Must(template.New("makefile").Parse(makefileTemplate))

	var buffer bytes.Buffer
	err := templ.Execute(&buffer, map[string]interface{}{
		"test":       *t,
		"bench":      *b,
		"shadow":     *s,
		"cover":      *c,
		"coverHTML":  *ch,
		"cpuProfile": *cp,
		"memProfile": *mp,
		"race":       *r,
		"testRace":   *tr,
		"library":    *l,
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
	if !(*l) {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"main.go", []byte(`package main

func main() {
}
`), 0744)
	} else {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+dirName+".go", []byte("package "+dirName+"\n"), 0744)
	}
	if err != nil {
		panic(err)
	}
	if *m != "" {
		err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"go.mod", []byte(fmt.Sprintf(`module %s

go 1.14
`, *m)), 0744)
		if err != nil {
			panic(err)
		}
	}
}
