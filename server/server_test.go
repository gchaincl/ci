package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gchaincl/ci/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

type mockBuilder struct {
	build func(context.Context, *models.Build, io.Writer) (int, error)
}

func (m *mockBuilder) Build(ctx context.Context, build *models.Build, w io.Writer) (int, error) {
	if m.build == nil {
		return 0, nil
	}
	return m.build(ctx, build, w)
}

// remote mock
type mockRemote struct {
	status func(context.Context, *models.Build) error
	hooks  func(http.ResponseWriter, *http.Request) (*models.Build, error)
}

func (m *mockRemote) Status(ctx context.Context, build *models.Build) error {
	if m.status == nil {
		return nil
	}
	return m.status(ctx, build)
}

func (m *mockRemote) Hooks(w http.ResponseWriter, req *http.Request) (*models.Build, error) {
	if m.hooks == nil {
		return &models.Build{}, nil
	}
	return m.hooks(w, req)
}

func TestServerRemoteRegistration(t *testing.T) {
	r := &mockRemote{}
	b := &mockBuilder{}

	srv := New(b)
	srv.RegisterRemote("/test", r)

	testSrv := httptest.NewServer(srv)
	defer testSrv.Close()

	t.Run("Handles the endpoint", func(t *testing.T) {
		resp, err := http.Get(testSrv.URL + "/test")
		require.NoError(t, err)

		assert.Equal(t, 200, resp.StatusCode)
	})
}
