package slaxy

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestDiscordPostMessage(t *testing.T) {
	hook := webhook{
		ProjectName:     "demo-project",
		Message:         "",
		ID:              "007",
		Culprit:         "createMessage()",
		ProjectSlug:     "demo-project",
		URL:             "https://www.google.com/",
		Level:           "error",
		TriggeringRules: nil,
		Event: sentryEvent{
			Culprit:     "createMessage()",
			Title:       "<this is 'title'>",
			EventID:     "",
			Environment: "develop",
			Platform:    "go",
			Version:     "0.12.0",
			Location:    "webhook.go",
			Logger:      "",
			Type:        "",
			Metadata:    sentryEvtMetadata{},
			Tags: []sentryTag{
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
				{"key", "value"},
			},
			Timestamp: float64(time.Now().Unix()),
			Received:  float64(time.Now().Add(time.Second * 5).Unix()),
			Level:     "error",
			Project:   0,
			Release:   "0.2.9",
			User:      sentryUser{},
			Sdk: Sdk{
				Version: "0.12.0",
				Name:    "sentry-go",
			},
			Exception: Exception{
				Values: []ExceptionValue{
					ExceptionValue{
						Stacktrace: Stacktrace{
							Frames: []StacktraceFrame{
								{
									AbsPath:     "/slaxy/webhook.go",
									PreContext:  nil,
									PostContext: nil,
									InApp:       false,
									Lineno:      168,
									Filename:    "webhook.go",
									ContextLine: `	if hook.Event.Timestamp != 0 {
		fields = append(fields, client.AttachmentField{
			Title: "Timestamp",
			Value: time.Unix(int64(int(hook.Event.Timestamp)), 0).Format(time.RFC3339),
			Short: true,
		})
	}`,
								},
							},
						},
						Type:  "",
						Value: "",
						Mechanism: struct {
							Type    string `json:"type"`
							Handled bool   `json:"handled"`
						}{},
					},
				},
			},
		},
	}

	cfg := Config{
		GracePeriod:       0,
		Addr:              "",
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		ExcludedFields:    nil,
	}
	s := &server{
		cfg:     cfg,
		logger:  logrus.New(),
		done:    make(chan struct{}, 1),
		errChan: make(chan error, 100),
	}
	s.setup(":8080", func(l net.Listener) {

	})
	attachment := s.createDiscordMessage(&hook)
	res, err := s.client.R().SetBody(attachment).Post(s.cfg.DiscordWebhookURL)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode() >= 300 {
		t.Fatal(res.Body())
	}
	t.Logf("success")
}
