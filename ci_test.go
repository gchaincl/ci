package ci

import (
	"testing"

	"github.com/gchaincl/ci/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerRun(t *testing.T) {
	dkr, err := docker.New("test")
	require.NoError(t, err)

	spec := &Spec{
		Image: "golang",
		Services: map[string]Service{
			"db": Service{
				Image: "mysql",
				Environment: map[string]string{
					"MYSQL_ALLOW_EMPTY_PASSWORD": "1",
				},
			},
			"cache": Service{Image: "redis"},
		},
		Script: []string{
			`echo "Hello World!"`,
			"go version",
			"not_a_command",
			"echo",
		},
	}

	t.Run("Outdoor mode", func(t *testing.T) {
		n, err := (&Runner{Outdoor: true}).Run(dkr, spec)
		require.NoError(t, err)
		assert.NotEqual(t, 0, n)
	})

	t.Run("Container mode", func(t *testing.T) {
		n, err := (&Runner{}).Run(dkr, spec)
		require.NoError(t, err)
		assert.NotEqual(t, 0, n)
	})
}

func TestSpecBash(t *testing.T) {
	s := &Spec{Script: []string{
		"go get ./...",
		"go test",
	}}

	script := `#!/bin/bash
set -e
echo '=> go get ./...'
go get ./...
echo '=> go test'
go test
`
	assert.Equal(t, script, s.Bash())
}
