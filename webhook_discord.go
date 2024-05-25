package slaxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (s *server) discordHandleHook(hook *webhook) error {
	if s.cfg.DiscordWebhookURL == "" {
		return nil
	}

	message := s.createDiscordMessage(hook)
	res, err := s.client.R().SetBody(message).Post(s.cfg.DiscordWebhookURL)
	if err != nil {
		message_json, _ := json.Marshal(message)
		return fmt.Errorf("failed to send discord message, err=%w, message=%v", err, string(message_json))
	}
	if res.StatusCode() >= 300 {
		message_json, _ := json.Marshal(message)
		return fmt.Errorf("failed to send discord message, response_body=%s, message=%v", res.Body(), string(message_json))
	}

	return nil
}

// createMessage will create the client message attachment
func (s *server) createDiscordMessage(hook *webhook) discordgo.MessageSend {
	buf := bytes.NewBuffer(nil)
	// default fields
	fmt.Fprintf(buf, "### Culprit\n`%s`\n", hook.Culprit)
	fmt.Fprintf(buf, "### Project\n`%s`\n", hook.ProjectName)
	fmt.Fprintf(buf, "### Level\n`%s`\n", hook.Level)

	if hook.Event.Location != "" {
		fmt.Fprintf(buf, "### Location\n`%s`\n", hook.Event.Location)
	}

	if hook.Event.Timestamp != 0 {
		fmt.Fprintf(buf, "### Timestamp `%s`\n", time.Unix(int64(int(hook.Event.Timestamp)), 0).Format(time.RFC3339))
	}

	if hook.Event.Environment != "" {
		fmt.Fprintf(buf, "### Environment: `%s`\n", hook.Event.Environment)
	}

	if hook.Event.Release != "" {
		fmt.Fprintf(buf, "### Release\n`%s`\n", hook.Event.Release)
	}

	if len(hook.Event.Exception.Values) > 0 && len(hook.Event.Exception.Values[0].Stacktrace.Frames) > 0 {
		frameLen := len(hook.Event.Exception.Values[0].Stacktrace.Frames)
		fmt.Fprintf(buf, "### Stacktrace\n%s\n", hook.Event.Exception.Values[0].Stacktrace.Frames[frameLen-1].String())
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
		fmt.Fprintf(buf, "**%s**: `%s`\n", title, tagValue)
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

	return discordgo.MessageSend{
		Content: title + "\n" + buf.String(),
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Title:       "Sentry",
				Description: hook.URL,
			},
		},
	}
}
