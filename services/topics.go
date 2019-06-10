package services

import (
	"github.com/jinzhu/gorm"
	"moowda/models"
)

type TopicService struct {
	db *gorm.DB
}

func NewTopicService(db *gorm.DB) *TopicService {
	return &TopicService{db: db}
}

func (s *TopicService) GetTopicCardForUser(topic *models.Topic, user *models.User) (*models.TopicCard, error) {
	var newTopic models.TopicCard

	var query *gorm.DB
	if user != nil {
		query = s.db.Select("id, title, (?) as unread_messages_count, (?) as messages_count",
			s.db.Table("topics_topicmessage").
				Select("COUNT(*)").
				Where("topics_topicmessage.topic_id = topics_topic.id and topics_topicmessage.user_id <> ? and topics_topicmessage.id > (select coalesce((?), 0))", user.ID,
					s.db.Table("topics_topicmessageread").Select("coalesce(id, 0)").Where("topics_topicmessageread.topic_id = ?", topic.ID).Order("id desc").Limit(1).QueryExpr(),
				).QueryExpr(),
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr(),
		).Where("id = ?", topic.ID).Find(&newTopic)
	} else {
		query = s.db.Where("id = ?", topic.ID).Select("id, title, (?) as unread_messages_count, (?) as messages_count",
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = ?", topic.ID).QueryExpr(),
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = ?", topic.ID).QueryExpr(),
		).Find(&newTopic)
	}
	if err := query.Error; err != nil {
		return nil, err
	}

	return &newTopic, nil
}
