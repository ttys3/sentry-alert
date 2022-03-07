package slaxy

import (
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"os"
	"testing"
	"time"
)

func TestSlackPostMessage(t *testing.T) {
	hook := webhook{
		ProjectName:     "demo-project",
		Message:         "",
		ID:              "007",
		Culprit:         "createAttachment()",
		ProjectSlug:     "demo-project",
		URL:             "https://www.google.com/",
		Level:           "error",
		TriggeringRules: nil,
		Event: sentryEvent{
			Culprit:     "createAttachment()",
			Title:       "<this is 'title'>",
			EventID:     "",
			Environment: "develop",
			Platform:    "go",
			Version:     "0.12.0",
			Location:    "webhook.go",
			Logger:      "",
			Type:        "",
			Metadata:    sentryEvtMetadata{},
			Tags:        nil,
			Timestamp:   float64(time.Now().Unix()),
			Received:    float64(time.Now().Add(time.Second * 5).Unix()),
			Level:       "error",
			Project:     0,
			Release:     "0.2.9",
			User:        sentryUser{},
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
		fields = append(fields, slack.AttachmentField{
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
		GracePeriod:    0,
		Addr:           "",
		Token:          os.Getenv("SLAXY_TOKEN"),
		ExcludedFields: nil,
	}
	s := &server{
		cfg:     cfg,
		logger:  logrus.New(),
		done:    make(chan struct{}, 1),
		errChan: make(chan error, 100),
	}
	s.setup(":8080", nil)
	attachment := s.createAttachment(&hook)
	channel := os.Getenv("SLACK_CHANNEL")
	channelID, timestamp, err := s.slack.PostMessage(channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("success, channelID=%v timestamp=%v", channelID, timestamp)
}
