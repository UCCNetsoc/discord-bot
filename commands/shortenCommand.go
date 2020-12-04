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
		return
	}

	params := strings.Split(m.Content, " ")
	if len(params) == 1 {
		s.ChannelMessageSend(m.ChannelID, "Missing argument: original-url")
		return
	}

	/* Requested shortened url are valid in any of these forms:
	http://xx.yy/slug
	https://xx.yy/slug
	xx.yy/slug
	slug (default domain will be used here)
	*/

	originalURL := params[1]
	domain := viper.GetString("shorten.host")
	slug := ""
	if len(params) > 2 {
		re := regexp.MustCompile("^http(s)?://")
		urlSplit := strings.Split(re.ReplaceAllString(params[2], ""), "/")
		if len(urlSplit) == 1 { // slug
			slug = strings.TrimSpace(urlSplit[0])
		} else if len(urlSplit) > 1 { // xx.yy/slug
			domain = strings.TrimSpace(urlSplit[0])
			slug = strings.TrimSpace(urlSplit[1])
		} else {
			s.ChannelMessageSend(m.ChannelID, "Invalid short URL format!")
			return
		}

	}

	values := map[string]string{"Slug": slug, "Domain": domain, "Target": originalURL}
	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", "https://"+viper.GetString("shorten.host"), bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not reach URL shortening server.")
		log.WithContext(ctx).WithError(err).Error("Error communicating with shorten server")
		return
	}
	defer resp.Body.Close()

	shortenedURL := "https://" + domain + "/" + slug
	switch resp.StatusCode {
	case 201:
		emb := embed.NewEmbed().SetTitle(shortenedURL)
		emb.AddField("Original URL", originalURL)
		// emb.URL = shortenedURL
		s.ChannelMessageSendEmbed(m.ChannelID, emb.MessageEmbed)
		break
	case 409:
		s.ChannelMessageSend(m.ChannelID, "<"+shortenedURL+"> already exists!")
		break
	default:
		log.WithContext(ctx).WithFields(log.Fields{
			"originalUrl":  originalURL,
			"shortenedUrl": shortenedURL,
			"host":         viper.GetString("shorten.host"),
			"responseCode": resp.Status,
		}).Error("Error while trying to shorten URL!")
		s.ChannelMessageSend(m.ChannelID, "Unexpected error occured: "+resp.Status)
		break
	}

}
