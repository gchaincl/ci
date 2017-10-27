package remotes

import (
	"net/http"

	"github.com/gchaincl/ci/models"
	"golang.org/x/net/context"
)

type Remote interface {
	Status(context.Context, *models.Build) error
	Hooks(w http.ResponseWriter, r *http.Request) (*models.Build, error)
}
