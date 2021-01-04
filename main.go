package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Nonne46/CuteBot/command"
	"github.com/joho/godotenv"

	"github.com/bwmarrin/discordgo"
)

const commandPrefix string = "~"

func main() {
	err := godotenv.Load()
	e("Error loading .env file", err)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	e("error creating Discord session,", err)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	e("error opening connection,", err)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {

	// Игнор ботом самого себя
	if message.Author.ID == session.State.User.ID {
		return
	}

	switch {
	case strings.HasPrefix(message.Content, commandPrefix+"xz"):
		command.Xz(session, message)
	case strings.HasPrefix(message.Content, commandPrefix+"o"):
		command.Optimize(session, message)
	}
}

func e(msg string, err error) {
	if err != nil {
		fmt.Printf("%s: %+v", msg, err)
		panic(err)
	}
}
