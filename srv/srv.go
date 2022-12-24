package srv

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tirathawat/assessment/config"
	"github.com/tirathawat/assessment/logs"
	"github.com/tirathawat/assessment/router"
)

type Server interface {
	Run()
	Shutdown() error
	Port() string
}

type server struct {
	*http.Server
	port string
}

func NewServer(cfg *config.AppConfig, handlers *router.Handlers) Server {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Default())

	router.Register(r, handlers)

	s := &http.Server{
		Addr:    cfg.Port,
		Handler: r,
	}

	return &server{
		Server: s,
		port:   cfg.Port,
	}
}

func (s *server) Run() {
	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Error().Err(err).Msg("Cannot initialize application")
			panic(err)
		}
	}()
}

func (s *server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

func (s *server) Port() string {
	return s.port
}
