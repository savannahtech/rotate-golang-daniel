package httpserver

import (
	"context"
	"log/slog"
	"net/http"
)

// Server represents an HTTP server
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// NewServer creates and returns a new Server instance
func NewServer(
	addr string,
	logger *slog.Logger,
	handler http.Handler,
) *Server {
	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting-http-server", slog.String("url", "http://"+s.httpServer.Addr))

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("error-starting-http-server", slog.String("error", err.Error()))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping-http-server")
	return s.httpServer.Shutdown(ctx)
}
