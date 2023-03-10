// drafting.go version 0.0.0

package main

import "bufio"
import "crypto/rand"
import "fmt"
import "math/big"
import "os"
import "os/exec"
import "strings"

import "github.com/bwmarrin/discordgo"

func authorset(discord *discordgo.Session, author string) {
	_, err := os.Stat("nomic/players/" + author)
	if err == nil {
		os.Remove("nomic/author")
		os.Symlink("players/" + author, "nomic/author")
		discord.ChannelMessageSend(os.Getenv("channel"), "author successfully set")
	} else {
		ls := exec.Command("ls", "nomic/players")
		players, _ := ls.Output()
		discord.ChannelMessageSend(os.Getenv("channel"), "Invalid player \"" +
			author + "\".  The following are valid players:\n" + string(players))
	}
}

func mention(usr *discordgo.User) string {
	_, debug := os.LookupEnv("DEBUG")
	if debug {
		return usr.Username
	}
	return usr.Mention()
}

func main() {

	if !(len(os.Args) == 2 && os.Args[1] == "reminder") &&
	   !(len(os.Args) == 3 && os.Args[1] == "msg") {
		fmt.Fprintf(os.Stderr, "usage: %s msg id\n" +
		                       "       %s reminder\n", os.Args[0])
		os.Exit(1)
	}

	discord, err := discordgo.New("Bot " + os.Getenv("key"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "discordgo.New: %s\n", err.Error())
		os.Exit(1)
	}

	buf, err := os.ReadFile("nomic/author")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.ReadFile: %s\n", err.Error())
		os.Exit(1)
	}

	authorid := string(buf[:len(buf) - 1])

	if os.Args[1] == "reminder" {
		buf, err := os.ReadFile("nomic/channels/announcements")
		if err != nil {
			fmt.Fprintf(os.Stderr, "os.ReadFile: %s\n", err.Error())
			os.Exit(1)
		}

		usr, _ := discord.User(authorid)

		var txt string
		found := false
		players, _ := os.ReadDir("nomic/insult")
		for _, player := range players {
			buf, _ := os.ReadFile("nomic/insult/" + player.Name())
			if authorid == string(buf[:len(buf) - 1]) {
				insults, _ := os.ReadDir("insults")
				res, _ := rand.Int(rand.Reader, big.NewInt(int64(len(insults))))
				insult, _ := os.ReadFile(
					"insults/" + insults[int(res.Int64())].Name())
				txt = strings.Replace(string(insult), "@", mention(usr), 1)
				found = true
				break
			}
		}
		if !found {
			txt = mention(usr) + ", this is your 48-hour reminder that it is " +
			"your turn to propose an RCP."
	 	}

		discord.ChannelMessageSend( string(buf[:len(buf) - 1]), txt)

		return
	}

	msg, err := discord.ChannelMessage(os.Getenv("channel"),
									   os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr,
					"discordgo.Session.ChannelMessage: %s",
					err.Error())
		os.Exit(1)
	}

	if msg.Content == "insult optin" {
		players, _ := os.ReadDir("nomic/insult")
		found := false
		for _, player := range players {
			buf, _ := os.ReadFile("nomic/insult/" + player.Name())
			if os.Getenv("player") == string(buf[:len(buf) - 1]) {
				found = true
				break
			}
		}
		if found {
			discord.ChannelMessageSend(
				os.Getenv("channel"),
				"You are already on the insult list.")
		} else {
			players, _ := os.ReadDir("nomic/players")
			found = false
			for _, player := range players {
				buf, _ := os.ReadFile("nomic/players/" + player.Name())
				if os.Getenv("player") == string(buf[:len(buf) - 1]) {
					found = true
					os.Symlink(
						"../players/" + player.Name(),
						"nomic/insult/" + player.Name())
					break
				}
			}
			if found {
				discord.ChannelMessageSend(
					os.Getenv("channel"),
					"You have been added to the insult list, you absolute " +
					"buffoon.")
			}
		}
		return
	}

	if msg.Content == "insult optout" {
		players, _ := os.ReadDir("nomic/insult")
		found := false
		for _, player := range players {
			buf, _ := os.ReadFile("nomic/insult/" + player.Name())
			if os.Getenv("player") == string(buf[:len(buf) - 1]) {
				found = true
				os.Remove("nomic/insult/" + player.Name())
				discord.ChannelMessageSend(
					os.Getenv("channel"),
					"You have been removed from the insult list.")
				break
			}
		}
		if !found {
			discord.ChannelMessageSend(
				os.Getenv("channel"),
				"You are not currently on the insult list.  If you are still " +
				"receiving insults, please contact an administrator.")
		}
		return
	}

	isauthor := os.Getenv("player") == authorid

	admin, err := os.ReadDir("nomic/admin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.ReadDir: %s\n", err.Error())
		os.Exit(1)
	}

	isadmin := false
	var adminname string
	for i := 0; i < len(admin); i ++ {
		if !admin[i].IsDir() {
			buf, _ := os.ReadFile("nomic/admin/" + admin[i].Name())
			if string(buf[:len(buf) - 1]) == os.Getenv("player") {
				isadmin = true
				adminname = admin[i].Name()
				break
			}
		}
	}

	if isauthor || isadmin {

		file, err := os.OpenFile("nomic/draft", os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			if isauthor {
				if msg.Content == "rcp end" {
					file.Close()
					os.Rename("nomic/draft", "nomic/rcp-messages")
					os.WriteFile("nomic/channels/author",
					             []byte(os.Getenv("channel") + "\n"),
					             0644)
				} else {
					file.Write([]byte(msg.ID + "\n"))
					file.Close()
				}
			} else {
				file.Close()
				usr, err := discord.User(authorid)
				if err != nil {
					fmt.Fprintf(os.Stderr,
					            "discordgo.Session.User: %s\n",
				                err.Error())
					os.Exit(1)
				}
				discord.ChannelMessageSend(os.Getenv("channel"),
				                           "The RCP is currently " +
				                           "being drafted by " + usr.Username +
				                           ". Contact an " +
				                           "administrator for help if" +
				                           " you think this is an " +
				                           "error.")
			}
		} else {
			file, err := os.OpenFile("nomic/drafttitle", os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				if isauthor {
					file.Write([]byte(msg.Content + "\n"))
					file.Close()
					os.Rename("nomic/drafttitle", "nomic/rcp-title")
					os.Create("nomic/draft")
					discord.ChannelMessageSend(
						os.Getenv("channel"),
						"Go ahead and send the RCP.  Type \"rcp end\" to end.")
				} else {
					file.Close()
					usr, err := discord.User(authorid)
					if err != nil {
						fmt.Fprintf(
							os.Stderr,
							"discordgo.Session.User: %s\n",
							err.Error())
						os.Exit(1)
					}
					discord.ChannelMessageSend(
						os.Getenv("channel"),
						"The RCP is currently being drafted by " +
						usr.Username + ". Contact an administrator for help " +
						"if you think this is an error.")
				}
			} else {
				cmd := strings.Fields(msg.Content)
				switch cmd[0] {
					case "author":
						authorset(discord, cmd[1])
					case "rcp":
						switch cmd[1] {
							case "create":
								if !isauthor {
									os.Remove("nomic/author")
									os.Symlink("players/" + adminname, "nomic/author")
								}
								os.Create("nomic/drafttitle")
								discord.ChannelMessageSend(os.Getenv("channel"),
														   "What is the RCP title?")

							case "send":
								msgfile, err := os.Open("nomic/rcp-messages")
								if err == nil {
									rcp, _ := os.ReadFile("nomic/rcp")
									title, _ := os.ReadFile("nomic/rcp-title")
									channel, _ := os.ReadFile("nomic/channels/author")
									rcpchan, _ := os.ReadFile("nomic/channels/rcp")
									thread, _ := discord.ThreadStart(
										string(rcpchan[:len(rcpchan) - 1]),
										"RCP " + string(rcp[:len(rcp) - 1]) + " - " +
											string(title[:len(title) - 1]),
										11,
										7 * 24 * 60)

									os.WriteFile("nomic/poll-thread-id",
									             []byte(thread.ID + "\n"),
									             0644)

									txt := "**Official Vote Thread**\n" +
										   "RCP " + string(rcp[:len(rcp) - 1]) + " -" +
											" " + string(title[:len(title) - 1]) +
											"\n"
									scanner := bufio.NewScanner(msgfile)
									for scanner.Scan() {
										msg, _ := discord.ChannelMessage(
											string(channel[:len(channel) - 1]),
											scanner.Text())
										txt += msg.Content + "\n"
									}
									_, debug := os.LookupEnv("DEBUG")
									var legislators string
									if debug {
										legislators = "@Eldritch Legislators"
									} else {
										legislators = "<@&1001702350571974716>"
									}
									txt += "----------------------------------------" +
										   "\n" + legislators + "\n" +
										   "Please react to this message with ✅ " +
										   "for a \"yes\" vote or ❌ for a \"no\"" +
										   " vote."

									var messages []string
									var msg *discordgo.Message

									if len(txt) >= 2000 {
										for len(txt) >= 1994 {
											var i int
											for i = 1992; i > 1; i -- {
												if txt[i] == '\n' {
													break
												}
											}
											if i == 1 {
												for i = 1992; i > 1; i -- {
													if txt[i] == ' ' {
														break
													}
												}
												if i == 1 {
													i = 1992
												}
											}
											messages = append(messages, txt[0:i])
											txt = txt[i:len(txt)]
										}
										for i := 0; i < len(messages); i ++ {
											discord.ChannelMessageSend(
												thread.ID,
												fmt.Sprintf(
													"[%d/%d]\n%s",
													i + 1,
													len(messages) + 1,
													messages[i]))
										}
										msg, _ = discord.ChannelMessageSend(
											thread.ID,
											fmt.Sprintf("[%d/%d]\n%s",
												len(messages) + 1,
												len(messages) + 1,
												txt))
									} else {
										msg, _ = discord.ChannelMessageSend(
											thread.ID,
											txt)
									}

									os.WriteFile("nomic/poll-msg-id",
												 []byte(msg.ID + "\n"),
									             0644)

									os.WriteFile("nomic/phase", 
									             []byte("rcp\n"),
									             0644)
								} else {
									discord.ChannelMessageSend(
										os.Getenv("channel"),
										"No RCP has been written!  Compose " +
										"an RCP with \"rcp create\".")
								}
						}
					case "help":
						discord.ChannelMessageSend(
							os.Getenv("channel"),
							"available commands:\n" +
							"    author lastname : set the rcp author\n" +
							"    rcp create | send : manage rcps")
					default:
						discord.ChannelMessageSend(
							os.Getenv("channel"),
							"Invalid command.  Type \"help\" for a list of " +
							"available commands.")
				}
			}
		}
	} else {
		discord.ChannelMessageSend(os.Getenv("channel"),
		                           "You are not authorized to manage " +
		                           "RCPs. Contact an administrator " +
		                           "for help if you think this is an " +
		                           "error.")
	}
}
