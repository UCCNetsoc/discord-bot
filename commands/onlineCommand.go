package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/bwmarrin/discordgo"
)

// https://wiki.vg/Server_List_Ping
// 1B - VarInt - Size of packet (length excluding this byte)
// 00 - VarInt - Packet ID (Handshaking)
// E0 05 - VarInt - Protocol version (734)
// 14 - VarInt - Server address string length
// 6D 69 6E 65 63 72 61 66 74 2E 6E 65 74 73 6F 63 2E 63 6F 2E - String - minecraft.netsoc.co.
// 04 AA - Unsigned Short - Port (1194)
// 01 - VarInt - Next state (1 for status)
var handshake = []byte{
	0x1b, 0x00, 0xe0, 0x05, 0x14, 0x6d, 0x69, 0x6e, 0x65, 0x63, 0x72, 0x61, 0x66, 0x74,
	0x2e, 0x6e, 0x65, 0x74, 0x73, 0x6f, 0x63, 0x2e, 0x63, 0x6f, 0x2e, 0x04, 0xaa, 0x01}

// Response of Server List Ping query
type Response struct {
	Version     Version
	Players     Players
	Description Description
	Favicon     string
}

// Version ...
type Version struct {
	Name     string
	Protocol int
}

// Players ...
type Players struct {
	Max    int
	Online int
	Sample []Player
}

// Player ...
type Player struct {
	Name string
	ID   string
}

// Description ...
type Description struct {
	Text string
}

// check the number of people online in minecraft.netsoc.co
func online(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	response, err := Query()
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Unable to get player count at the moment. @sysadmins if issues persist")
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d players online right now", response.Players.Online))
	}
}

// Query Server List Ping
func Query() (Response, error) {
	res := Response{}

	conn, err := net.Dial("tcp", "minecraft.netsoc.co:1194")
	if err != nil {
		return res, err
	}
	defer conn.Close()

	// Handshake - https://wiki.vg/Server_List_Ping#Handshake
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	_, err = conn.Write(handshake)
	if err != nil {
		return res, err
	}

	// Request - https://wiki.vg/Server_List_Ping#Request
	_, err = conn.Write([]byte{0x01, 0x00})

	// Calculate VarInt length of packet
	var buf bytes.Buffer
	pktLen := int64(0)
	for shift := int64(0); ; shift++ {
		io.CopyN(&buf, conn, 1)
		val := int64(buf.Next(1)[0])
		pktLen = (val << (shift * 7)) | pktLen
		if val>>7 == 0 {
			break
		}
	}

	// Server response - https://wiki.vg/Server_List_Ping#Response
	io.CopyN(&buf, conn, pktLen)

	// Packet starts with two VarInts; Packet ID and Data Length
	// https://wiki.vg/Protocol#VarInt_and_VarLong
	for skip := 0; skip < 2; skip++ {
		for buf.Next(1)[0]>>7 != 0 {
		}
	}

	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
