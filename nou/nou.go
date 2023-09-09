package nou

import (
	"fmt"
)

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
    ChannelID string
    JoinMessageID string
}

var gameInstances map[string]GameInstance = make(map[string]GameInstance)

// creates a new game
func Start(ownerID string, channelID string) {
    if _, ok := gameInstances[ownerID]; ok {
        MessageToChannel(ownerID, channelID, "You have already started an UNO game!")
    } else {
        messageID := MessageToChannel(ownerID, channelID, "has started an UNO game! React to join.")
        AddReactionOption(channelID, messageID, "✅", JoinGame, LeaveGame)
        gameInstances[ownerID] = GameInstance{OwnerID: ownerID, Players: make(map[string]Player), State: WaitingForPlayerJoin, ChannelID: channelID, JoinMessageID: messageID}
    }
}

func Begin(ownerID string, defaultChannelID string) {
    if instance, ok := gameInstances[ownerID]; ok {
        if instance.State != WaitingForPlayerJoin {
            MessageToChannel(ownerID, instance.ChannelID, "You have already begun an UNO game!")
        } else {
            MessageToPlayers(gameInstances[ownerID].Players, "You've joined the UNO game.")

            instance.State = GameStarted
            gameInstances[ownerID] = instance

            MessageToChannel(ownerID, instance.ChannelID, "'s UNO game has begun! Players need to check their DMs!")
        }
    } else {
        MessageToChannel(ownerID, defaultChannelID, "You have not started an UNO game!")
    }
}

func Stop(ownerID string, defaultChannelID string) {
    if instance, ok := gameInstances[ownerID]; ok {
        delete(gameInstances, ownerID)
        MessageToChannel(ownerID, instance.ChannelID, "The UNO game has stopped!")
    } else {
        MessageToChannel(ownerID, defaultChannelID, "You have not started an UNO game!")
    }
}

func JoinGame(messageID string, playerID string) {
    for k,v := range gameInstances {
        if v.JoinMessageID == messageID && v.State == WaitingForPlayerJoin  {
            v.Players[playerID] = Player{UserID: playerID}
            fmt.Printf("%s joined %s's game\n", playerID, k)
        }
    }
}

func LeaveGame(messageID string, playerID string) {
    for k,v := range gameInstances {
        if v.JoinMessageID == messageID && v.State == WaitingForPlayerJoin  {
            delete(v.Players, playerID)
            fmt.Printf("%s left %s's game\n", playerID, k)
        }
    }
}