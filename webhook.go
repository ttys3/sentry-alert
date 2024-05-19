package slaxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type webhook struct {
	ProjectName     string   `json:"project_name"` // "nexus-tracker-test"
	Message         string   `json:"message"`      // most time this is empty
	ID              string   `json:"id"`
	Culprit         string   `json:"culprit"`
	ProjectSlug     string   `json:"project_slug"`
	URL             string   `json:"url"`
	Level           string   `json:"level"`            // "error"
	TriggeringRules []string `json:"triggering_rules"` // eg: [ "Send a notification for new issues" ]

	Event sentryEvent
}

type sentryEvent struct {
	Culprit     string `json:"culprit"`     // the same as parent culprit
	Title       string `json:"title"`       // "*fmt.wrapError: this is an test error, err=file does not exist"
	EventID     string `json:"event_id"`    // "fec9f96296cb47d89e652d183e2752cf"
	Environment string `json:"environment"` // also event.tags ["environment", "develop"]
	Platform    string `json:"platform"`    //  "go", "rust"
	Version     string `json:"version"`
	Location    string `json:"location"` // "/home/ttys3/repo/go/sentry-go-test/main.go"
	Logger      string `json:"logger"`
	Type        string `json:"type"` // "error"

	Metadata sentryEvtMetadata `json:"metadata"`
	Tags     []sentryTag

	Timestamp float64 `json:"timestamp"` // 1645672116.893372
	Received  float64 `json:"received"`  // 1645672117.030224

	Level string `json:"level"` //  also event.tags ["level", "error"]

	Project int    `json:"project"`
	Release string `json:"release"` // also event.tags ["sentry:release", "v1.1.0"]

	User      sentryUser `json:"user,omitempty"`
	Sdk       Sdk        `json:"sdk"`
	Exception Exception  `json:"exception"`
}

type sentryEvtMetadata struct {
	Function string `json:"function"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Filename string `json:"filename"`
}

// sentryTag is an array as two elements, in [key, value] format
type sentryTag [2]string

type sentryUser struct {
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
	Geo       struct {
		Region      string `json:"region"`
		CountryCode string `json:"country_code"`
	} `json:"geo"`
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Sdk struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

type Exception struct {
	Values []ExceptionValue `json:"values"`
}

type ExceptionValue struct {
	Stacktrace Stacktrace `json:"stacktrace"`
	Type       string     `json:"type"`
	Value      string     `json:"value"`
	Mechanism  struct {
		Type    string `json:"type"`
		Handled bool   `json:"handled"`
	} `json:"mechanism"`
}

type Stacktrace struct {
	Frames []StacktraceFrame `json:"frames"`
}

type StacktraceFrame struct {
	AbsPath     string        `json:"abs_path"`
	PreContext  []interface{} `json:"pre_context"`
	PostContext []interface{} `json:"post_context"`
	InApp       bool          `json:"in_app"`
	Lineno      int           `json:"lineno"`
	Filename    string        `json:"filename"`
	ContextLine string        `json:"context_line"`
}

func (s *StacktraceFrame) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("filename: `%v`\nline: `%v`\nabs_path: `%v`\ncontext_line:\n```\n%v\n```\n",
		s.Filename, s.Lineno, s.AbsPath, s.ContextLine)
}

type Request struct {
	URL                 string                 `json:"url"`
	Headers             [][]string             `json:"headers"` // "Referer", "Origin"
	Data                map[string]interface{} `json:"data"`
	Method              string                 `json:"method"`
	InferredContentType string                 `json:"inferred_content_type"`
}

// handleWebhook handles one webhook request
func (s *server) handleWebhook(w http.ResponseWriter, req *http.Request) {
	// validations
	if req.Method != http.MethodPost {
		w.WriteHeader(405)

		return
	}

	// the last part is slack channel id
	// /webhook/sentry/:SlackChannelID
	parts := strings.Split(req.RequestURI, "/")
	channel := parts[len(parts)-1]
	if channel == "" {
		w.WriteHeader(400)
		w.Write([]byte("empty slack channel ID"))
		return
	}

	// read body
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		s.logger.Errorf("Could not read response body: %s", err.Error())

		return
	}
	defer req.Body.Close()

	s.logger.Debugf("read request payload success, body=%s", string(buf))

	// parse webhook
	var hook webhook

	err = json.Unmarshal(buf, &hook)
	if err != nil {
		w.WriteHeader(500)
		s.logger.Errorf("Could not parse webhook payload: %s", err.Error())

		return
	}
	s.logger.Debugf("parse webhook payload success, payload=%+v", hook)

	err = errors.Join(s.slackHandleHook(&hook, channel), s.discordHandleHook(&hook))
	if err != nil {
		w.WriteHeader(500)
		s.logger.Errorf("Error while posting message: %s", err.Error())
		return
	}

	w.WriteHeader(200)
}
