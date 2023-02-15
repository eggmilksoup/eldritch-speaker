// rcp.go version 4.0.0

package main

import "fmt"
import "os"
import "strconv"
import "time"

import "github.com/bwmarrin/discordgo"

func main() {

	if (len(os.Args) < 3) &&
	   (len(os.Args) < 2 || os.Args[1] != "reminder") {
		fmt.Fprintf(os.Stderr, "usage: %s vote emoji\n" +
		                       "       %s reminder", os.Args[0])
		os.Exit(1)
	}

	if os.Args[1] == "msg" {
		return
	}

	buf, _ := os.ReadFile("nomic/poll-thread-id")
	thread := string(buf[:len(buf) - 1])
	buf, _ = os.ReadFile("nomic/poll-msg-id")
	msg := string(buf[:len(buf) - 1])

	players, _ := os.ReadDir("nomic/players")
	numplayers := len(players)

	discord, err := discordgo.New("Bot " + os.Getenv("key"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "discordgo.New: %s\n", err.Error())
		os.Exit(1)
	}

	y, _ := discord.MessageReactions(thread, msg, "✅", numplayers, "", "")
	n, _ := discord.MessageReactions(thread, msg, "❌", numplayers, "", "")

	if os.Args[1] == "reminder" {
		players, _ := os.ReadDir("nomic/players")
		
		var novote []string
		for _, player := range players {
			found := false
			buf, _ := os.ReadFile("nomic/players/" + player.Name())
			id := string(buf[:len(buf) - 1])
			for _, vote := range y {
				if vote.ID == id {
					found = true
					break
				}
			}
			if ! found {
				for _, vote := range n {
					if vote.ID == id {
						found = true
						break
					}
				}
			}
			if ! found {
				novote = append(novote, id)
			}
		}
		txt := "The following players have not voted:"
		for _, id := range novote {
			user, _ := discord.User(id)
			txt = txt + "\n" + user.Mention()
		}
		buf, _ := os.ReadFile("nomic/channels/announcements")
		channel := string(buf[:len(buf) - 1])
		discord.ChannelMessageSend(
			channel,
			txt + "\n" + "This is your 48-hour reminder to vote.")
		return
	}

	if len(y) + len(n) == numplayers {
		buf, _ := os.ReadFile("nomic/channels/announcements")
		channel := string(buf[:len(buf) - 1])
		discord.ChannelMessageSend(
			channel,
			"<@&1001702350571974716>, " +
			"all the votes are in!  In 15 minutes the votes will be tallied " +
			"unless someone retracts their vote.")
		os.WriteFile("nomic/phase", []byte("dummy\n"), 0644)

		dur, _ := time.ParseDuration("15m")
		time.Sleep(dur)

		y, _ = discord.MessageReactions(thread, msg, "✅", numplayers, "", "")
		n, _ = discord.MessageReactions(thread, msg, "❌", numplayers, "", "")
		if len(y) + len(n) == numplayers {
			buf, _ := os.ReadFile("nomic/rcp")
			rcp, _ := strconv.Atoi(string(buf))
			if len(n) > len(y) {
				discord.ChannelMessageSend(
					channel,
					"<@&1001702350571974716>, " +
					"the RCP failed to receive a majority vote.  No points " +
					"have been awarded, and the player who proposed this " +
					"rule has lost 10 points.")
			} else if rcp > 319 && len(n) > 1 {
				discord.ChannelMessageSend(
					channel,
					"<@&1001702350571974716>, " +
					"the RCP failed to receive a unanimous vote.  No points " +
					"have been awarded, and the player who proposed this " +
					"rule has lost 10 points.")
			} else {
				discord.ChannelMessageSend(
					channel,
					"<@&1001702350571974716>, " +
					"the RCP has passed the vote. " +
					strconv.Itoa((rcp - 291) * len(y) / len(n)) +
					" points have been awarded to the player who proposed " +
					"this rule.")
			}
			os.WriteFile(
				"nomic/rcp",
				[]byte(strconv.Itoa(rcp + 1) + "\n"),
				0644)
			os.Remove("nomic/poll-msg-id")
			os.Remove("nomic/author")
			players, _ := os.ReadDir("nomic/players")
			os.Symlink(
				"../players/" + players[(rcp - 300) % len(players)].Name(),
				"nomic/author")
		} else {
			discord.ChannelMessageSend(
				channel,
				"<@&1001702350571974716>, " +
				"votes were retracted; resuming the poll.")
		}
	}

}
