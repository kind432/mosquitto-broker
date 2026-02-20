package gateways

import (
	"errors"
	"net/http"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"gorm.io/gorm"
)

type UserGatewayImpl struct {
	postgresClient db.PostgresClient
}

func NewUserGateway(pc db.PostgresClient) *UserGatewayImpl {
	return &UserGatewayImpl{postgresClient: pc}
}

func (u *UserGatewayImpl) GetUserByEmail(email string) (models.UserCore, error) {
	var user models.UserCore

	if err := u.postgresClient.DB.Where("email = ?", email).Take(&user).Error; err != nil {
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

func (u *UserGatewayImpl) DoesExistEmail(id uint, email string) (bool, error) {
	if err := u.postgresClient.DB.Where("id != ? AND email = ?", id, email).
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

func (u *UserGatewayImpl) CreateUser(user models.UserCore) error {
	if err := u.postgresClient.DB.Create(&user).Error; err != nil {
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func (u *UserGatewayImpl) GetUserById(id uint) (models.UserCore, error) {
	var user models.UserCore

	if err := u.postgresClient.DB.First(&user, id).Error; err != nil {
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

func (u *UserGatewayImpl) SetMosquittoOn(id uint, mosquittoOn bool) error {
	var user models.UserCore

	if err := u.postgresClient.DB.First(&user, id).Error; err != nil {
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
	if err := u.postgresClient.DB.Model(&user).Updates(updateStruct).Error; err != nil {
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
