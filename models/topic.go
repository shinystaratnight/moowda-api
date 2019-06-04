package models

type BaseModel struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	//CreatedAt time.Time `json:"-"`
	//UpdatedAt time.Time `json:"-"`
	//DeletedAt *time.Time `sql:"index" json:"-"`
}

type Topic struct {
	BaseModel

	Title       string `gorm:"title" json:"title"`
	OwnerID     uint   `gorm:"owner_id" json:"-"`
	Owner       User   `gorm:"foreignkey:OwnerID" json:"-"`
	UnreadCount uint   `gorm:"-" json:"unread_count"`
}

func (Topic) TableName() string {
	return "topics_topic"
}

type TopicCard struct {
	BaseModel

	Title         string `gorm:"title" json:"title"`
	MessagesCount uint   `gorm:"-" json:"messages_count"`
}

func (TopicCard) TableName() string {
	return "topics_topic"
}

type TopicMessage struct {
	BaseModel

	TopicID uint  `gorm:"topic_id"`
	Topic   Topic `gorm:"foreignkey:TopicID"`
	UserID  uint  `gorm:"user_id"`
	User    User  `gorm:"foreignkey:UserID"`
	Images  []TopicMessageImage
	Content string `gorm:"content"`
}

func (TopicMessage) TableName() string {
	return "topics_topicmessage"
}

type TopicRead struct {
	BaseModel

	TopicID        uint         `gorm:"topic_id"`
	Topic          Topic        `gorm:"foreignkey:TopicID"`
	UserID         uint         `gorm:"user_id"`
	User           User         `gorm:"foreignkey:UserID"`
	TopicMessageID uint         `gorm:"topic_messages_id"`
	LastMessage    TopicMessage `gorm:"foreignkey:TopicMessageID"`
}

func (TopicRead) TableName() string {
	return "topics_topicread"
}

type TopicMessageImage struct {
	BaseModel

	ImageID        uint         `gorm:"image_id"`
	Image          Image        `gorm:"foreignkey:ImageID"`
	TopicMessageID uint         `gorm:"topic_messages_id"`
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
