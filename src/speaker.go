// speaker.go version 4.0.0

package main

import "fmt"
import "os"
import "os/exec"
import "time"

import "github.com/bwmarrin/discordgo"

func main() {

	buf, err := os.ReadFile("nomic/key")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.ReadFile: %s\n", err.Error())
		os.Exit(1)
	}

	key := string(buf[:len(buf) - 1])

	discord, err := discordgo.New("Bot " + key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "discordgo.New: %s\n", err.Error())
		os.Exit(1)
	}

	discord.Open()

	// DM Handler
	discord.AddHandler( func(discord *discordgo.Session,
			event *discordgo.MessageCreate) {

		me, _ := discord.User("@me")
		if event.Message.GuildID == "" &&
				event.Message.Author.ID != me.ID {

			buf, err := os.ReadFile("nomic/phase")
			if err != nil {
				fmt.Fprintf(os.Stderr, "os.ReadFile: %s\n", err.Error())
				os.Exit(1)
			}

			cmd := exec.Command("bin/" + string(buf[:len(buf) - 1]),
			                    "msg",
			                    event.Message.ID)
			cmd.Env = append(os.Environ(), "key=" + key,
			                 "channel=" + event.Message.ChannelID,
			                 "player=" + event.Message.Author.ID)
			cmd.Run()
		}
	} )

	// Vote Handler
	discord.AddHandler( func(discord *discordgo.Session,
			event *discordgo.MessageReactionAdd) {

		buf, err := os.ReadFile("nomic/poll-msg-id")
		if err == nil {
			if event.MessageReaction.MessageID == string(buf[:len(buf) - 1]) {

				buf, err := os.ReadFile("nomic/phase")
				if err != nil {
					fmt.Fprintf(os.Stderr, "os.ReadFile: %s\n", err.Error())
					os.Exit(1)
				}

				cmd := exec.Command("bin/" + string(buf[:len(buf) - 1]),
				                    "vote",
				                    event.MessageReaction.Emoji.Name)
				cmd.Env = append(os.Environ(), "key=" + key)
				cmd.Run()
			}
		}
	})

	// 48-hour pings
	for true {
		delay, _ := time.ParseDuration("48h")
		time.Sleep(delay)

		file, err := os.Open("nomic/phase")
		if err != nil {
			fmt.Fprintf(os.Stderr, "os.Open: %s\n", err.Error())
			os.Exit(1)
		}

		var buf []byte
		file.Read(buf)
		cmd := exec.Command("bin/" + string(buf[:len(buf) - 1]), "reminder")
		cmd.Env = append(os.Environ(), "key=" + key)
		cmd.Run()
	}
}

