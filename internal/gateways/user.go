package gateways

import (
	"errors"
	"net/http"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"gorm.io/gorm"
)

type UserGatewayImpl struct {
	db *gorm.DB
}

func NewUserGateway(db *gorm.DB) *UserGatewayImpl {
	return &UserGatewayImpl{db: db}
}

func (u *UserGatewayImpl) Create(user models.UserCore) error {
	if err := u.db.Create(&user).Error; err != nil {
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func (u *UserGatewayImpl) GetById(id uint) (models.UserCore, error) {
	var user models.UserCore

	if err := u.db.First(&user, id).Error; err != nil {
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

func (u *UserGatewayImpl) GetByEmail(email string) (models.UserCore, error) {
	var user models.UserCore

	if err := u.db.Where("email = ?", email).Take(&user).Error; err != nil {
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
	if err := u.db.Where("id != ? AND email = ?", id, email).
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

func (u *UserGatewayImpl) SetMosquittoOn(id uint, mosquittoOn bool) error {
	var user models.UserCore

	if err := u.db.First(&user, id).Error; err != nil {
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
	if err := u.db.Model(&user).Updates(updateStruct).Error; err != nil {
		return utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}
