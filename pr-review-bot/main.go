package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	// Replace with your Slack Bot Token
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Load .env file error")
	}

	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		log.Fatal("SLACK_BOT_TOKEN is required")
	}

	api := slack.New(token)

	// Keep your existing /slack/commands handler
	http.HandleFunc("/slack/commands", handleSlashCommand(api))

	// Add new handler for interactions (modal submissions)
	http.HandleFunc("/slack/interactivity", handleInteractivity(api))

	http.HandleFunc("/slack/events", handleEvents(api))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}

func handleInteractivity(api *slack.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload slack.InteractionCallback
		err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
		if err != nil {
			log.Printf("Error parsing interaction payload: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if payload.Type == slack.InteractionTypeViewSubmission {
			// Extract values from the submission
			values := payload.View.State.Values
			channelID := payload.View.PrivateMetadata
			prURL := values["pr_url_block"]["pr_url"].Value
			description := values["description_block"]["description"].Value
			reviewers := values["reviewers_block"]["reviewers"].SelectedUsers

			// Create message blocks
			blocks := []slack.Block{
				// Header
				slack.NewHeaderBlock(
					slack.NewTextBlockObject(slack.PlainTextType, "🔍 New PR Review Request", false, false),
				),
				// PR Link
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf("*PR Link:* <%s>", prURL),
						false, false,
					),
					nil, nil,
				),
				// Description
				slack.NewSectionBlock(
					slack.NewTextBlockObject(
						slack.MarkdownType,
						fmt.Sprintf("*Description:*\n%s", description),
						false, false,
					),
					nil, nil,
				),
			}

			// Add reviewers section if any were selected
			if len(reviewers) > 0 {
				var mentionText string
				for i, reviewer := range reviewers {
					if i > 0 {
						mentionText += ", "
					}
					mentionText += fmt.Sprintf("<@%s>", reviewer)
				}
				blocks = append(blocks,
					slack.NewSectionBlock(
						slack.NewTextBlockObject(
							slack.MarkdownType,
							fmt.Sprintf("*Reviewers:* %s", mentionText),
							false, false,
						),
						nil, nil,
					),
				)
			}

			// Add divider and reaction guide
			blocks = append(blocks,
				slack.NewDividerBlock(),
				slack.NewContextBlock(
					"CONTEXT BLOCK",
					slack.NewTextBlockObject(slack.MarkdownType, "👀 = reviewing | ✅ = approved", false, false),
				),
			)

			// Post message to channel
			_, _, err := api.PostMessage(
				// payload.User.ID, // DM to user who submitted
				channelID,
				slack.MsgOptionBlocks(blocks...),
				slack.MsgOptionText("New PR Review Request", false),
			)
			if err != nil {
				log.Printf("Error posting message: %v", err)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleSlashCommand(api *slack.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Printf("Error parsing slash command: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if s.Command == "/pr" {
			modal := slack.ModalViewRequest{
				Type: slack.VTModal,
				Title: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Submit PR for Review",
				},
				Submit: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Submit",
				},
				Close: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: "Cancel",
				},
				Blocks: slack.Blocks{
					BlockSet: []slack.Block{
						// PR URL Input
						slack.InputBlock{
							Type:    slack.MBTInput,
							BlockID: "pr_url_block",
							Label: &slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: "Pull Request URL",
							},
							Element: &slack.PlainTextInputBlockElement{
								Type:     slack.METPlainTextInput,
								ActionID: "pr_url",
							},
						},
						// Description Input
						slack.InputBlock{
							Type:    slack.MBTInput,
							BlockID: "description_block",
							Label: &slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: "Description",
							},
							Element: &slack.PlainTextInputBlockElement{
								Type:      slack.METPlainTextInput,
								ActionID:  "description",
								Multiline: true,
							},
						},
						// Reviewers Multi-select
						slack.InputBlock{
							Type:    slack.MBTInput,
							BlockID: "reviewers_block",
							Label: &slack.TextBlockObject{
								Type: slack.PlainTextType,
								Text: "Reviewers",
							},
							Element: &slack.MultiSelectBlockElement{
								Type:     slack.MultiOptTypeUser,
								ActionID: "reviewers",
							},
						},
					},
				},
				PrivateMetadata: s.ChannelID,
			}

			_, err := api.OpenView(s.TriggerID, modal)
			if err != nil {
				log.Printf("Error opening modal: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleEvents(api *slack.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := make(map[string]interface{})
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.NewDecoder(bytes.NewReader(rawBody)).Decode(&body)
		if err != nil {
			log.Printf("Error parsing event body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Handle URL verification challenge
		if body["type"] == "url_verification" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(body["challenge"].(string)))
			return
		}

		// Parse event
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(rawBody), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Printf("Error parsing event: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if eventsAPIEvent.Type != slackevents.CallbackEvent {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		innerEvent := eventsAPIEvent.InnerEvent
		if reaction, ok := innerEvent.Data.(*slackevents.ReactionAddedEvent); ok {
			user, err := api.GetUserInfo(reaction.User)
			if err != nil {
				log.Printf("Error getting user info: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			switch reaction.Reaction {
			case "eyes":
				log.Printf("%s is reviewing the PR", user.RealName)

			case "white_check_mark":
				// Post approval message in thread
				_, _, err := api.PostMessage(
					reaction.Item.Channel,
					slack.MsgOptionText(fmt.Sprintf("✅ PR approved by %s", user.RealName), false),
					slack.MsgOptionTS(reaction.Item.Timestamp), // This creates a thread
				)
				if err != nil {
					log.Printf("Error posting approval message: %v", err)
				}
				log.Printf("%s approved the PR", user.RealName)
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}