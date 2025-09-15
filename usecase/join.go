package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daiki-trnsk/MoneyLine/constants"
	"github.com/daiki-trnsk/MoneyLine/infra"
	"github.com/daiki-trnsk/MoneyLine/models"
	"github.com/daiki-trnsk/MoneyLine/utils"

	"github.com/google/uuid"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// HandleJoinEvent グループ参加時の処理
func HandleJoinEvent(ctx context.Context, bot *linebot.Client, groupID string) linebot.SendingMessage {
	res, err := bot.GetGroupMemberCount(groupID).Do()
	if err != nil {
		log.Printf("Error fetching group members count: %v", err)
		return linebot.NewTextMessage("グループ情報の取得に失敗しました。")
	}

	joinGroup := models.JoinGroup{
		ID:        uuid.New(),
		GroupID:   groupID,
		Number:    int64(res.Count),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := infra.DB.Create(&joinGroup).Error; err != nil {
		log.Printf("[ERROR] saving group info: %v", err)
	}

	subject := "New Group Joined"
	body := fmt.Sprintf("A new group (ID: %s) has added the LINE bot. Members: %d", groupID, res.Count)
	if err := utils.SendEmail(subject, body); err != nil {
		log.Printf("Error sending notification email: %v", err)
	}

	return linebot.NewTextMessage(constants.JoinMessage)
}

// HandleLeaveEvent グループ退会時の処理
func HandleLeaveEvent(ctx context.Context, groupID string) {
	if err := infra.DB.Model(&models.JoinGroup{}).
		Where("group_id = ?", groupID).
		Update("is_now_in", false).Error; err != nil {
		log.Printf("[ERROR] updating group info: %v", err)
	}
}
