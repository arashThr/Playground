package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/slack-go/slack"
)

// Add the reminder system
func startReminderSystem(api *slack.Client, db *sql.DB) {
	log.Printf("Starting reminder system")
	s := gocron.NewScheduler()

	// Run every day at 9 AM
	// s.Every(1).Day().At("09:00").Do(func() {
	s.Every(5).Seconds().Do(func() {
		log.Printf("Running reminder system")
		prs, err := getPendingPRs(db)
		if err != nil {
			log.Printf("Error getting pending PRs: %v", err)
			return
		}

		for _, pr := range prs {
			// daysSinceCreation := int(time.Since(pr.CreatedAt).Hours() / 24)
			daysSinceCreation := int(time.Since(pr.CreatedAt).Seconds())

			// Skip PRs less than 1 day old
			if daysSinceCreation < 1 {
				continue
			}

			var mentionUsers []string

			if daysSinceCreation >= 3 {
				// After 3 days, mention everyone
				mentionUsers = []string{"<!channel>"}
			} else {
				// Get users who reacted with eyes
				reactions, err := api.GetReactions(slack.ItemRef{
					Channel:   pr.ChannelID,
					Timestamp: pr.MessageTS,
				}, slack.NewGetReactionsParameters())

				if err != nil {
					log.Printf("Error getting reactions: %v", err)
					continue
				}

				// Look for eyes reactions
				for _, reaction := range reactions {
					if reaction.Name == "eyes" {
						for _, user := range reaction.Users {
							mentionUsers = append(mentionUsers, fmt.Sprintf("<@%s>", user))
						}
					}
				}

				// If no eyes reactions, mention original reviewers
				if len(mentionUsers) == 0 {
					for _, reviewer := range pr.Reviewers {
						mentionUsers = append(mentionUsers, fmt.Sprintf("<@%s>", reviewer))
					}
				}
			}

			// Create reminder message
			text := fmt.Sprintf("🔔 *Reminder:* PR needs review\n<%s|Open PR>\n", pr.PRUrl)
			if len(mentionUsers) > 0 {
				text += "Hey " + strings.Join(mentionUsers, ", ") + "! "
				if daysSinceCreation >= 3 {
					text += "This PR has been waiting for review for 3+ days."
				} else {
					text += "This PR is awaiting your review."
				}
			}

			// Post reminder as thread reply
			_, _, err = api.PostMessage(
				pr.ChannelID,
				slack.MsgOptionText(text, false),
				slack.MsgOptionTS(pr.MessageTS),
			)
			if err != nil {
				log.Printf("Error posting reminder: %v", err)
			}
		}
	})

	s.Start()
}
