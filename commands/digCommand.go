package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Strum355/log"
	"github.com/miekg/dns"

	"github.com/bwmarrin/discordgo"
)

func dig(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	args := i.ApplicationCommandData().Options

	domain := args[1].StringValue() + "."

	resolver := "1.1.1.1"
	if len(i.ApplicationCommandData().Options) == 3 {
		resolver = args[2].StringValue()
	}

	var (
		client dns.Client
		msg    dns.Msg

		resp *dns.Msg
		time time.Duration
		err  error

		recordType string
	)

	switch args[0].IntValue() {
	case 0:
		recordType = "A"
		msg.SetQuestion(domain, dns.TypeA)
	case 1:
		recordType = "NS"
		msg.SetQuestion(domain, dns.TypeNS)
	case 2:
		recordType = "CNAME"
		msg.SetQuestion(domain, dns.TypeCNAME)
	case 3:
		recordType = "SRV"
		msg.SetQuestion(domain, dns.TypeSRV)
	case 4:
		recordType = "TXT"
		msg.SetQuestion(domain, dns.TypeTXT)
	}

	func() {
		defer func() {
			if err != nil {
				log.WithContext(ctx).
					WithError(err).
					WithFields(log.Fields{
						"tcp":  client.Net == "tcp",
						"time": time.String(),
					}).
					Error("error querying DNS record")
			}
		}()

		resp, time, err = client.ExchangeContext(ctx, &msg, resolver+":53")
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Encountered error: %v", err),
					Flags:   1 << 6,
				},
			})
			return
		}

		if resp.Truncated {
			client.Net = "tcp"
			resp, time, err = client.ExchangeContext(ctx, &msg, resolver+":53")
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Encountered error: %v", err),
						Flags:   1 << 6,
					},
				})
				return
			}
		}
	}()

	// return here because returns above are anon function scoped
	if err != nil {
		return
	}

	log.WithContext(ctx).
		WithFields(log.Fields{
			"responses": fmt.Sprintf("%#v", resp),
			"tcp":       client.Net == "tcp",
			"answers":   resp.Answer,
			"time":      time.String(),
		}).
		Info("got DNS response")

	var b strings.Builder
	b.WriteString("```\n")

	if len(resp.Answer) == 0 {
		b.WriteString("No results\n")
	}

	for _, r := range resp.Answer {
		b.WriteString(fmt.Sprintf("%s\t%d\t%s\t", domain, r.Header().Ttl, recordType))
		switch rec := r.(type) {
		case *dns.A:
			b.WriteString(fmt.Sprintf("%s\n", rec.A.String()))
		case *dns.NS:
			b.WriteString(fmt.Sprintf("%s\n", rec.Ns))
		case *dns.CNAME:
			b.WriteString(fmt.Sprintf("%s\n", rec.Target))
		case *dns.SRV:
			b.WriteString(fmt.Sprintf("%d  %d  %d  %s\n", rec.Priority, rec.Weight, rec.Port, rec.Target))
		case *dns.TXT:
			for _, txt := range rec.Txt {
				b.WriteString(fmt.Sprintf("%s\n", txt))
			}
		}
	}

	b.WriteString(fmt.Sprintf("\nResponse time: %s\n", time.String()))

	b.WriteString("```")
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: b.String(),
		},
	})
	if err != nil {
		log.WithError(err)
	}
}
