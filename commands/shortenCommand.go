package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

type Link struct {
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func shortenCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if committee channel, don't allow in public server
	if !isCommittee(s, m) {
		s.ChannelMessageSend(m.ChannelID, "You must be a committee member to use this command")
		return
	}

	params := strings.Split(m.Content, " ")

	if len(params) < 2 {
		req, err := http.NewRequest("GET", viper.GetString("shorten.host")+"/links", nil)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not create request")
			log.WithContext(ctx).WithError(err).Error("Failed to make request object")
			return
		}

		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server")
			log.WithContext(ctx).WithError(err).Error("Error communicating with server")
			return
		}

		data := []Link{}
		err = json.NewDecoder(resp.Body).Decode(&data)

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not parse data")
			log.WithContext(ctx).WithError(err).Error("Failed to unmarshall json data")
			return
		}

		emb := embed.NewEmbed().SetTitle("Links")

		for _, link := range data {
			emb.AddField(link.Slug, link.URL)
		}

		s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
		return
	}

	var method string

	if strings.ToLower(params[1]) == "delete" {
		method = "DELETE"
	} else {
		method = "POST"
	}

	if method == "DELETE" {
		if len(params) > 1 {

			// for each shortened link to delete, make a new request
			for _, i := range params[2:] {
				req, err := http.NewRequest(method, viper.GetString("shorten.host")+"/"+i, nil)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not create request")
					log.WithContext(ctx).WithError(err).Error("Failed to make request object")
					return
				}

				req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
				client := http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server")
					log.WithContext(ctx).WithError(err).Error("Error communicating with server")
					return
				}

				switch resp.StatusCode {
				case http.StatusAccepted:
					s.ChannelMessageSend(m.ChannelID, "Deleted \""+i+"\"")
				case http.StatusNotFound:
					s.ChannelMessageSend(m.ChannelID, "Shortened link \""+i+"\" does not exist")
				default:
					log.WithContext(ctx).WithFields(log.Fields{
						"method": method,

						"responseCode": resp.Status,
					}).Error("Error while trying to delete shorten URL")
					s.ChannelMessageSend(m.ChannelID, "Unexpected error occured: "+resp.Status)
				}

			}
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Missing argument: shortened-slug")
		return
	} else {
		if len(params) >= 2 {
			// POST request
			method = "POST"

			var shortenedURL string

			if len(params) > 2 {
				shortenedURL = params[2]
			}

			reURL := regexp.MustCompile(`^http(s)?:\/\/*[a-zA-Z0-9/\.-_]*$`)
			if ok := reURL.MatchString(params[1]); !ok {
				s.ChannelMessageSend(m.ChannelID, "Invalid URL")
				log.WithContext(ctx).Error("URL did not match regex")
				return
			}
			values := make(map[string]interface{}, 2)
			values["url"], values["slug"] = params[1], shortenedURL
			encoded, err := json.Marshal(values)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not create request")
				log.WithContext(ctx).WithError(err).Error("Failed to create encoded json")
				return
			}

			req, err := http.NewRequest(method, viper.GetString("shorten.host"), bytes.NewBuffer((encoded)))
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not create request")
				log.WithContext(ctx).WithError(err).Error("Failed to make request object")
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))

			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server")
				log.WithContext(ctx).WithError(err).Error("Error communicating with server")
				return
			}

			switch resp.StatusCode {
			case http.StatusCreated:

				bd, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Failed to decode json")
					log.WithContext(ctx).WithError(err)
					return
				}
				link := &Link{}

				err = json.Unmarshal(bd, link)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Failed to decode json")
					log.WithContext(ctx).WithError(err)
					return
				}

				emb := embed.NewEmbed().SetTitle(link.Slug)
				emb.AddField("Original URL", link.URL)
				emb.AddField("Shortened URL", viper.GetString("shorten.public.host")+"/"+link.Slug)
				s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
				return

			case http.StatusConflict:
				var returnString string
				if len(params) < 3 {
					returnString = "Failed to shortened link, please try again"
				} else {
					returnString = "Failed to shorten link, try a different shortened-slug"
				}
				s.ChannelMessageSend(m.ChannelID, returnString)
			default:
				log.WithContext(ctx).WithFields(log.Fields{
					"method":       method,
					"originalUrl":  params[1],
					"shortenedUrl": viper.GetString("shorten.public.host") + "/" + params[2],
					"responseCode": resp.Status,
				}).Error("Error while trying to shorten URL!")
				s.ChannelMessageSend(m.ChannelID, "Unexpected error occured: "+resp.Status)
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Missing argument original-url")
		}
	}
}
