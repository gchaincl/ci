package github

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gchaincl/ci/remotes/github/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseWebHook(t *testing.T) {
	body := bytes.NewBufferString(fixtures.HookPush)
	r, _ := http.NewRequest("POST", "/", body)
	r.Header.Set("X-Github-Event", "push")
	build, err := parseWebHook(r)
	require.NoError(t, err)
	require.NotNil(t, build)

	assert.Equal(t, "https://github.com/gchaincl/ci.git", build.CloneURL)
	assert.Equal(t, "7262ad78a85f78cde63185fd9869b10248587e16", build.Commit)
	assert.Equal(t, "gchaincl", build.Owner)
	assert.Equal(t, "ci", build.Repo)
}
