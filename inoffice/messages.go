package inoffice

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func BuildErrorMessage(err error) slack.Message {
	msg := slack.NewBlockMessage(slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, "*An error occurred :disappointed:*", true, true), []*slack.TextBlockObject{
		slack.NewTextBlockObject(slack.PlainTextType, err.Error(), false, true),
	}, nil))
	msg.ResponseType = slack.ResponseTypeEphemeral
	return msg
}

func BuildInOfficeMessage(weekStart time.Time, o map[Day][]InOffice) slack.Message {
	var blocks = []slack.Block{
		slack.NewSectionBlock(&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "*When are you planning on coming in?*",
		}, nil, nil),
		generateWeekdayBlock(DayMonday, weekStart, o[DayMonday]),
		generateWeekdayBlock(DayTuesday, weekStart, o[DayTuesday]),
		generateWeekdayBlock(DayWednesday, weekStart, o[DayWednesday]),
		generateWeekdayBlock(DayThursday, weekStart, o[DayThursday]),
		generateWeekdayBlock(DayFriday, weekStart, o[DayFriday]),
	}

	// Truncate nil blocks, blergh
	for s, v := range blocks {
		if val := reflect.ValueOf(v); val.IsValid() && val.Interface() != nil && val.IsNil() {
			blocks = append(blocks[:s], blocks[s+1:]...)
		}
	}

	return slack.NewBlockMessage(blocks...)
}

func generateWeekdayBlock(day Day, weekStart time.Time, inOffice []InOffice) slack.Block {
	if IsInPast(weekStart, day) {
		return nil
	}

	var usersTexts []*slack.TextBlockObject
	if len(inOffice) > 0 {
		usersText := &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "",
		}

		for _, v := range inOffice {
			usersText.Text = fmt.Sprintf("%s %s", usersText.Text, v.Username)
		}
		usersTexts = append(usersTexts, usersText)
	}

	return slack.NewSectionBlock(
		&slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "*" + strings.Title(string(day)) + "*",
		},
		usersTexts,
		slack.NewAccessory(
			slack.NewButtonBlockElement(
				fmt.Sprintf("%d-toggle", weekStart.Unix()),
				string(day),
				&slack.TextBlockObject{
					Type:  slack.PlainTextType,
					Text:  getDayEmoji(day),
					Emoji: true,
				},
			),
		),
	)
}

func getDayEmoji(day Day) string {
	switch day {
	case DayTuesday:
		return ":tinyhat:"
	default:
		return fmt.Sprintf(":alphabet-yellow-%s:", strings.ToLower(string(string(day)[0])))
	}
}
