package commands

import (
	"context"
	"fmt"

	"github.com/Strum355/log"

	"github.com/bwmarrin/discordgo"
)

func members(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	responseContent, err := createMembersResponse(s, i)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("failed to get members")
		return
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         responseContent,
			AllowedMentions: &discordgo.MessageAllowedMentions{Parse: nil},
		},
	}

	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		log.WithError(err)
		return
	}
}

func createMembersResponse(s *discordgo.Session, i *discordgo.InteractionCreate) (responseContent string, err error) {
	role := i.ApplicationCommandData().Options[0].RoleValue(s, i.GuildID)

	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		return responseContent, err
	}

	if role.Name == "@everyone" {
		return fmt.Sprintf("Number of members is %d", len(members)), nil
	}

	var num int
	for _, member := range members {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				num++
			}
		}
	}
	return fmt.Sprintf("Number of members in %s is %d", role.Name, num), nil
}
