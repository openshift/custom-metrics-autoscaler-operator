package main

import (
	"bytes"
	"os"

	yamlv2 "gopkg.in/yaml.v2"
	"sigs.k8s.io/yaml"
)

func main() {
	var err error
	r := os.Stdin
	if len(os.Args) > 1 {
		r, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer r.Close()
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(r)
	if err != nil {
		panic(err)
	}
	yamlv2.FutureLineWrap()
	yamlPayload, err := yaml.JSONToYAML(buf.Bytes())
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(yamlPayload)
}
