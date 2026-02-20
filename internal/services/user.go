package services

import (
	"net/http"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type userService struct {
	userGateway gateways.UserGateway
}

func NewUserService(userGateway gateways.UserGateway) *userService {
	return &userService{
		userGateway: userGateway,
	}
}

func (u *userService) GetById(id uint, clientId uint, clientRole models.Role) (models.UserCore, error) {
	user, err := u.userGateway.GetById(id)
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
