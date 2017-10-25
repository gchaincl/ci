package main

import (
	"log"
	"net/http"

	"github.com/gchaincl/ci/remotes/github"
)

func main() {
	gh, err := github.New()
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/gh/hooks", gh.Hooks)
	http.ListenAndServe(":10080", nil)
}
