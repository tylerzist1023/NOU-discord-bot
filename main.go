package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
);

const prefix string = "!nou"

type GameState int
const (
    WaitingForPlayerJoin GameState = 0
    GameStarted GameState = 1
);

type Player struct {
    UserID string
    DmChannelID string
}

type GameInstance struct {
    OwnerID string
    Players map[string]Player
    State GameState
    JoinMessageID string
}

func waitForTerminate() {
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
}

func main() {
    secret, err := os.ReadFile("secret.txt")
    if err != nil {
        log.Fatal(err)
    }
    sess, err := discordgo.New(fmt.Sprintf("Bot %s", string(secret)))
    if err != nil {
        log.Fatal(err)
    }

    gameInstances := make(map[string]GameInstance)

    sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
        if m.Author.ID == s.State.User.ID {
            return
        }

        args := strings.Split(m.Content, " ")

        if args[0] != prefix {
            return
        }

        if args[1] == "start" {
            if _, ok := gameInstances[m.Author.ID]; ok {
                content := fmt.Sprintf("<@%s> You have already started an UNO game!", m.Author.ID)
                _, err := s.ChannelMessageSend(m.ChannelID, content)
                if err != nil {
                    log.Fatal(err)
                }
            } else {
                content := fmt.Sprintf("<@%s> has started an UNO game! React to join.", m.Author.ID)
                msg, err := s.ChannelMessageSend(m.ChannelID, content)
                if err != nil {
                    log.Fatal(err)
                }
                err = s.MessageReactionAdd(m.ChannelID, msg.ID, "✅")
                if err != nil {
                    log.Fatal(err)
                }
                gameInstances[m.Author.ID] = GameInstance{OwnerID: m.Author.ID, Players: make(map[string]Player), State: WaitingForPlayerJoin, JoinMessageID: msg.ID}
            }
        } else if args[1] == "begin" {
            if instance, ok := gameInstances[m.Author.ID]; ok {
                if instance.State != WaitingForPlayerJoin {
                    content := fmt.Sprintf("<@%s> You have already begun the UNO game!", m.Author.ID)
                    _, err := s.ChannelMessageSend(m.ChannelID, content)
                    if err != nil {
                        log.Fatal(err)
                    }
                } else {
                    content := fmt.Sprintf("<@%s> The UNO game has begun! Players need to check their DMs!", m.Author.ID)
                    for k,v := range instance.Players {
                        dmChannel, err := s.UserChannelCreate(k)
                        if err != nil {
                            continue
                        }
                        _, err = s.ChannelMessageSend(dmChannel.ID, "You've joined the UNO game.")
                        if err != nil {
                            log.Fatal(err)
                        }
                        v.DmChannelID = dmChannel.ID
                    }

                    instance.State = GameStarted
                    gameInstances[m.Author.ID] = instance

                    _, err := s.ChannelMessageSend(m.ChannelID, content)
                    if err != nil {
                        log.Fatal(err)
                    }
                }
            } else {
                content := fmt.Sprintf("<@%s> You have not started an UNO game!", m.Author.ID)
                _, err := s.ChannelMessageSend(m.ChannelID, content)
                if err != nil {
                    log.Fatal(err)
                }
            }
        } else if args[1] == "stop" {
            if _, ok := gameInstances[m.Author.ID]; ok {
                content := fmt.Sprintf("<@%s> The UNO game has stopped!", m.Author.ID)
                delete(gameInstances, m.Author.ID)
                _, err := s.ChannelMessageSend(m.ChannelID, content)
                if err != nil {
                    log.Fatal(err)
                }
            } else {
                content := fmt.Sprintf("<@%s> You have not started an UNO game!", m.Author.ID)
                _, err := s.ChannelMessageSend(m.ChannelID, content)
                if err != nil {
                    log.Fatal(err)
                }
            }
        }
    })

    sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
        if r.UserID == s.State.User.ID {
            return
        }

        if r.Emoji.APIName() == "✅" {
            for k,v := range gameInstances {
                if v.JoinMessageID == r.MessageID && v.State == WaitingForPlayerJoin {
                    v.Players[r.UserID] = Player{UserID: r.UserID}
                    fmt.Printf("%s joined %s's game\n", r.UserID, k)
                }
            }
        }
    })
    sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
        if r.UserID == s.State.User.ID {
            return
        }

        if r.Emoji.APIName() == "✅" {
            for k,v := range gameInstances {
                if v.JoinMessageID == r.MessageID && v.State == WaitingForPlayerJoin  {
                    delete(v.Players, r.UserID)
                    fmt.Printf("%s left %s's game\n", r.UserID, k)
                }
            }
        }
    })

    sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

    err = sess.Open()
    if err != nil {
        log.Fatal(err)
    }
    defer sess.Close()

    fmt.Println("Bot is online!")

    waitForTerminate()
}