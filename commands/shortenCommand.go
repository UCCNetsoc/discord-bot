package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/embed"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func shortenCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if committee channel, don't allow in public server
	if !isCommittee(s, m) {
		s.ChannelMessageSend(m.ChannelID, "You must be a committee member to use this command")
		return
	}

	params := strings.Split(m.Content, " ")
	if len(params) < 3 {
		if len(params) == 2 {
			s.ChannelMessageSend(m.ChannelID, "Missing argument: shortened-slug")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "Missing arguments: original-url, shortened-slug")
		return
	}
	method := ""
	if params[1] == "delete" {
		method = "DELETE"
	} else {
		method = "POST"
		// regex for http(s)://*.*...
		reURL := regexp.MustCompile(`^http(s)?:\/\/*[a-zA-Z0-9/\.-]*$`)
		if urlOk := reURL.MatchString(params[1]); !urlOk {
			s.ChannelMessageSend(m.ChannelID, "Invalid URL format.")
			log.WithContext(ctx).Error("URL did not match regex")
			return
		}
	}
	// regex for text characters incl. hyphens
	reSlug := regexp.MustCompile(`^[a-zA-Z-]*$`)
	if ok := reSlug.MatchString(params[2]); !ok {
		s.ChannelMessageSend(m.ChannelID, "Invalid short URL format.")
		log.WithContext(ctx).Error("Slug did not match regex")
		return
	}
	if method == "DELETE" {
		req, err := http.NewRequest("DELETE", "https://"+viper.GetString("shorten.host")+"/"+params[2], nil)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not form request.")
			log.WithContext(ctx).WithError(err).Error("Error forming request")
			return
		}
		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server.")
			log.WithContext(ctx).WithError(err).Error("Error communicating with shorten server")
			return
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			s.ChannelMessageSend(m.ChannelID, "Deleted shortened URL!")
			break
		case http.StatusNotFound:
			s.ChannelMessageSend(m.ChannelID, "Couldn't find given shortened URL.")
		default:
			log.WithContext(ctx).WithFields(log.Fields{
				"method":       method,
				"shortenedUrl": "https://" + viper.GetString("shorten.host") + "/" + params[2],
				"responseCode": resp.Status,
			}).Error("Error while trying to shorten URL!")
			s.ChannelMessageSend(m.ChannelID, "Unexpected error occured: "+resp.Status)
			break
		}
		return
	}
	if method == "POST" {
		values := map[string]string{"slug": params[2], "url": params[1]}
		jsonValue, _ := json.Marshal(values)
		req, err := http.NewRequest("POST", "https://"+viper.GetString("shorten.host"), bytes.NewBuffer(jsonValue))
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not form request.")
			log.WithContext(ctx).WithError(err).Error("Error forming request")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(viper.GetString("shorten.username"), viper.GetString("shorten.password"))
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server.")
			log.WithContext(ctx).WithError(err).Error("Error communicating with shorten server")
			return
		}
		switch resp.StatusCode {
		case http.StatusCreated:
			emb := embed.NewEmbed().SetTitle(params[2])
			emb.AddField("Original URL", params[1])
			emb.AddField("Shortened URL", "https://"+strings.Split(viper.GetString("shorten.host"), "api")[0]+params[2])
			// emb.URL = shortenedURL
			s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
			break
		case http.StatusConflict:
			s.ChannelMessageSend(m.ChannelID, "<"+params[2]+"> already exists!")
			break
		default:
			log.WithContext(ctx).WithFields(log.Fields{
				"method":       method,
				"originalUrl":  params[1],
				"shortenedUrl": "https://" + viper.GetString("shorten.host") + "/" + params[2],
				"responseCode": resp.Status,
			}).Error("Error while trying to shorten URL!")
			s.ChannelMessageSend(m.ChannelID, "Unexpected error occured: "+resp.Status)
			break
		}
	}
}
