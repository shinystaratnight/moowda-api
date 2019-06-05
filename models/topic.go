package models

import (
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt time.Time `json:"-"`
	//DeletedAt *time.Time `sql:"index" json:"-"`
}

type Topic struct {
	BaseModel

	Title       string `gorm:"column:title" json:"title"`
	OwnerID     uint   `gorm:"column:owner_id" json:"-"`
	Owner       User   `gorm:"foreignkey:OwnerID" json:"-"`
	UnreadCount uint   `gorm:"-" json:"unread_count"`
}

func (Topic) TableName() string {
	return "topics_topic"
}

type TopicCard struct {
	BaseModel

	Title         string `gorm:"column:title" json:"title"`
	MessagesCount uint   `gorm:"-" json:"messages_count"`
}

func (TopicCard) TableName() string {
	return "topics_topic"
}

type TopicDetail struct {
	BaseModel

	Title           string    `gorm:"column:title" json:"title"`
	MessagesCount   uint      `gorm:"-" json:"messages_count"`
	LastMessageDate time.Time `gorm:"-" json:"last_message_date"`
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
}

func (Image) TableName() string {
	return "topics_image"
}
