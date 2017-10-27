package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gchaincl/ci/builder"
	"github.com/gchaincl/ci/models"
	"github.com/gchaincl/ci/remotes"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"
)

type Server struct {
	router  *mux.Router
	builder *builder.Builder
	logs    map[string]io.Reader
}

func New() *Server {
	b := &builder.Builder{}
	router := mux.NewRouter()
	srv := &Server{
		router:  router,
		builder: b,
		logs:    make(map[string]io.Reader),
	}

	router.HandleFunc("/builds/{id}/logs", srv.logHandler)
	return srv
}

func (s *Server) logHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := mux.Vars(req)["id"]
	r, ok := s.logs[id]
	if !ok {
		http.Error(w, "ID not found", 404)
		return
	}

	rd := bufio.NewReader(r)
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			log.Printf("err = %+v\n", err)
			return
		}

		w.Write(line)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *Server) RegisterRemote(path string, remote remotes.Remote) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		build, err := remote.Hooks(w, r)
		if err != nil {
			return
		}
		go s.Build(build, remote)
	}

	s.router.HandleFunc(path, fn)
}

func (s *Server) Build(build *models.Build, remote remotes.Remote) error {
	ctx := context.Background()

	// TODO
	id := uuid.New()
	build.ID = id
	build.Link = "http://localhost:10080/builds/" + id + "/logs"
	build.Number = int(time.Now().Unix())

	r, w := io.Pipe()
	s.logs[id] = r
	defer func() {
		w.Close()
		delete(s.logs, id)
	}()

	build.Status = "pending"
	if err := remote.Status(ctx, build); err != nil {
		return err
	}

	status, err := s.builder.Build(ctx, build, w)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\n")

	if status == 0 {
		build.Status = "success"
	} else {
		build.Status = "failure"
	}

	return remote.Status(ctx, build)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
