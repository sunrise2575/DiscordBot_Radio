package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/seehuhn/mt19937"
	"github.com/tidwall/gjson"
)

type DiscordCommand struct {
	Description string
	Callback    func(arg []string, sess *discordgo.Session, msg *discordgo.Message)
}

type DiscordCommands map[string]DiscordCommand

func makeCommands() DiscordCommands {
	// 커맨드 만든다
	result := make(DiscordCommands)

	result["rand"] = DiscordCommand{
		Description: "주사위를 굴린다. `rand` == [1,6], `rand <n>` == [1,n]",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())
			target := 6

			if len(arg) >= 2 {
				if _target, e := strconv.ParseInt(arg[1], 10, 64); e != nil {
					log.Println(e)
					return
				} else {
					target = int(_target)
				}
			} else {
				arg = append(arg, strconv.Itoa(target))
			}

			if target <= 1 {
				return
			}

			if target == 2 {
				str := ""
				if rng.Intn(target) == 0 {
					str = "앞면"
				} else {
					str = "뒷면"
				}
				sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("동전을 던졌다:", "`"+str+"`"))
			} else {
				sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("`"+arg[1]+"`면체 주사위를 굴렸다:", fmt.Sprint("`", rng.Intn(target)+1, "`")))
			}
		},
	}

	result["anime"] = DiscordCommand{
		Description: "애니추천좀",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			files1, e := ioutil.ReadDir("/TANK/Anime")
			if e != nil {
				log.Println(e)
				return
			}
			files2, e := ioutil.ReadDir("/TANK/.workshop/EncReq-Anime")
			if e != nil {
				log.Println(e)
				return
			}
			files3, e := ioutil.ReadDir("/TANK/.workshop/SubReq-Anime")
			if e != nil {
				log.Println(e)
				return
			}

			files := append(files1, files2...)
			files = append(files, files3...)

			target := files[rng.Intn(len(files))].Name()
			sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("애니추천:", "`"+target+"`"))
		},
	}

	result["movie"] = DiscordCommand{
		Description: "영화추천좀",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			files1, e := ioutil.ReadDir("/TANK/Movie")
			if e != nil {
				log.Println(e)
				return
			}
			files2, e := ioutil.ReadDir("/TANK/.workshop/EncReq-Movie")
			if e != nil {
				log.Println(e)
				return
			}
			files3, e := ioutil.ReadDir("/TANK/.workshop/SubReq-Movie")
			if e != nil {
				log.Println(e)
				return
			}

			files := append(files1, files2...)
			files = append(files, files3...)

			target := files[rng.Intn(len(files))].Name()
			sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("영화추천:", "`"+target+"`"))
		},
	}

	result["music"] = DiscordCommand{
		Description: "음악추천좀",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			files := []string{}

			if e := filepath.Walk("/TANK/Music",
				func(path string, info os.FileInfo, e error) error {
					if e != nil {
						return e
					}
					if info.IsDir() {
						return nil
					}
					files = append(files, path)
					return nil
				}); e != nil {
				log.Println(e)
				return
			}
			path := files[rng.Intn(len(files))]
			base := filepath.Base(path)
			target := strings.TrimSuffix(base, filepath.Ext(path))
			file, e := os.OpenFile(path, os.O_RDONLY, 0644)
			if e != nil {
				log.Println(e)
				return
			}
			sess.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
				Content: fmt.Sprintln("음악추천:", "`"+target+"`"),
				Files: []*discordgo.File{
					{
						Name:        base,
						ContentType: "audio/ogg",
						Reader:      file,
					},
				},
			})
		},
	}

	result["pick"] = DiscordCommand{
		Description: "당첨시킨다.",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			members, e := sess.GuildMembers(msg.GuildID, "", 1000)
			if e != nil {
				log.Println(e)
				return
			}

			for {
				target := members[rng.Intn(len(members))]
				if target.User.ID != sess.State.User.ID {
					if len(target.Nick) > 0 {
						sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("여기 멤버 중 당첨자:", "`"+target.Nick+"`", "(`"+target.User.Username+"`)"))
					} else {
						sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("여기 멤버 중 당첨자:", "`"+target.User.Username+"`"))
					}
					break
				}
			}
		},
	}

	result["couple"] = DiscordCommand{
		Description: "짝짓기 한다.",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			members, e := sess.GuildMembers(msg.GuildID, "", 1000)
			if e != nil {
				log.Println(e)
				return
			}

			for {
				rng := rand.New(mt19937.New())
				rng.Seed(time.Now().UnixNano())

				target := [2]*discordgo.Member{}
				target[0], target[1] = members[rng.Intn(len(members))], members[rng.Intn(len(members))]

				if target[0].User.ID != sess.State.User.ID && target[1].User.ID != sess.State.User.ID {
					name := [2]string{}
					for i := range name {
						if len(target[i].Nick) > 0 {
							name[i] = "`" + target[i].Nick + "` (`" + target[i].User.Username + "`)"
						} else {
							name[i] = "`" + target[i].User.Username + "`"
						}
					}

					acts := gjson.Parse(readFileAsString("./txt/couple.json")).Get("content").Array()
					idx := rng.Intn(len(acts))
					activity := acts[idx].Get("activity").String()
					message := acts[idx].Get("message").String()

					sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln(name[0], activity, name[1], " -- ", message))

					break
				}
			}
		},
	}

	result["rank"] = DiscordCommand{
		Description: "랭킹을 본다. `rank` == 5명 랭킹, `rank <n>` == n명 랭킹",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			members, e := sess.GuildMembers(msg.GuildID, "", 1000)
			if e != nil {
				log.Println(e)
				return
			}

			rng.Shuffle(len(members), func(i, j int) {
				members[i], members[j] = members[j], members[i]
			})

			maxCount := 5

			if len(arg) >= 2 {
				if _target, e := strconv.ParseInt(arg[1], 10, 64); e != nil {
					log.Println(e)
					return
				} else {
					maxCount = int(_target)
				}
			}

			if maxCount > len(members)-1 {
				maxCount = len(members) - 1
			}

			result := ">>> "
			for i, count := 0, 0; count < maxCount && i < len(members); i++ {
				target := members[i]

				if target.User.ID == sess.State.User.ID {
					continue
				}

				result += strconv.Itoa(count+1) + "위: "

				if len(target.Nick) > 0 {
					result += "`" + target.Nick + "` (`" + target.User.Username + "`)"
				} else {
					result += "`" + target.User.Username + "`"
				}

				result += "\n"

				count++
			}
			sess.ChannelMessageSend(msg.ChannelID, result)
		},
	}

	result["member"] = DiscordCommand{
		Description: "멤버를 본다",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			members, e := sess.GuildMembers(msg.GuildID, "", 1000)
			if e != nil {
				log.Println(e)
				return
			}

			sort.Slice(members, func(i, j int) bool {
				iName, jName := "", ""
				if len(members[i].Nick) > 0 {
					iName = members[i].Nick
				} else {
					iName = members[i].User.Username
				}

				if len(members[j].Nick) > 0 {
					jName = members[j].Nick
				} else {
					jName = members[j].User.Username
				}

				return iName < jName
			})

			result := "```"
			for i := 0; i < len(members); i++ {
				target := members[i]

				if target.User.ID == sess.State.User.ID {
					continue
				}

				if len(target.Nick) > 0 {
					result += target.Nick
				} else {
					result += target.User.Username
				}

				result += "\n"
			}

			result += "```"
			sess.ChannelMessageSend(msg.ChannelID, result)
		},
	}

	result["food"] = DiscordCommand{
		Description: "오늘 뭐 먹지",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			rng := rand.New(mt19937.New())
			rng.Seed(time.Now().UnixNano())

			str := readFileAsString("./txt/food_list.txt")
			list := strings.FieldsFunc(str, func(in rune) bool {
				return in == '\n'
			})

			target := list[rng.Intn(len(list))]
			sess.ChannelMessageSend(msg.ChannelID, fmt.Sprintln("이거나 먹어라:", "`"+target+"`"))
		},
	}

	result["?"] = DiscordCommand{
		Description: "해법 출력",
		Callback: func(arg []string, sess *discordgo.Session, msg *discordgo.Message) {
			temp := [][2]string{}
			for name, cmd := range result {
				temp = append(temp, [2]string{name, cmd.Description})
			}
			sort.Slice(temp, func(i, j int) bool {
				return temp[i][0] < temp[j][0]
			})
			content := ">>> "
			for _, cmd := range temp {
				content += "`" + cmd[0] + "`"
				content += " : "
				content += cmd[1]
				content += "\n"
			}
			sess.ChannelMessageSend(msg.ChannelID, content)
		},
	}

	return result
}
