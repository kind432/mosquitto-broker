package services

import (
	"github.com/robboworld/mosquitto-broker/internal/gateways"
)

type MosquittoService interface {
	MosquittoLaunch(id uint, mosquittoOn bool) error
	MosquittoStop()
}

type MosquittoServiceImpl struct {
	mosquittoGateway gateways.MosquittoGateway
	userGateway      gateways.UserGateway
}

func (m MosquittoServiceImpl) MosquittoLaunch(id uint, mosquittoOn bool) error {
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

func (m MosquittoServiceImpl) MosquittoStop() {
	m.mosquittoGateway.MosquittoStop()
}
