package gateways

import (
	"errors"
	"net/http"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type topicGateway struct {
	postgresClient db.PostgresClient
}

func NewTopicGateway(pc db.PostgresClient) *topicGateway {
	return &topicGateway{pc}
}

func (t *topicGateway) CreateTopic(topic models.TopicCore) (models.TopicCore, error) {
	if err := t.postgresClient.Db.Create(&topic).Clauses(clause.Returning{}).Error; err != nil {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return topic, nil
}

func (t *topicGateway) DoesExistTopic(id, userId uint, name string) (bool, error) {
	if err := t.postgresClient.Db.Where("id != ? AND user_id = ? AND name = ?", id, userId, name).
		Take(&models.TopicCore{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return true, nil
}

func (t *topicGateway) DeleteTopic(id uint) error {
	if err := t.postgresClient.Db.Delete(&models.TopicCore{}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ResponseError{
				Code:    http.StatusBadRequest,
				Message: consts.ErrNotFoundInDB,
			}
		}
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func (t *topicGateway) UpdateTopicPermissions(topic models.TopicCore) (models.TopicCore, error) {
	var existingTopic models.TopicCore
	if err := t.postgresClient.Db.First(&existingTopic, topic.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.TopicCore{}, utils.ResponseError{
				Code:    http.StatusBadRequest,
				Message: consts.ErrNotFoundInDB,
			}
		}
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	if err := t.postgresClient.Db.Model(&existingTopic).
		Updates(map[string]interface{}{
			"can_read":  topic.CanRead,
			"can_write": topic.CanWrite,
		}).Error; err != nil {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return existingTopic, nil
}

func (t *topicGateway) GetTopicById(id uint) (models.TopicCore, error) {
	var topic models.TopicCore

	if err := t.postgresClient.Db.First(&topic, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.TopicCore{}, utils.ResponseError{
				Code:    http.StatusBadRequest,
				Message: consts.ErrNotFoundInDB,
			}
		}
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return topic, nil
}

func (t *topicGateway) GetTopicsByUserId(userId uint, offset, limit int) ([]models.TopicCore, uint, error) {
	var topics []models.TopicCore
	var count int64

	result := t.postgresClient.Db.
		Limit(limit).
		Offset(offset).
		Where("user_id = ?", userId).
		Find(&topics)
	if result.Error != nil {
		return []models.TopicCore{}, 0, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	}
	result.Count(&count)
	return topics, uint(count), nil
}

func (t *topicGateway) GetAllTopics(offset, limit int) ([]models.TopicCore, uint, error) {
	var topics []models.TopicCore
	var count int64

	result := t.postgresClient.Db.Limit(limit).Offset(offset).Find(&topics)
	if result.Error != nil {
		return []models.TopicCore{}, 0, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	}
	result.Count(&count)
	return topics, uint(count), nil
}
