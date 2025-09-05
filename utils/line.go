package utils

import (
	"log"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func GetCachedProfileName(bot *linebot.Client, groupID, userID string, cache map[string]string) string {
	if name, exists := cache[userID]; exists {
		return name
	}
	profile, err := bot.GetGroupMemberProfile(groupID, userID).Do()
	if err != nil {
		log.Println(err, "Failed to get profile for user: "+userID)
		return "@不明"
	}
	name := safeName(profile)
	cache[userID] = name
	return name
}

func safeName(p *linebot.UserProfileResponse) string {
	if p == nil {
		return "@不明"
	}
	return "@" + p.DisplayName
}