package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Strum355/log"
	"github.com/miekg/dns"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func digCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Fields(strings.TrimPrefix(m.Content, viper.GetString("bot.prefix")+"dig"))

	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Missing arguments: TYPE DOMAIN <@RESOLVER>")
	}

	domain := args[1] + "."

	resolver := "1.1.1.1"
	if len(args) == 3 {
		resolver = strings.TrimPrefix(args[2], "@")
	}

	var (
		client dns.Client
		msg    dns.Msg
	)

	switch args[0] {
	case "A":
		msg.SetQuestion(domain, dns.TypeA)
	case "NS":
		msg.SetQuestion(domain, dns.TypeNS)
	case "CNAME":
		msg.SetQuestion(domain, dns.TypeCNAME)
	case "SRV":
		msg.SetQuestion(domain, dns.TypeSRV)
	case "TXT":
		msg.SetQuestion(domain, dns.TypeTXT)
	}

	resp, time, err := client.ExchangeContext(ctx, &msg, resolver+":53")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Encountered error: %v", err))
		return
	}

	if resp.Truncated {
		client.Net = "tcp"
		resp, time, err = client.ExchangeContext(ctx, &msg, resolver+":53")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Encountered error: %v", err))
			return
		}
	}

	var b strings.Builder
	b.WriteString("```\n")

	if len(resp.Answer) == 0 {
		b.WriteString("No results\n")
	}

	log.Info(fmt.Sprintf("%t", resp.Truncated))

	for _, r := range resp.Answer {
		b.WriteString(fmt.Sprintf("%s\t%d\t%s\t", domain, r.Header().Ttl, args[0]))
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

	b.WriteString(fmt.Sprintf("\nResponse time: %dms\n", time.Milliseconds()))

	b.WriteString("```")
	s.ChannelMessageSend(m.ChannelID, b.String())
}
