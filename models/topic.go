package models

import "time"

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

type TopicRead struct {
	BaseModel

	TopicID        uint         `gorm:"column:topic_id"`
	Topic          Topic        `gorm:"foreignkey:TopicID"`
	UserID         uint         `gorm:"column:user_id"`
	User           User         `gorm:"foreignkey:UserID"`
	TopicMessageID uint         `gorm:"column:topic_messages_id"`
	LastMessage    TopicMessage `gorm:"foreignkey:TopicMessageID"`
}

func (TopicRead) TableName() string {
	return "topics_topicread"
}

type TopicMessageImage struct {
	BaseModel

	ImageID        uint         `gorm:"column:image_id"`
	Image          Image        `gorm:"foreignkey:ImageID"`
	TopicMessageID uint         `gorm:"column:topic_messages_id"`
	TopicMessage   TopicMessage `gorm:"foreignkey:TopicMessageID"`
}

func (TopicMessageImage) TableName() string {
	return "topics_topicread"
}

type Image struct {
	BaseModel

	UserID uint `gorm:"user_id"`
	User   User `gorm:"foreignkey:UserID"`
}
