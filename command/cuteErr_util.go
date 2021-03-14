package command

import (
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
)

// CuteErr ...
type CuteErr struct {
	s *discordgo.Session
	m *discordgo.MessageCreate
}

// WrongCommand выводит смешное сообщение когда пользователь неправильно воспользовался коммандой
func (e CuteErr) WrongCommand() {
	wrongCommandUse := []string{
		"Спорим не сможешь корректно написать?",
		"Внимательней нужно быть",
		"Херню пишешь",
		"Это не так работает",
		"https://cdn.discordapp.com/emojis/768954786938486845.png?v=1",
	}

	indexJoke := rand.Intn(len(wrongCommandUse))
	randJoke := wrongCommandUse[indexJoke]
	e.s.ChannelMessageSend(e.m.ChannelID, randJoke)
}

// InsideError выводит смешное сообщение в случае внутренней ошибки
func (e CuteErr) InsideError(err error) {
	wrongCommandUse := []string{
		"Лень",
		"*Спонсировано ШУЕ*",
		"Иди нахуй, короче",
		"Если вкратце, то иди нахуй",
	}

	fmt.Printf("%s: %+v", "External command", err)

	indexJoke := rand.Intn(len(wrongCommandUse))
	randJoke := wrongCommandUse[indexJoke]
	e.s.ChannelMessageSend(e.m.ChannelID, randJoke)
}
