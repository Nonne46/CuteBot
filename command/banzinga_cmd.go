package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type banInfo struct {
	Username   string `json:"keyname"`
	Ckey       string `json:"ckey"`
	Registered string `json:"registered"`
	Updated    string `json:"date_updated"`
	Onyx       string `json:"onyx"`
}

type Onyxs struct {
	Ckey     string
	BanTime  string
	Desc     string
	Pedal    string
	Unbanned string
	Reason   string
}

func GetData(s *discordgo.Session, m *discordgo.MessageCreate, ckey string) {
	cErr := CuteErr{s, m}

	url := "https://banzinga.station13.ru/api/?ckey=" + ckey
	banzinga := http.Client{
		Timeout: time.Minute, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		cErr.InsideError(err)
	}

	req.Header.Set("User-Agent", "CuteBot")

	res, getErr := banzinga.Do(req)
	if getErr != nil {
		cErr.InsideError(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		cErr.InsideError(readErr)
	}

	rest := bytes.Compare(body, []byte("0"))
	if rest == 0 {
		cErr.WrongCommand()
		return
	}

	var player banInfo
	json.Unmarshal([]byte(body), &player)

	embed := &discordgo.MessageEmbed{
		Title: "Main Info",
		Color: 0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Username", Value: fmt.Sprint(player.Username), Inline: false},
			{Name: "Registered", Value: fmt.Sprint(player.Registered), Inline: true},
			{Name: "Updated", Value: fmt.Sprint(player.Updated), Inline: true},
		},
	}
	s.ChannelMessageSendEmbed(m.ChannelID, embed)

	if player.Onyx != "[]" {
		playerOnyx := onyxParse(player)
		color := 0x0
		if playerOnyx.Unbanned == "no" {
			color = 0xFF0000
		} else {
			color = 0x00FF00
		}
		embed := &discordgo.MessageEmbed{
			Title: "Onyx Bans",
			Color: color,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Pedal", Value: fmt.Sprint(playerOnyx.Pedal), Inline: true},
				{Name: "Ban time", Value: fmt.Sprint(playerOnyx.BanTime), Inline: true},
				{Name: "Desc", Value: fmt.Sprint(playerOnyx.Desc), Inline: true},
				{Name: "Reason", Value: fmt.Sprint(playerOnyx.Reason), Inline: false},
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
	}
}

func onyxParse(p banInfo) Onyxs {
	var playerOnyx Onyxs

	xz := strings.ReplaceAll(p.Onyx, "\\\"", "\"")
	xz = strings.ReplaceAll(xz, "\\\\\"", "\"")
	xz = strings.ReplaceAll(xz, "[[", "")
	xz = strings.ReplaceAll(xz, "]]", "")
	xz = strings.ReplaceAll(xz, "\\n", "\n")
	xz = strings.ReplaceAll(xz, "]", "")
	xz = strings.ReplaceAll(xz, "[", "")

	xzz := strings.Split(xz, "\",\"")

	for i := 0; i < len(xzz); i++ {
		xzz[i] = strings.ReplaceAll(xzz[i], "\"", "")
	}
	//
	playerOnyx.Ckey = xzz[0]
	playerOnyx.BanTime = xzz[1]
	playerOnyx.Desc = xzz[2]
	playerOnyx.Pedal = xzz[3]
	playerOnyx.Unbanned = xzz[4]
	playerOnyx.Reason = xzz[6]
	return playerOnyx
}
