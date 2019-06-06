package apis

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	apiErrors "moowda/errors"
	"moowda/models"
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

type CreateTopicMessageRequest struct {
	Content string `json:"content"`
	Images  []uint `json:"images"`
}

func (r CreateTopicMessageRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Content, validation.Required),
		validation.Field(&r.Images, validation.NilOrNotEmpty),
	)
}

func (s *TopicAPI) CreateTopicMessage(c echo.Context) error {
	topicID, _ := strconv.Atoi(c.Param("id"))

	createTopicMessageRequest := new(CreateTopicMessageRequest)
	if err := c.Bind(createTopicMessageRequest); err != nil {
		return err
	}
	if err := createTopicMessageRequest.Validate(); err != nil {
		return apiErrors.InvalidData(err.(validation.Errors))
	}

	user := c.Get("user").(*models.User)

	message := models.TopicMessage{
		Content: createTopicMessageRequest.Content,
		TopicID: uint(topicID),
		UserID:  user.ID,
	}

	var image models.Image
	for _, imageID := range createTopicMessageRequest.Images {
		if err := s.db.Where("id = ?", imageID).First(&image).Error; err != nil {
			return err
		}

		message.Images = append(message.Images, models.TopicMessageImage{
			ImageID: image.ID,
		})
	}

	if err := s.db.Create(&message).Error; err != nil {
		return err
	}

	if err := s.db.Preload("User").Preload("Images.Image").Where("id = ?", message.ID).Find(&message).Error; err != nil {
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

	if err := s.db.Preload("User").Preload("Images.Image").Where("topic_id = ?", topicID).Find(&messages).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":   len(messages),
		"results": messages,
	})
}
