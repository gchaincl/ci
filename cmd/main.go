package main

import (
	"flag"
	"log"
	"os"

	"github.com/gchaincl/ci"
	"github.com/gchaincl/ci/docker"
	"gopkg.in/yaml.v2"
)

func parseSpec(f *os.File) (*ci.Spec, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, info.Size())
	if _, err := f.Read(buf); err != nil {
		return nil, err
	}

	var spec *ci.Spec
	if err := yaml.Unmarshal(buf, &spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func main() {
	var (
		file    string
		sess    string
		destroy bool
	)

	flag.StringVar(&file, "c", ".ci.yml", "CI Spec file")
	flag.StringVar(&sess, "s", "", "Session")
	flag.BoolVar(&destroy, "d", false, "Destroy Services on exit")
	flag.Parse()

	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	spec, err := parseSpec(f)
	if err != nil {
		log.Fatalln(err)
	}

	dkr, err := docker.New(sess)
	if err != nil {
		log.Fatalln(err)
	}

	status, err := (&ci.Runner{Destroy: destroy}).Run(dkr, spec)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(status)
}
