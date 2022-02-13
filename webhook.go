package slaxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
)

type event struct {
	Tags [][]string
}

type webhook struct {
	ProjectName string `json:"project_name"`
	Message     string
	Culprit     string
	URL         string
	Level       string
	Event       event
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
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		s.logger.Errorf("Could not read response body: %s", err.Error())

		return
	}
	defer req.Body.Close()

	// parse webhook
	var hook webhook

	err = json.Unmarshal(buf, &hook)
	if err != nil {
		w.WriteHeader(500)
		s.logger.Errorf("Could not parse webhook payload: %s", err.Error())

		return
	}

	// create message attachment
	attachment := s.createAttachment(hook)

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
func (s *server) createAttachment(hook webhook) slack.Attachment {
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

	// put all sentry tags as attachment fields
	for _, tag := range hook.Event.Tags {
		// skip the default fields
		if tag[0] == "culprit" || tag[0] == "project" || tag[0] == "level" {
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

	lines := strings.Split(hook.Message, "\n")

	return slack.Attachment{
		Text:   fmt.Sprintf("<%s|*%s*>", hook.URL, lines[0]),
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
