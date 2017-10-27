package github

import (
	"errors"
	"net/http"
	"os"

	"github.com/gchaincl/ci/models"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type Client struct {
	*github.Client
}

func New() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("Please specify a GITHUB_TOKEN env var")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		Client: github.NewClient(tc),
	}, nil
}

func (c *Client) Status(ctx context.Context, b *models.Build) error {
	status := &github.RepoStatus{
		State:       github.String(b.Status),
		TargetURL:   github.String(b.Link),
		Context:     github.String("ci/pr"),
		Description: github.String("CI building..."),
	}

	_, _, err := c.Repositories.CreateStatus(ctx, b.Owner, b.Repo, b.Commit, status)
	return err
}

func (c *Client) Hooks(w http.ResponseWriter, r *http.Request) (*models.Build, error) {
	build, err := parseWebHook(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil, err
	}

	return build, nil
}
