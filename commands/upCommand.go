package commands

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/UCCNetsoc/discord-bot/embed"

	"github.com/bwmarrin/discordgo"
	"github.com/matryer/try"
	"github.com/spf13/viper"
)

type statusCheck struct {
	Site    string
	Time    int64
	Success bool
	Error   error
}

// Up command to check the status of various websites hosted on Netsoc servers
func checkUpCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	sites := strings.Split(viper.GetString("netsoc.sites"), ",")
	// Run on a separate goroutine to not block bot
	checkStatuses(s, i, sites)
}

func checkStatuses(s *discordgo.Session, i *discordgo.InteractionCreate, sites []string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Checking status...",
		},
	})
	if err != nil {
		log.WithError(err)
		return
	}
	// Create a channel to receive the status checks, whenever they complete
	statuses := make(chan statusCheck)
	// Run each status check on a separate goroutine as to not block each other
	for _, site := range sites {
		go checkStatus(site, 3, statuses)
	}

	results := make([]statusCheck, 0)
	for i := 0; i < len(sites); i++ {
		results = append(results, <-statuses)
	}

	emb := embed.NewEmbed().SetTitle("Website Statuses")
	for _, result := range results {
		title := ""
		if result.Success {
			title = "🆗 " + result.Site
		} else {
			title = "🔥 " + result.Site
		}
		emb = emb.AddField(title, fmt.Sprintf("Up: %v\nLatency: %dms\nError: %+v", result.Success, result.Time, result.Error))
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{emb.MessageEmbed},
	})
	if err != nil {
		log.WithError(err)
	}
}

// Check the up status of a website, returning an error if not up - checks if can connect and response is 200
func checkStatus(site string, retryCount int, statuses chan statusCheck) {
	var timeTaken int64 = 0
	client := http.Client{Timeout: 5 * time.Second}

	err := try.Do(func(attempt int) (bool, error) {
		startTime := time.Now()
		resp, err := client.Get(site)
		timeTaken = time.Now().Sub(startTime).Milliseconds()

		if resp != nil {
			if resp.StatusCode >= 400 {
				err = fmt.Errorf("error status code %d returned", resp.StatusCode)
			}
		}

		if err != nil {
			// If an error has been found, sleep an increasing amount of time before trying again
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		return attempt < retryCount, err
	})
	statuses <- statusCheck{site, timeTaken, err == nil, err}
}
