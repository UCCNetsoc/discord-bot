package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

type Link struct {
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func shortenCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	topLevelArgs := i.ApplicationCommandData().Options[0]
	subLevelArgs := topLevelArgs.Options
	var method string

	switch topLevelArgs.Name {
	case "create":
		method = "POST"

		var shortenedURL string
		var originalURL string

		originalURL = fmt.Sprintf("%v", subLevelArgs[0].Value)

		if len(subLevelArgs) > 1 {
			shortenedURL = fmt.Sprintf("%v", subLevelArgs[1].Value)
		}

		reURL := regexp.MustCompile(`^http(s)?://.*$`)
		if ok := reURL.MatchString(originalURL); !ok {
			log.WithContext(ctx).Error("URL did not match regex")
			InteractionResponseError(s, i, "Invalid URL", true)
			return
		}
		values := make(map[string]interface{}, 2)
		values["url"], values["slug"] = originalURL, shortenedURL
		encoded, err := json.Marshal(values)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to create encoded json")
			InteractionResponseError(s, i, "Could not create request", true)
			return
		}

		req, err := http.NewRequest(method, viper.GetString("shorten.host"), bytes.NewBuffer((encoded)))
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to make request object")
			InteractionResponseError(s, i, "Could not create request", true)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error communicating with server")
			InteractionResponseError(s, i, "Could not reach URL shortening server", true)
			return
		}

		switch resp.StatusCode {
		case http.StatusCreated:
			bd, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.WithContext(ctx).WithError(err)
				InteractionResponseError(s, i, "Failed to decode json", true)
				return
			}
			link := &Link{}

			err = json.Unmarshal(bd, link)
			if err != nil {
				log.WithContext(ctx).WithError(err)
				InteractionResponseError(s, i, "Failed to decode json", true)
				return
			}

			emb := embed.NewEmbed().SetTitle(link.Slug)
			emb.AddField("Original URL", link.URL)
			emb.AddField("Shortened URL", viper.GetString("shorten.public.host")+"/"+link.Slug)

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{emb.MessageEmbed},
				},
			})
			if err != nil {
				log.WithContext(ctx).WithError(err)
			}
			return
		case http.StatusConflict:
			if len(subLevelArgs) < 2 {
				InteractionResponseError(s, i, "Failed to shortened link, please try again", false)
			} else {
				InteractionResponseError(s, i, "Failed to shorten link, try a different shortened-slug", false)
			}
		default:
			log.WithContext(ctx).WithFields(log.Fields{
				"method":       method,
				"originalUrl":  originalURL,
				"shortenedUrl": viper.GetString("shorten.public.host") + "/" + shortenedURL,
				"responseCode": resp.Status,
			}).Error("Error while trying to shorten URL!")
			InteractionResponseError(s, i, resp.Status, true)
			return
		}
		InteractionResponseError(s, i, "Missing argument original-url", true)

	case "delete":
		method = "DELETE"
		req, err := http.NewRequest(method, fmt.Sprintf("%s/%v", viper.GetString("shorten.host"), subLevelArgs[0].Value), nil)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to make request object")
			InteractionResponseError(s, i, "Could not create request", true)
			return
		}

		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error communicating with server")
			InteractionResponseError(s, i, "Could not reach URL shortening server", true)
			return
		}

		switch resp.StatusCode {
		case http.StatusAccepted:
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Deleted %v", subLevelArgs[0].Value),
				},
			})
			if err != nil {
				log.WithContext(ctx).WithError(err)
			}
		case http.StatusNotFound:
			InteractionResponseError(s, i, fmt.Sprintf("Shortened link %v does not exist", subLevelArgs[0].Value), true)
		default:
			log.WithContext(ctx).WithFields(log.Fields{
				"method": method,

				"responseCode": resp.Status,
			}).Error("Error while trying to delete shorten URL")
			InteractionResponseError(s, i, resp.Status, true)
		}

	default:
		req, err := http.NewRequest("GET", viper.GetString("shorten.host")+"/links", nil)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to make request object")
			InteractionResponseError(s, i, "Could not create request", true)
			return
		}

		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Error communicating with server")
			InteractionResponseError(s, i, "Could not reach URL shortening server", true)
			return
		}

		data := []Link{}
		err = json.NewDecoder(resp.Body).Decode(&data)

		if err != nil {
			log.WithContext(ctx).WithError(err).Error("Failed to unmarshall json data")
			InteractionResponseError(s, i, "Could not parse data", true)
			return
		}

		emb := embed.NewEmbed().SetTitle("Links")

		for _, link := range data {
			emb.AddField(link.Slug, link.URL)
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{emb.MessageEmbed},
			},
		})
		if err != nil {
			log.WithContext(ctx).WithError(err)
		}
	}
}
