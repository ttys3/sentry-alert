package slaxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"

	"github.com/innogames/slaxy/version"
)

func (s *server) slackHandleHook(hook *webhook, channel string) error {
	if s.slack == nil {
		return nil
	}

	// create message attachment
	attachment := s.createAttachment(hook)

	// post the message
	s.logger.Debugf("begin post message to slack, channel=%v attachment=%v", channel, attachment)
	channelID, timestamp, err := s.slack.PostMessage(channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("error while posting message: %w", err)
	}

	s.logger.Infof("Message successfully sent to channel %s (%s) at %s", channelID, channel, timestamp)
	return nil
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
		frameLen := len(hook.Event.Exception.Values[0].Stacktrace.Frames)
		fields = append(fields, slack.AttachmentField{
			Title: "Stacktrace",
			Value: hook.Event.Exception.Values[0].Stacktrace.Frames[frameLen-1].String(),
		})
	}

	// put all sentry tags as attachment fields
	for _, tag := range hook.Event.Tags {
		tagKey := tag[0]
		tagValue := tag[1]
		// skip the default fields we already set
		if tagKey == "culprit" || tagKey == "project" || tagKey == "level" ||
			tagKey == "location" || tagKey == "environment" || tagKey == "release" || tagKey == "sentry:release" {
			continue
		}

		// skip everything that is user-excluded
		if s.isExcluded(tagKey) {
			continue
		}

		title := strings.Title(strings.ReplaceAll(tagKey, "_", " "))
		fields = append(fields, slack.AttachmentField{
			Title: title,
			Value: tagValue,
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
		Title:     title,
		TitleLink: hook.URL,
		// Text:   fmt.Sprintf("<%s|*%s*>", html.EscapeString(hook.URL), html.EscapeString(title)),
		Color:  "#f43f20",
		Fields: fields,
		Footer: "Slaxy v" + version.Version,
		// icon from https://github.com/getsentry
		FooterIcon: "https://avatars.githubusercontent.com/u/1396951?s=200&v=4",
		Ts:         json.Number(fmt.Sprint(time.Now().Unix())),
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
