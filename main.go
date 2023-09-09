package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"

    nou "github.com/tylerzist1023/NOU-discord-bot/nou"
);

const prefix string = "/nou"

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
    nou.SetSession(sess)

    sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
        if m.Author.ID == s.State.User.ID {
            return
        }

        args := strings.Split(m.Content, " ")

        if args[0] != prefix {
            return
        }

        if args[1] == "start" {
            nou.Start(m.Author.ID, m.ChannelID)
        } else if args[1] == "begin" {
            nou.Begin(m.Author.ID, m.ChannelID)
        } else if args[1] == "stop" {
            nou.Stop(m.Author.ID, m.ChannelID)
        }
    })

    sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
        if r.UserID == s.State.User.ID {
            return
        }

        nou.ReactionCallbacks[r.ChannelID][r.MessageID][r.Emoji.APIName()].Add(r.MessageID, r.UserID)
    })
    sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
        if r.UserID == s.State.User.ID {
            return
        }

        nou.ReactionCallbacks[r.ChannelID][r.MessageID][r.Emoji.APIName()].Remove(r.MessageID, r.UserID)
    })

    sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

    err = sess.Open()
    if err != nil {
        log.Fatal(err)
    }
    defer sess.Close()

    fmt.Println("Bot is online!")

    waitForTerminate()

    fmt.Println("Bot is offline!")
}