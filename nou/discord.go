package nou

import (
	"log"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

type AddRemoveCallbackFuncs struct {
	Add func(string, string)
	Remove func(string, string)
}

var ReactionCallbacks = make(map[string]map[string]map[string]AddRemoveCallbackFuncs)

func SetSession(session_ *discordgo.Session) {
	session = session_
}

func MessageToChannel(mentionID string, channelID string, message string) string {
	messageWithMention := fmt.Sprintf("<@%s> %s", mentionID, message)
	msg,err := session.ChannelMessageSend(channelID, messageWithMention)
	if err != nil {
		log.Println(err)
		return ""
	}
	fmt.Println(messageWithMention)
	return msg.ID
}

func MessageToPlayers(players map[string]Player, message string) {
	for k,_ := range players {
		players[k] = MessageToPlayer(players[k], message)
	}
}

func MessageToPlayer(player Player, message string) Player {
	dmChannel, err := session.UserChannelCreate(player.UserID)
    if err != nil {
        log.Println(err)
        return player
    }
    MessageToChannel(player.UserID, dmChannel.ID, message)
    player.DmChannelID = dmChannel.ID
    return player
}

func AddReactionOption(channelID string, messageID string, emoji string, addCallback func(string, string), removeCallback func(string, string)) {
	err := session.MessageReactionAdd(channelID, messageID, emoji)
    if err != nil {
        log.Println(err)
        return
    }
    var funcs AddRemoveCallbackFuncs = AddRemoveCallbackFuncs{
    	Add: addCallback,
    	Remove: removeCallback,
    }

    if _,ok := ReactionCallbacks[channelID]; !ok {
    	ReactionCallbacks[channelID] = make(map[string]map[string]AddRemoveCallbackFuncs)
    }
    if _,ok := ReactionCallbacks[channelID][messageID]; !ok {
    	ReactionCallbacks[channelID][messageID] = make(map[string]AddRemoveCallbackFuncs)
    }
    ReactionCallbacks[channelID][messageID][emoji] = funcs
}