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
	"gorm.io/gorm"
)

// HandleJoinEvent グループ参加時の処理
func HandleJoinEvent(ctx context.Context, bot *linebot.Client, groupID string, userID string) linebot.SendingMessage {
	res, err := bot.GetGroupMemberCount(groupID).Do()
	if err != nil {
		log.Printf("Error fetching group members count: %v", err)
		return linebot.NewTextMessage("グループ情報の取得に失敗しました。")
	}

	// 既存のレコードを確認
	var joinGroup models.JoinGroup
	if err := infra.DB.Where("group_id = ?", groupID).First(&joinGroup).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// レコードが存在しない場合、新規作成
			joinGroup = models.JoinGroup{
				ID:        uuid.New(),
				GroupID:   groupID,
				Number:    int64(res.Count),
				IsNowIn:   true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := infra.DB.Create(&joinGroup).Error; err != nil {
				log.Printf("[ERROR] saving group info: %v", err)
			}
		} else {
			// その他のエラー
			log.Printf("[ERROR] fetching group info: %v", err)
			return linebot.NewTextMessage("グループ情報の取得に失敗しました。")
		}
	} else {
		// レコードが存在する場合、更新
		if err := infra.DB.Model(&joinGroup).Updates(models.JoinGroup{
			Number:    int64(res.Count),
			IsNowIn:   true,
			UpdatedAt: time.Now(),
		}).Error; err != nil {
			log.Printf("[ERROR] updating group info: %v", err)
		}
	}

	subject := "New Group Joined"
	cache := make(map[string]string)
	displayName := utils.GetCachedProfileName(bot, "", userID, cache)
	body := fmt.Sprintf("A new group (ID: %s) has added the LINE bot by %s. Members: %d", groupID, displayName, res.Count)
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
