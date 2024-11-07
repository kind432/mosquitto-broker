package services

import (
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"net/http"
)

type UserService interface {
	GetUserById(id uint, clientId uint, clientRole models.Role) (models.UserCore, error)
}

type UserServiceImpl struct {
	userGateway gateways.UserGateway
}

func (u UserServiceImpl) GetUserById(id uint, clientId uint, clientRole models.Role) (models.UserCore, error) {
	user, err := u.userGateway.GetUserById(id)
	if err != nil {
		return models.UserCore{}, err
	}

	if clientRole.String() != models.RoleSuperAdmin.String() && user.ID != clientId {
		return models.UserCore{}, utils.ResponseError{
			Code:    http.StatusForbidden,
			Message: consts.ErrAccessDenied,
		}
	}

	return user, nil
}
