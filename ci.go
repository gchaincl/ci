package ci

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/gchaincl/ci/docker"
)

type Spec struct {
	Image    string
	Services map[string]Service
	Script   []string
}

func (s *Spec) Bash() string {
	var buf bytes.Buffer
	buf.WriteString("#!/bin/bash\n")
	buf.WriteString("set -e\n")

	for _, line := range s.Script {
		buf.WriteString("echo '=> " + line + "'\n")
		buf.WriteString(line + "\n")
	}
	return buf.String()
}

type Service struct {
	Image       string
	Environment map[string]string
}

func (s *Service) EnvSlice() []string {
	var env []string
	for key, val := range s.Environment {
		env = append(env, key+"="+val)
	}
	return env
}

type Runner struct {
	Outdoor bool
	Destroy bool
	Stdout  io.Writer
	Stderr  io.Writer
}

func (r *Runner) Run(dkr *docker.Docker, spec *Spec) (int, error) {
	if r.Stdout == nil {
		r.Stdout = os.Stdout
	}

	if r.Stderr == nil {
		r.Stderr = os.Stdout
	}

	for name, srv := range spec.Services {
		fmt.Fprintf(r.Stdout, "Starting service %s (%s)\n", name, srv.Image)

		opts := docker.UpOptions{
			Env: srv.EnvSlice(),
		}
		if _, err := dkr.Up(name, srv.Image, opts); err != nil {
			return 0, err
		}
	}

	if r.Destroy {
		defer destroyServices(dkr, spec)
	}

	// Create script
	f, err := createTempFile(spec.Bash())
	if err != nil {
		return 0, err
	}
	defer os.Remove(f.Name())

	// Build env
	var env []string
	links, err := getLinks(dkr, spec)
	if err != nil {
		return 0, err
	}
	for name, link := range links {
		env = append(env, fmt.Sprintf("CI_%s=%s", name, link))
	}

	if r.Outdoor {
		log.Println("Running in outdoor")
		return r.runOutdoor("./"+f.Name(), env)
	}

	fmt.Fprintf(r.Stdout, "Running in container %s\n", spec.Image)
	return r.runInContainer(dkr, spec.Image, "./"+f.Name(), env)
}

func getLinks(dkr *docker.Docker, spec *Spec) (map[string]string, error) {
	links := make(map[string]string)
	for name, _ := range spec.Services {
		ip, err := dkr.IP(name)
		if err != nil {
			return nil, err
		}
		links[name] = ip
	}
	return links, nil
}

func (r *Runner) runOutdoor(script string, env []string) (int, error) {
	// Use the current process env too
	for _, e := range os.Environ() {
		env = append(env, e)
	}

	cmd := exec.Command(script)
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return 0, err
		}

		fmt.Fprintf(r.Stdout, "%s", out)
		if exitErr.Success() {
			return 0, nil
		}
		return 1, nil
	}
	fmt.Fprintf(r.Stdout, "%s", out)
	return 0, nil
}

func (r *Runner) runInContainer(dkr *docker.Docker, image, script string, env []string) (int, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return 0, err
	}

	_, err = dkr.Up("runner", image, docker.UpOptions{
		Cmd:        []string{script},
		Env:        env,
		AutoRemove: true,
		Log:        true,
		Binds: []string{
			pwd + ":" + "/ci/src",
		},
		WorkingDir: "/ci/src",
		Stdout:     r.Stdout,
		Stderr:     r.Stderr,
	})
	return dkr.Wait("runner")
}

func destroyServices(dkr *docker.Docker, spec *Spec) error {
	for name, _ := range spec.Services {
		if err := dkr.Remove(name); err != nil {
			return err
		}
	}
	return nil
}

func createTempFile(content string) (*os.File, error) {
	f, err := ioutil.TempFile(".", "ci_")
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(f.Name(), 0755); err != nil {
		return nil, err
	}

	if _, err := f.WriteString(content); err != nil {
		return nil, err
	}

	return f, f.Close()
}
