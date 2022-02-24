package slaxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
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
	return fmt.Sprintf("filename: `%v`\tline: `%v`\nabs_path: `%v`\ncontext_line:\n```\n%v\n```\n",
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

	// create message attachment
	attachment := s.createAttachment(&hook)

	// post the message
	s.logger.Debugf("begin post message to slack, channel=%v attachment=%v", channel, attachment)
	channelID, timestamp, err := s.slack.PostMessage(channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		w.WriteHeader(500)
		s.logger.Errorf("Error while posting message: %s", err.Error())

		return
	}
	s.logger.Infof("Message successfully sent to channel %s (%s) at %s", channelID, channel, timestamp)

	w.WriteHeader(200)
}

// createAttachment will create the slack message attachment
func (s *server) createAttachment(hook *webhook) slack.Attachment {
	// default fields
	fields := []slack.AttachmentField{
		{
			Title: "Culprit",
			Value: hook.Culprit,
		},
		{
			Title: "Project",
			Value: hook.ProjectName,
			Short: true,
		},
		{
			Title: "Level",
			Value: hook.Level,
			Short: true,
		},
	}

	if hook.Event.Location != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Location",
			Value: hook.Event.Location,
			Short: true,
		})
	}

	if hook.Event.Timestamp != 0 {
		fields = append(fields, slack.AttachmentField{
			Title: "Timestamp",
			Value: time.Unix(int64(int(hook.Event.Timestamp)), 0).Format(time.RFC3339),
			Short: true,
		})
	}

	if hook.Event.Environment != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Environment",
			Value: hook.Event.Environment,
			Short: true,
		})
	}

	if hook.Event.Release != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Release",
			Value: hook.Event.Release,
			Short: true,
		})
	}

	if len(hook.Event.Exception.Values) > 0 && len(hook.Event.Exception.Values[0].Stacktrace.Frames) > 0 {
		fields = append(fields, slack.AttachmentField{
			Title: "Stacktrace",
			Value: hook.Event.Exception.Values[0].Stacktrace.Frames[0].String(),
		})
	}

	// put all sentry tags as attachment fields
	for _, tag := range hook.Event.Tags {
		// skip the default fields we already set
		if tag[0] == "culprit" || tag[0] == "project" || tag[0] == "level" ||
			tag[0] == "location" || tag[0] == "release" || tag[0] == "sentry:release" {
			continue
		}

		// skip everything that is user-excluded
		if s.isExcluded(tag[0]) {
			continue
		}

		title := strings.Title(strings.ReplaceAll(tag[0], "_", " "))
		fields = append(fields, slack.AttachmentField{
			Title: title,
			Value: tag[1],
			Short: true,
		})
	}

	var title string
	// message is empty most of the time
	if hook.Message != "" {
		lines := strings.Split(hook.Message, "\n")
		title = lines[0]
	}

	if title == "" {
		// fallback to event.title
		title = fmt.Sprintf("[%s] %s", hook.Event.Location, hook.Event.Title)
	}

	return slack.Attachment{
		Text:   fmt.Sprintf("<%s|*%s*>", hook.URL, title),
		Color:  "#f43f20",
		Fields: fields,
	}
}

// isExcluded checks whether str should be excluded
func (s *server) isExcluded(str string) bool {
	for _, regex := range s.excludedFields {
		if regex.MatchString(str) {
			return true
		}
	}

	return false
}
