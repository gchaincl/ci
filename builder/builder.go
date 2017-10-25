package builder

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/gchaincl/ci"
	"github.com/gchaincl/ci/docker"
	"github.com/gchaincl/ci/models"
	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v1"
)

type Builder struct {
}

func (b *Builder) Build(ctx context.Context, build *models.Build) (int, error) {
	// TODO: implement build numbers
	build.Number = int(time.Now().Unix())

	target, err := b.clone(ctx, os.Stdout, build)
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(target)

	spec, err := parseSpec(target + "/ci.yml")
	if err != nil {
		return 0, err
	}

	return b.build(ctx, spec, build)
}

func (b *Builder) clone(ctx context.Context, w io.Writer, build *models.Build) (string, error) {
	target := fmt.Sprintf("builds/%s/%s/%d/%s/%s",
		build.Owner, build.Repo,
		build.Number,
		build.Owner, build.Repo,
	)
	log.Printf("Cloning into %s", target)

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

func (b *Builder) build(ctx context.Context, spec *ci.Spec, build *models.Build) (int, error) {
	name := fmt.Sprintf("%s/%s#%d", build.Owner, build.Repo, build.Number)
	log.Printf("Building %s", name)
	dkr, err := docker.New(build.Commit)
	if err != nil {
		return 0, err
	}

	defer log.Printf("Build %s done", name)
	return (&ci.Runner{Destroy: true}).Run(dkr, spec)
}
