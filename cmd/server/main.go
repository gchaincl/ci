package main

import (
	"net/http"

	"github.com/gchaincl/ci/remotes/github"
)

func main() {
	gh := github.New()
	http.HandleFunc("/gh/hooks", gh.Hooks)
	http.ListenAndServe(":10080", nil)
}
