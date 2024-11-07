package gateways

import (
	"errors"
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
)

type TopicGateway interface {
	CreateTopic(topic models.TopicCore) (newTopic models.TopicCore, err error)
	DeleteTopic(id uint) (err error)
	UpdateTopicPermissions(topic models.TopicCore) (updatedTopic models.TopicCore, err error)
	DoesExistTopic(id, userId uint, name string) (exists bool, err error)
	GetTopicById(id uint) (topic models.TopicCore, err error)
	GetTopicsByUserId(userId uint, offset, limit int) (topics []models.TopicCore, countRows uint, err error)
	GetAllTopics(offset, limit int) (topics []models.TopicCore, countRows uint, err error)
}

type TopicGatewayImpl struct {
	postgresClient db.PostgresClient
}

func (t TopicGatewayImpl) CreateTopic(topic models.TopicCore) (newTopic models.TopicCore, err error) {
	if err = t.postgresClient.Db.Create(&topic).Clauses(clause.Returning{}).Error; err != nil {
		return models.TopicCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return topic, nil
}

func (t TopicGatewayImpl) DoesExistTopic(id, userId uint, name string) (bool, error) {
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

func (t TopicGatewayImpl) DeleteTopic(id uint) error {
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

func (t TopicGatewayImpl) UpdateTopicPermissions(topic models.TopicCore) (models.TopicCore, error) {
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

func (t TopicGatewayImpl) GetTopicById(id uint) (topic models.TopicCore, err error) {
	if err = t.postgresClient.Db.First(&topic, id).Error; err != nil {
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

func (t TopicGatewayImpl) GetTopicsByUserId(userId uint, offset, limit int) (topics []models.TopicCore, countRows uint, err error) {
	var count int64
	result := t.postgresClient.Db.Limit(limit).Offset(offset).Where("user_id = ?", userId).
		Find(&topics)
	if result.Error != nil {
		return []models.TopicCore{}, 0, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	}
	result.Count(&count)
	return topics, uint(count), result.Error
}

func (t TopicGatewayImpl) GetAllTopics(offset, limit int) (topics []models.TopicCore, countRows uint, err error) {
	var count int64
	result := t.postgresClient.Db.Limit(limit).Offset(offset).Find(&topics)
	if result.Error != nil {
		return []models.TopicCore{}, 0, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: result.Error.Error(),
		}
	}
	result.Count(&count)
	return topics, uint(count), result.Error
}
