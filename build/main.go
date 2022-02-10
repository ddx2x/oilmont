package main

import (
	"fmt"
	"io/ioutil"

	flag "github.com/spf13/pflag"
	"github.com/ddx2x/oilmont/pkg/utils/obj"
)

const (
	dockerTemplate = "./images/template/Dockerfile.template"
)

var dir string
var dest string

func main() {
	flag.StringVar(&dir, "dir", "cmd", "")
	flag.StringVar(&dest, "dest", "images", "")
	flag.Parse()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	bs, err := ioutil.ReadFile(dockerTemplate)
	if err != nil {
		panic(err)
	}

	type params struct {
		App string
	}

	for _, f := range files {
		output, err := obj.RenderTemplate(params{f.Name()}, string(bs))
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(fmt.Sprintf("%s/Dockerfile.%s", dest, f.Name()), []byte(output), 0644); err != nil {
			panic(err)
		}
	}
}
