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

func (t *topicGateway) Create(topic models.TopicCore) (models.TopicCore, error) {
	if err := t.postgresClient.DB.Create(&topic).Clauses(clause.Returning{}).Error; err != nil {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return topic, nil
}

func (t *topicGateway) GetById(id uint) (models.TopicCore, error) {
	var topic models.TopicCore

	if err := t.postgresClient.DB.First(&topic, id).Error; err != nil {
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

func (t *topicGateway) GetByUserId(userId uint, offset, limit int) ([]models.TopicCore, uint, error) {
	var topics []models.TopicCore
	var count int64

	result := t.postgresClient.DB.
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

func (t *topicGateway) GetAll(offset, limit int) ([]models.TopicCore, uint, error) {
	var topics []models.TopicCore
	var count int64

	result := t.postgresClient.DB.Limit(limit).Offset(offset).Find(&topics)
	if result.Error != nil {
		return []models.TopicCore{}, 0, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	}
	result.Count(&count)
	return topics, uint(count), nil
}

func (t *topicGateway) UpdatePermissions(topic models.TopicCore) (models.TopicCore, error) {
	var existingTopic models.TopicCore
	if err := t.postgresClient.DB.First(&existingTopic, topic.ID).Error; err != nil {
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

	if err := t.postgresClient.DB.Model(&existingTopic).
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

func (t *topicGateway) Delete(id uint) error {
	if err := t.postgresClient.DB.Delete(&models.TopicCore{}, id).Error; err != nil {
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

func (t *topicGateway) DoesExist(id, userId uint, name string) (bool, error) {
	if err := t.postgresClient.DB.Where("id != ? AND user_id = ? AND name = ?", id, userId, name).
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
