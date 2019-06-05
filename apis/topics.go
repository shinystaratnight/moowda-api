package apis

import (
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"moowda/models"
	"net/http"
	"strconv"
)

type TopicAPI struct {
	db *gorm.DB
}

func NewTopicAPI(db *gorm.DB) *TopicAPI {
	return &TopicAPI{db: db}
}

func (s *TopicAPI) CreateTopic(c echo.Context) error {
	user := c.Get("user").(*models.User)

	topic := new(models.Topic)
	if err := c.Bind(topic); err != nil {
		return err
	}
	topic.OwnerID = user.ID

	if err := s.db.Create(topic).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topic)
}

func (s *TopicAPI) GetTopics(c echo.Context) error {
	var topics []models.TopicCard

	if err := s.db.Select("id, title, (?) as messages_count", s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = topics_topic.id").QueryExpr()).Find(&topics).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topics)
}

func (s *TopicAPI) GetTopic(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var topic models.TopicDetail

	if err := s.db.Where("id = ?", id).
		Select("id, title, (?) as messages_count, (?) as last_message_date",
			s.db.Table("topics_topicmessage").Select("COUNT(*)").Where("topics_topicmessage.topic_id = ?", id).QueryExpr(),
			s.db.Table("topics_topicmessage").Select("created_at").Where("topics_topicmessage.topic_id = ?", id).Order("id DESC").Limit(1).QueryExpr(),
		).Find(&topic).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, topic)
}

func (s *TopicAPI) CreateTopicMessage(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("id"))

	message := new(models.TopicMessage)
	if err := c.Bind(message); err != nil {
		return err
	}

	user := c.Get("user").(*models.User)

	message.TopicID = uint(topicID)
	message.UserID = user.ID

	if err := s.db.Create(message).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, message)
}

func (s *TopicAPI) ReadTopicMessage(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("topicID"))
	messageID, _ := strconv.Atoi(c.Param("messageID"))

	message := new(models.TopicMessage)
	if err := s.db.Where("id = ? and topic_id = ?", topicID, messageID).Find(message).Error; err != nil {
		return err
	}

	user := c.Get("user").(*models.User)

	readMessage := models.TopicMessageRead{
		TopicID:        message.TopicID,
		UserID:         user.ID,
		TopicMessageID: message.ID,
	}
	if err := s.db.Create(&readMessage).Error; err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (s *TopicAPI) GetTopicMessages(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("id"))

	var messages []models.TopicMessage

	//if err := s.db.Where("topic_id = ?", topicID).
	//	Select("topics_topicmessage.id, created_at, content, (?) as images",
	//		s.db.Table("topics_image").Select("url").Where("topics_image.id = topics_topicmessage_images.image_id").QueryExpr(),
	//	).Joins("left join topics_topicmessage_images on topics_topicmessage_images.topicmessage_id = topics_topicmessage.id").Find(&messages).Error; err != nil {
	//	return err
	//}

	if err := s.db.Preload("User").Preload("Images.Image").Where("topic_id = ?", topicID).Find(&messages).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":   len(messages),
		"results": messages,
	})
}
