package gateways

import (
	"runtime"

	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
	"github.com/spf13/viper"
)

type mosquittoGateway struct {
	mosquitto mosquitto.Mosquitto
}

func NewMosquittoGateway(mosquitto mosquitto.Mosquitto) *mosquittoGateway {
	return &mosquittoGateway{mosquitto}
}

func (m *mosquittoGateway) MosquittoLaunch(mosquittoOn bool) {
	if mosquittoOn {
		args := []string{
			"-c",
			viper.GetString("mosquitto_dir_file") + "mosquitto.conf",
			"-v",
		}
		_, _, code :=
			m.mosquitto.RunCommand(
				viper.GetString("mosquitto_dir_exe")+"mosquitto",
				args...,
			)

		if code != 0 {
			//TODO: Обработка ошибки
			return
		}
		return
	}

	m.MosquittoStop()
}

func (m *mosquittoGateway) MosquittoStop() {
	var command string
	var args []string

	if runtime.GOOS == "windows" {
		command = "taskkill"
		args = []string{"/IM", "mosquitto.exe"}
	} else {
		command = "pkill"
		args = []string{"mosquitto"}
	}

	m.mosquitto.RunCommand(command, args...)
}

func (m *mosquittoGateway) WriteMosquittoPasswd(email, password string) {
	args := []string{
		"-b",
		viper.GetString("mosquitto_dir_file") + "passwordfile",
		email,
		password,
	}
	_, _, code := m.mosquitto.RunCommand("mosquitto_passwd", args...)
	if code != 0 {
		//TODO: Обработка ошибки
		return
	}
}

func (m *mosquittoGateway) WriteNewUserToAcl(email string) {
	m.mosquitto.WriteNewUserToAcl(email)
}

func (m *mosquittoGateway) WriteNewTopicToAcl(email, name string, canRead, canWrite bool) {
	m.mosquitto.WriteNewTopicToAcl(email, name, canRead, canWrite)
}

func (m *mosquittoGateway) WriteUpdatedTopicToAcl(email, name string, canRead, canWrite bool) {
	m.mosquitto.WriteUpdatedTopicToAcl(email, name, canRead, canWrite)
}

func (m *mosquittoGateway) DeleteTopicFromAcl(username, name string) {
	m.mosquitto.DeleteTopicFromAcl(username, name)
}
