package slaxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/slack-go/slack"
)

type handler func(l net.Listener)

// Config holds all config values
type Config struct {
	GracePeriod       time.Duration `mapstructure:"grace-period"`
	Addr              string        `mapstructure:"addr"`
	SlackToken        string        `mapstructure:"token"`
	DiscordWebhookURL string        `mapstructure:"discord-webhook-url"`
	ExcludedFields    []string      `mapstructure:"excluded-fields"`
}

// server types
type server struct {
	cfg            Config
	logger         Logger
	done           chan struct{}
	srv            *http.Server
	errChan        chan error
	slack          *slack.Client
	client         *resty.Client
	excludedFields []*regexp.Regexp
}

// Server represents a server instance
type Server interface {
	Start() error
	Stop() error
	Errors() <-chan error
}

// New creates a new server instance
func New(cfg Config, logger Logger) Server {
	return &server{
		cfg:     cfg,
		logger:  logger,
		done:    make(chan struct{}, 1),
		errChan: make(chan error, 100),
	}
}

// Start starts up the server
func (s *server) Start() error {
	return s.setup(s.cfg.Addr, s.handleWeb)
}

// Stop gracefully shuts down the server
func (s *server) Stop() error {
	s.done <- struct{}{}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.GracePeriod)
	err := s.srv.Shutdown(ctx)
	cancel()

	return err
}

// Errors returns the error channel
func (s *server) Errors() <-chan error {
	return s.errChan
}

// setup starts up a server with its own listener and handler function
func (s *server) setup(addr string, handler handler) error {
	// pre-compile regexes
	excludedFields := make([]*regexp.Regexp, 0, len(s.cfg.ExcludedFields))
	for _, regex := range s.cfg.ExcludedFields {
		excludedFields = append(excludedFields, regexp.MustCompile(regex))
	}
	s.excludedFields = excludedFields

	if s.cfg.SlackToken != "" {
		client := slack.New(s.cfg.SlackToken)
		_, err := client.AuthTest()
		if err != nil {
			return fmt.Errorf("slack auth failed, err=%w", err)
		}
		s.slack = client
	}

	if s.cfg.DiscordWebhookURL != "" {
		s.client = resty.New()
		res, err := s.client.R().Get(s.cfg.DiscordWebhookURL)
		if err != nil {
			return fmt.Errorf("failed to check webhook connection err: %w", err)
		}
		if res.StatusCode() >= 300 {
			return fmt.Errorf("failed to get webhook info: %s", res.Body())
		}
	}

	// start tcp listener
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s, err: %w", addr, err)
	}

	s.logger.Info(fmt.Sprintf("Listening on %s", addr))
	go s.handleListener(l, addr, handler)

	return nil
}

// handleListener handles a listener using the specified handler function
func (s *server) handleListener(l net.Listener, addr string, handler handler) {
	defer s.logger.Info(fmt.Sprintf("Listener %s shutdown", addr))

	handler(l)
}

// handleWeb handles all incoming connections to the webhook server
func (s *server) handleWeb(l net.Listener) {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/webhook/sentry/", s.handleWebhook)

	s.srv = &http.Server{
		Handler: mux,
	}
	err := s.srv.Serve(l)

	// server closed abnormally
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		err = fmt.Errorf("server failed, err: %w", err)
		s.errChan <- err
	}
}
