package http

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gutmensch/goweck/internal/api"
	"github.com/gutmensch/goweck/internal/common"
	"github.com/gutmensch/goweck/internal/middleware"
	"github.com/gutmensch/goweck/internal/webapp"
	"log"
	"net/http"
	"time"
)

type Server struct {
	*http.Server
}

func NewServer(listen string) *Server {
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(middleware.LogAccessRequest)
	apiRouter.HandleFunc("/alarm/all", api.AlarmListHandler).Methods("GET")
	apiRouter.HandleFunc("/zone/all", api.ZoneListHandler).Methods("GET")
	apiRouter.HandleFunc("/stream/all", api.StreamListHandler).Methods("GET")
	apiRouter.HandleFunc("/alarm/create", api.AlarmCreateHandler).Methods("POST")
	apiRouter.HandleFunc("/alarm/stop", api.AlarmStopHandler).Methods("POST")
	apiRouter.HandleFunc("/alarm/running", api.AlarmRunningHandler).Methods("GET")
	apiRouter.HandleFunc("/alarm/update/{id:[0-9a-f]{24}}", api.AlarmUpdateHandler).Methods("POST")
	apiRouter.HandleFunc("/alarm/delete/{id:[0-9a-f]{24}}", api.AlarmDeleteHandler).Methods("DELETE")

	staticRouter := router.PathPrefix("/static").Subrouter()
	staticRouter.Handle("/", http.StripPrefix("/static/", webapp.FS)).Methods("GET")

	internalRouter := router.PathPrefix("/internal").Subrouter()
	internalRouter.HandleFunc("/ping", webapp.PingHandler).Methods("GET")
	internalRouter.HandleFunc("/health", webapp.HealthHandler).Methods("GET")
	internalRouter.HandleFunc("/versions", webapp.VersionHandler).Methods("GET")

	webRouter := router.PathPrefix("/").Subrouter()
	webRouter.HandleFunc("/", webapp.TemplateHandler).Methods("GET")
	webRouter.HandleFunc("/index.html", webapp.TemplateHandler).Methods("GET")

	s := &http.Server{
		Handler:      router,
		Addr:         listen,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	srv := Server{
		s,
	}
	return &srv
}

func (s Server) Run(ctx context.Context) error {
	// trigger shutdown when context done
	go func() {
		<-ctx.Done()
		common.Log.Info("terminating http server")
		_ = s.Server.Shutdown(context.Background())
	}()

	// ErrServerClosed is always returned on graceful shutdown
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
