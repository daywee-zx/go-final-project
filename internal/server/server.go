package server

import (
	"crypto/sha256"
	"encoding/hex"
	"go_final/internal/scheduler_db"
	"log"
	"net/http"
	"time"
)

type Server struct {
	Server   *http.Server
	db       *scheduler_db.SchedulerDB
	password string
}

func New(port string, logs *log.Logger, db *scheduler_db.SchedulerDB, password string) *Server {
	hashString := ""
	if len(password) > 0 {
		passwordHash := sha256.Sum256([]byte(password))
		hashString = hex.EncodeToString(passwordHash[:])
	}

	server := &http.Server{
		Addr:         ":" + port,
		ErrorLog:     logs,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Server{
		Server:   server,
		db:       db,
		password: hashString,
	}
}

func (s *Server) Init(webDir string) error {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	mux.HandleFunc("GET /api/nextdate", s.NextDayHandler)

	mux.HandleFunc("POST /api/task", s.checkAuth(s.PostTaskHandler))
	mux.HandleFunc("GET /api/task", s.checkAuth(s.GetTaskHandler))
	mux.HandleFunc("PUT /api/task", s.checkAuth(s.UpdateTaskHandler))
	mux.HandleFunc("DELETE /api/task", s.checkAuth(s.DeleteTaskHandler))

	mux.HandleFunc("POST /api/task/done", s.checkAuth(s.MarkDoneHandler))

	mux.HandleFunc("GET /api/tasks", s.checkAuth(s.GetTasksHandler))

	mux.HandleFunc("POST /api/signin", s.LoginHandler)

	s.Server.Handler = mux

	return s.Server.ListenAndServe()
}

func (s *Server) checkAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(s.password) == 0 {
			next(w, r)
			return
		}

		var jwt string
		cookie, err := r.Cookie("token")
		if err == nil {
			jwt = cookie.Value
		}

		if jwt != generateToken(s.password) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
