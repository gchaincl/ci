package main

import (
	"log"
	"net/http"

	"github.com/gchaincl/ci/remotes/github"
	"github.com/gchaincl/ci/server"
)

func main() {
	srv := server.New()

	gh, err := github.New()
	if err != nil {
		log.Fatalln(err)
	}

	srv.RegisterRemote("/gh/hooks", gh)

	bind := ":10080"
	log.Println("Listening on", bind)
	if err := http.ListenAndServe(bind, srv); err != nil {
		log.Fatalln(err)
	}
}
