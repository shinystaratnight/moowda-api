package models

import (
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"math"
	"moowda/app"
	"strings"
	"time"
)

const (
	MaxChatImageWidth = 500.00
)

type BaseModel struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Topic struct {
	BaseModel

	Title       string `gorm:"column:title;unique_index" json:"title" conform:"trim"`
	OwnerID     uint   `gorm:"column:owner_id" json:"-"`
	Owner       User   `gorm:"foreignkey:OwnerID" json:"-"`
	UnreadCount uint   `gorm:"-" json:"unread_count"`
}

func (Topic) TableName() string {
	return "topics_topic"
}

func (t Topic) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Title, validation.Required),
	)
}

type TopicCard struct {
	BaseModel

	Title               string `gorm:"column:title" json:"title"`
	UnreadMessagesCount uint   `gorm:"-" json:"unread_messages_count"`
	MessagesCount       uint   `gorm:"-" json:"messages_count"`
}

func (TopicCard) TableName() string {
	return "topics_topic"
}

type TopicDetail struct {
	BaseModel

	Title         string `gorm:"column:title" json:"title"`
	MessagesCount uint   `gorm:"-" json:"messages_count"`
}

func (TopicDetail) TableName() string {
	return "topics_topic"
}

type TopicMessage struct {
	BaseModel

	TopicID uint                `gorm:"column:topic_id" json:"-"`
	Topic   Topic               `gorm:"column:foreignkey:TopicID" json:"-"`
	UserID  uint                `gorm:"column:user_id" json:"-"`
	User    User                `gorm:"foreignkey:UserID" json:"user"`
	Images  []TopicMessageImage `json:"images"`
	Content string              `gorm:"column:content" json:"content"`
}

func (TopicMessage) TableName() string {
	return "topics_topicmessage"
}

type TopicMessageRead struct {
	BaseModel

	TopicID        uint         `gorm:"column:topic_id"`
	Topic          Topic        `gorm:"foreignkey:TopicID"`
	UserID         uint         `gorm:"column:user_id"`
	User           User         `gorm:"foreignkey:UserID"`
	TopicMessageID uint         `gorm:"column:message_id"`
	LastMessage    TopicMessage `gorm:"foreignkey:TopicMessageID"`
}

func (TopicMessageRead) TableName() string {
	return "topics_topicmessageread"
}

type TopicMessageImage struct {
	ImageID        uint         `gorm:"column:image_id" json:"-"`
	Image          Image        `gorm:"foreignkey:ImageID" json:"image"`
	TopicMessageID uint         `gorm:"column:topicmessage_id" json:"-"`
	TopicMessage   TopicMessage `gorm:"foreignkey:TopicMessageID" json:"-"`
}

func (TopicMessageImage) TableName() string {
	return "topics_topicmessage_images"
}

type Image struct {
	BaseModel

	UserID uint   `gorm:"column:user_id" json:"-"`
	User   User   `gorm:"foreignkey:UserID"  json:"-"`
	URL    string `gorm:"column:url" json:"url"`
	Height int    `gorm:"column:height" json:"height"`
	Width  int    `gorm:"column:width" json:"width"`
}

func (Image) TableName() string {
	return "topics_image"
}

func (i *Image) MarshalJSON() ([]byte, error) {
	width := int(math.Min(MaxChatImageWidth, float64(i.Width)))
	height := int(float64(i.Height) * float64(float64(width)/float64(i.Width)))

	return json.Marshal(&struct {
		ID     uint   `json:"id"`
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	}{
		ID:     i.ID,
		URL:    i.GetImageURL(),
		Width:  width,
		Height: height,
	})
}

func (i Image) GetImageURL() string {
	if strings.HasPrefix("https://", i.URL) {
		return i.URL
	}
	return fmt.Sprintf("%s/%s", app.Config.BaseURL, i.URL)
}

type CreateTopicMessageRequest struct {
	Content string `json:"content"`
	Images  []uint `json:"images"`
}

func (r CreateTopicMessageRequest) Validate() error {
	if len(r.Images) == 0 {
		return validation.ValidateStruct(&r,
			validation.Field(&r.Content, validation.Required),
		)
	}

	return nil
}
