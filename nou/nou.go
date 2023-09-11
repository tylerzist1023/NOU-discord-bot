package nou

import (
    "fmt"
    "math/rand"
    "strconv"
    "strings"
    "time"
)

type GameState int
const (
    GameStarted GameState = 0
    GameBegun GameState = 1
    GameFinished GameState = 2
);

type CardColor int
const (
    CardColorYellow CardColor = 0
    CardColorRed CardColor = 1
    CardColorBlue CardColor = 2
    CardColorGreen CardColor = 3
);

type CardValue int
const (
    CardValue0 CardValue = 0
    CardValue1 CardValue = 1
    CardValue2 CardValue = 2
    CardValue3 CardValue = 3
    CardValue4 CardValue = 4
    CardValue5 CardValue = 5
    CardValue6 CardValue = 6
    CardValue7 CardValue = 7
    CardValue8 CardValue = 8
    CardValue9 CardValue = 9
    CardValuePlus2 CardValue = 10
    CardValueSkip CardValue = 11
    CardValueReverse CardValue = 12
    CardColorWild CardValue = 13
    CardColorPlus4 CardValue = 14
);

type Card struct {
    color CardColor
    value CardValue
}

type Player struct {
    UserID string
    DmChannelID string
    hand []Card
    awake bool
}

type GameInstance struct {
    OwnerID string
    Players map[string]Player
    State GameState
    ChannelID string
    JoinMessageID string
    TurnOrder []string
    TurnOrderStep int
    CurrentTurnIndex int
    StackTop Card
}

var gameInstances map[string]GameInstance = make(map[string]GameInstance)

// creates a new game
func Start(ownerID string, channelID string) {
    if _, ok := gameInstances[ownerID]; ok {
        MessageToChannel(ownerID, channelID, "You have already started an UNO game!")
    } else {
        messageID := MessageToChannel(ownerID, channelID, "has started an UNO game! React to join.")
        AddReactionOption(channelID, messageID, "✅", JoinGame, LeaveGame)
        gameInstances[ownerID] = GameInstance{OwnerID: ownerID, Players: make(map[string]Player), State: GameStarted, ChannelID: channelID, JoinMessageID: messageID}
    }
}

func Begin(ownerID string, defaultChannelID string) {
    if instance, ok := gameInstances[ownerID]; ok {
        if instance.State == GameBegun {
            MessageToChannel(ownerID, instance.ChannelID, "You have already begun an UNO game!")
        } else {
            for k,v := range gameInstances[ownerID].Players {
                v.hand = dealHand()
                v = MessageToPlayer(v, handToString(v.hand))
                gameInstances[ownerID].Players[k] = v
            }

            instance.State = GameBegun
            instance.StackTop = randomCard()
            instance.decideTurnOrder()
            instance.providePlayableCardsMenuToCurrentPlayer()
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
        if v.JoinMessageID == messageID && v.State == GameStarted  {
            v.Players[playerID] = Player{UserID: playerID}
            fmt.Printf("%s joined %s's game\n", playerID, k)
        }
    }
}

func LeaveGame(messageID string, playerID string) {
    for k,v := range gameInstances {
        if v.JoinMessageID == messageID && v.State == GameStarted  {
            v.State = GameFinished
            gameInstances[k] = v
            delete(gameInstances[k].Players, playerID)
            fmt.Printf("%s left %s's game\n", playerID, k)
        }
    }
}

func randomCard() Card {
    return Card{
        color: CardColor(rand.Intn(3+1)),
        value: CardValue(rand.Intn(14+1)),
    }
}

func (c Card) ToString() string {
    var sb strings.Builder

    if c.value == CardColorWild {
        sb.WriteString("Wild")
        return sb.String()
    } else if c.value == CardColorPlus4 {
        sb.WriteString("+4")
        return sb.String()
    }

    switch color := c.color; color {
    case CardColorYellow:   sb.WriteString("Yellow")
    case CardColorRed:      sb.WriteString("Red")
    case CardColorBlue:     sb.WriteString("Blue")
    case CardColorGreen:    sb.WriteString("Green")
    }
    sb.WriteString(" ")
    switch value := c.value; value {
    case CardValue0:        fallthrough
    case CardValue1:        fallthrough
    case CardValue2:        fallthrough
    case CardValue3:        fallthrough
    case CardValue4:        fallthrough
    case CardValue5:        fallthrough
    case CardValue6:        fallthrough
    case CardValue7:        fallthrough
    case CardValue8:        fallthrough
    case CardValue9:        sb.WriteString(strconv.Itoa(int(c.value)))
    case CardValuePlus2:    sb.WriteString("+2")
    case CardValueSkip:     sb.WriteString("Skip")
    case CardValueReverse:  sb.WriteString("Reverse")
    }
    return sb.String()
}

func handToString(hand []Card) string {
    var sb strings.Builder
    sb.WriteString("Your hand: ")
    for _,v := range hand {
        sb.WriteString(v.ToString())
        sb.WriteString(", ")
    }
    return sb.String()
}

func dealHand() []Card {
    rand.Seed(time.Now().UnixNano())
    hand := make([]Card, 0, 7)
    for i := 0; i < 7; i++ {
        hand = append(hand, randomCard())
    }
    return hand
}

func (g *GameInstance) decideTurnOrder() {
    rand.Seed(time.Now().UnixNano())
    playerIDs := make([]string, len(g.Players))
    i := 0
    for k := range g.Players {
        playerIDs[i] = k
    }

    rand.Shuffle(len(playerIDs), func(i int, j int) {
        playerIDs[i], playerIDs[j] = playerIDs[j], playerIDs[i]
    })

    g.TurnOrder = playerIDs
    g.TurnOrderStep = 1
    g.CurrentTurnIndex = 0
}

func (g *GameInstance) reverseTurnOrder() int {
    g.TurnOrderStep *= -1
    return g.TurnOrderStep
}

func (g *GameInstance) advanceTurn() {
    g.CurrentTurnIndex += g.TurnOrderStep
    for g.CurrentTurnIndex >= len(g.TurnOrder) {
        g.CurrentTurnIndex -= len(g.TurnOrder)
    }
    for g.CurrentTurnIndex < 0 {
        g.CurrentTurnIndex += len(g.TurnOrder)
    }
}

func cardDoesMatch(player Card, stackTop Card) bool {
    if player.color == stackTop.color {
        return true
    } else if player.value == stackTop.value {
        return true
    } else if player.value == CardColorWild || player.value == CardColorPlus4 {
        return true
    } else if stackTop.value == CardColorWild || stackTop.value == CardColorPlus4 {
        return true
    }
    return false
}

func (g *GameInstance) providePlayableCardsMenuToCurrentPlayer() {
    player := g.Players[g.TurnOrder[g.CurrentTurnIndex]]

    playableCardIndicies := make([]int, 0, len(player.hand))

    for i,_ := range player.hand {
        if cardDoesMatch(player.hand[i],g.StackTop) {
            playableCardIndicies = append(playableCardIndicies, i)
        }
    }

    var message = "Cards you can play:\n"
    for i,v := range playableCardIndicies {
        message += strconv.Itoa(i+1) + ". " + player.hand[v].ToString() + "\n"
    }
    message += "Top of the Stack: " + g.StackTop.ToString()

    player = MessageToPlayer(player, message)
    g.Players[player.UserID] = player
}