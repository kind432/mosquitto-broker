package services

import (
	"github.com/robboworld/mosquitto-broker/internal/gateways"
)

type mosquittoService struct {
	userGateway      gateways.UserGateway
	mosquittoGateway gateways.MosquittoGateway
}

func NewMosquittoService(
	userGateway gateways.UserGateway,
	mosquittoGateway gateways.MosquittoGateway,
) *mosquittoService {
	return &mosquittoService{
		userGateway:      userGateway,
		mosquittoGateway: mosquittoGateway,
	}
}

func (m *mosquittoService) MosquittoLaunch(id uint, mosquittoOn bool) error {
	err := m.userGateway.SetMosquittoOn(id, mosquittoOn)
	if err != nil {
		return err
	}

	if mosquittoOn {
		m.mosquittoGateway.MosquittoLaunch(mosquittoOn)
	} else {
		m.mosquittoGateway.MosquittoStop()
	}

	return nil
}

func (m *mosquittoService) MosquittoStop() {
	m.mosquittoGateway.MosquittoStop()
}
