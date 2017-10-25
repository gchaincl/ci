package github

import (
	"log"
	"net/http"

	"github.com/gchaincl/ci/builder"
	"github.com/gchaincl/ci/models"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type Client struct {
	*github.Client
	builder *builder.Builder
}

func New() *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "066239b071d2111bfcfa1f988247112d384fe9a0"},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		Client:  github.NewClient(tc),
		builder: &builder.Builder{},
	}
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

func (c *Client) Hooks(w http.ResponseWriter, r *http.Request) {
	log.Println("> GH Webhook")
	build, err := parseWebHook(r)
	if err != nil {
		log.Printf("err = %+v\n", err)
		return
	}

	build.Status = "pending"
	if err := c.Status(r.Context(), build); err != nil {
		log.Printf("err = %+v\n", err)
		return
	}

	status, err := c.builder.Build(r.Context(), build)
	if err != nil {
		log.Printf("err = %+v\n", err)
		return
	}

	if status != 0 {
		build.Status = "error"
		if err := c.Status(r.Context(), build); err != nil {
			log.Printf("err = %+v\n", err)
			return
		}
	}

	build.Status = "success"
	if err := c.Status(r.Context(), build); err != nil {
		log.Printf("err = %+v\n", err)
		return
	}
}
