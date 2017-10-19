package docker

import (
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

type Container docker.Container

type Docker struct {
	*docker.Client
	session  string
	services map[string]*Container
}

type UpOptions struct {
	// Conf
	Cmd        []string
	Env        []string
	WorkingDir string

	// HostConf
	Binds      []string
	AutoRemove bool

	// Behavior
	Log bool
}

func New(s string) (*Docker, error) {
	client, err := docker.NewClient("unix://var/run/docker.sock")
	if err != nil {
		return nil, err
	}

	return &Docker{
		Client:   client,
		session:  s,
		services: make(map[string]*Container),
	}, nil
}

func (d *Docker) name(name string) string {
	return d.session + "_" + name
}

func (d *Docker) Up(name, image string, opts UpOptions) (string, error) {
	cName := d.name(name)

	c, err := d.findOrCreate(cName, image, &opts)
	if err != nil {
		return "", err
	}

	if !c.State.Running {
		if err := d.StartContainer(c.ID, nil); err != nil {
			return "", err
		}
	}

	if opts.Log == true {
		lOpts := docker.LogsOptions{
			Container:    c.ID,
			Follow:       true,
			OutputStream: os.Stdout,
			ErrorStream:  os.Stdout,
			Stderr:       true,
			Stdout:       true,
		}
		if err := d.Logs(lOpts); err != nil {
			return "", err
		}
	}

	return cName, nil
}

func (d *Docker) findOrCreate(name, image string, opts *UpOptions) (*docker.Container, error) {
	c, err := d.InspectContainer(name)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); !ok {
			return nil, err
		}
	}

	// container has been found
	if c != nil {
		return c, nil
	}

	cOpts := docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			Image:      image,
			Cmd:        opts.Cmd,
			Env:        opts.Env,
			WorkingDir: opts.WorkingDir,
		},
		HostConfig: &docker.HostConfig{
			AutoRemove: opts.AutoRemove,
			Binds:      opts.Binds,
		},
	}

	return d.CreateContainer(cOpts)
}

func (d *Docker) Remove(name string) error {
	return d.RemoveContainer(docker.RemoveContainerOptions{
		ID:    d.name(name),
		Force: true,
	})
}

func (d *Docker) Wait(name string) (int, error) {
	return d.Client.WaitContainer(d.name(name))
}

func (d *Docker) IP(name string) (string, error) {
	c, err := d.Client.InspectContainer(d.name(name))
	if err != nil {
		return "", err
	}

	return c.NetworkSettings.IPAddress, nil
}
