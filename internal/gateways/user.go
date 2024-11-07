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

type UserGateway interface {
	CreateUser(user models.UserCore) (newUser models.UserCore, err error)
	GetUserById(id uint) (user models.UserCore, err error)
	GetUserByEmail(email string) (user models.UserCore, err error)
	DoesExistEmail(id uint, email string) (bool, error)
	SetMosquittoOn(id uint, mosquittoOn bool) error
}

type UserGatewayImpl struct {
	postgresClient db.PostgresClient
}

func (u UserGatewayImpl) GetUserByEmail(email string) (user models.UserCore, err error) {
	if err = u.postgresClient.Db.Where("email = ?", email).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, utils.ResponseError{
				Code:    http.StatusBadRequest,
				Message: consts.ErrUserWithEmailNotFound,
			}
		}
		return user, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return user, nil
}

func (u UserGatewayImpl) DoesExistEmail(id uint, email string) (bool, error) {
	if err := u.postgresClient.Db.Where("id != ? AND email = ?", id, email).
		Take(&models.UserCore{}).Error; err != nil {
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

func (u UserGatewayImpl) CreateUser(user models.UserCore) (newUser models.UserCore, err error) {
	if err = u.postgresClient.Db.Create(&user).Clauses(clause.Returning{}).Error; err != nil {
		return models.UserCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return user, nil
}

func (u UserGatewayImpl) GetUserById(id uint) (user models.UserCore, err error) {
	if err = u.postgresClient.Db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.UserCore{}, utils.ResponseError{
				Code:    http.StatusBadRequest,
				Message: consts.ErrNotFoundInDB,
			}
		}
		return models.UserCore{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return user, nil
}

func (u UserGatewayImpl) SetMosquittoOn(id uint, mosquittoOn bool) error {
	var user models.UserCore
	if err := u.postgresClient.Db.First(&user, id).Error; err != nil {
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

	updateStruct := map[string]interface{}{
		"mosquitto_on": mosquittoOn,
	}
	if err := u.postgresClient.Db.Model(&user).Updates(updateStruct).Error; err != nil {
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
