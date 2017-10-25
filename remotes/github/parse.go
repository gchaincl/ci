package github

import (
	"io/ioutil"
	"net/http"

	"github.com/gchaincl/ci/models"
	"github.com/google/go-github/github"
)

func parseWebHook(r *http.Request) (*models.Build, error) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	event, err := github.ParseWebHook(github.WebHookType(r), buf)
	if err != nil {
		return nil, err
	}

	switch event := event.(type) {
	case *github.PushEvent:
		return parsePushEvent(event)
	}

	return nil, nil
}

func parsePushEvent(event *github.PushEvent) (*models.Build, error) {
	build := &models.Build{
		CloneURL: event.Repo.GetCloneURL(),
		Commit:   event.HeadCommit.GetID(),
		Owner:    event.Repo.Owner.GetName(),
		Repo:     event.Repo.GetName(),
	}
	return build, nil
}
