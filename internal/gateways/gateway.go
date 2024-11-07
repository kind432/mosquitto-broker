package gateways

import (
	"github.com/robboworld/mosquitto-broker/internal/db"
	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
	"go.uber.org/fx"
)

type Gateways struct {
	fx.Out
	UserGateway      UserGateway
	MosquittoGateway MosquittoGateway
	TopicGateway     TopicGateway
}

func SetupGateways(
	pc db.PostgresClient,
	mosquitto mosquitto.Mosquitto,
) Gateways {
	return Gateways{
		UserGateway:      UserGatewayImpl{pc},
		MosquittoGateway: MosquittoGatewayImpl{mosquitto},
		TopicGateway:     TopicGatewayImpl{pc},
	}
}
