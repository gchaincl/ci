package docker

import (
	"testing"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertContainerIsRunning(t *testing.T, d *Docker, id string) {
	c, err := d.Client.InspectContainer(id)
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.True(t, c.State.Running, "Container should be running")
}

func TestUp(t *testing.T) {
	d, err := New("test")
	require.NoError(t, err)

	opts := UpOptions{
		Cmd: []string{"nc", "-l", "10000"},
	}
	name, err := d.Up("box", "alpine", opts)
	require.NoError(t, err)
	assertContainerIsRunning(t, d, name)

	t.Run("WhenContainerIsStopped", func(t *testing.T) {
		require.NoError(t, d.StopContainer(name, 10))

		name, err := d.Up("box", "alpine", opts)
		require.NoError(t, err)
		assertContainerIsRunning(t, d, name)
	})

	t.Run("WhenContainerDoesntExist", func(t *testing.T) {
		err := d.RemoveContainer(docker.RemoveContainerOptions{
			ID:    "test_box",
			Force: true,
		})
		require.NoError(t, err)

		name, err := d.Up("box", "alpine", opts)
		require.NoError(t, err)
		assertContainerIsRunning(t, d, name)
	})
}
