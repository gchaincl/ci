package builder

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	"github.com/gchaincl/ci"
	"github.com/gchaincl/ci/docker"
	"github.com/gchaincl/ci/models"
	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v1"
)

type Builder interface {
	Build(context.Context, *models.Build, io.Writer) (int, error)
}

type DockerBuilder struct {
}

// Build triggers a build defined by *models.Build, it will log the output to w.
// if w is nil os.Stdout will be used
func (b *DockerBuilder) Build(ctx context.Context, build *models.Build, w io.Writer) (int, error) {
	target, err := b.clone(ctx, build, w)
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(target)

	spec, err := parseSpec(target + "/ci.yml")
	if err != nil {
		return 0, err
	}

	return b.build(ctx, spec, build, w)
}

func (b *DockerBuilder) clone(ctx context.Context, build *models.Build, w io.Writer) (string, error) {
	target := fmt.Sprintf("builds/%s/%s/%d/%s/%s",
		build.Owner, build.Repo,
		build.Number,
		build.Owner, build.Repo,
	)
	fmt.Fprintf(w, "Cloning into %s\n", target)

	_, err := git.PlainClone(target, false, &git.CloneOptions{
		URL:      build.CloneURL,
		Progress: w,
		Depth:    1,
	})

	return target, err
}

func parseSpec(file string) (*ci.Spec, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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

func (b *DockerBuilder) build(ctx context.Context, spec *ci.Spec, build *models.Build, w io.Writer) (int, error) {
	name := fmt.Sprintf("%s/%s#%d", build.Owner, build.Repo, build.Number)
	fmt.Fprintf(w, "Building %s\n", name)
	dkr, err := docker.New(build.Commit)
	if err != nil {
		return 0, err
	}

	s, err := (&ci.Runner{Destroy: true, Stdout: w, Stderr: w}).Run(dkr, spec)
	fmt.Fprintf(w, "Build %s done\n", name)
	return s, err
}
