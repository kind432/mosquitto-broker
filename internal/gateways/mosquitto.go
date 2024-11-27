package gateways

import (
	"github.com/robboworld/mosquitto-broker/internal/mosquitto"
	"github.com/spf13/viper"
	"runtime"
)

type MosquittoGateway interface {
	MosquittoLaunch(mosquittoOn bool)
	MosquittoStop()
	WriteMosquittoPasswd(email, password string)
	WriteNewUserToAcl(email string)
	WriteNewTopicToAcl(email, name string, canRead, canWrite bool)
	WriteUpdatedTopicToAcl(email, name string, canRead, canWrite bool)
	DeleteTopicFromAcl(username, name string)
}

type MosquittoGatewayImpl struct {
	mosquitto mosquitto.Mosquitto
}

func (m MosquittoGatewayImpl) MosquittoLaunch(mosquittoOn bool) {
	if mosquittoOn {
		args := []string{"-c", viper.GetString("mosquitto_dir_file") + "mosquitto.conf", "-v"}
		go m.mosquitto.RunCommand(viper.GetString("mosquitto_dir_exe")+"mosquitto", args...)
	} else {
		var args []string
		var command string
		if runtime.GOOS == "windows" {
			command = "taskkill"
			args = []string{"/IM", "mosquitto.exe", "/F"}
		} else {
			command = "pkill"
			args = []string{"-9", "mosquitto"}
		}
		go m.mosquitto.RunCommand(command, args...)
	}
}

func (m MosquittoGatewayImpl) MosquittoStop() {
	var args []string
	var command string
	if runtime.GOOS == "windows" {
		command = "taskkill"
		args = []string{"/IM", "mosquitto.exe", "/F"}
	} else {
		command = "pkill"
		args = []string{"-9", "mosquitto"}
	}
	go m.mosquitto.RunCommand(command, args...)
}

func (m MosquittoGatewayImpl) WriteMosquittoPasswd(email, password string) {
	args := []string{"-b", viper.GetString("mosquitto_dir_file") + "passwordfile", email, password}
	go m.mosquitto.RunCommand("mosquitto_passwd", args...)
}

func (m MosquittoGatewayImpl) WriteNewUserToAcl(email string) {
	go m.mosquitto.WriteNewUserToAcl(email)
}

func (m MosquittoGatewayImpl) WriteNewTopicToAcl(email, name string, canRead, canWrite bool) {
	go m.mosquitto.WriteNewTopicToAcl(email, name, canRead, canWrite)
}

func (m MosquittoGatewayImpl) WriteUpdatedTopicToAcl(email, name string, canRead, canWrite bool) {
	go m.mosquitto.WriteUpdatedTopicToAcl(email, name, canRead, canWrite)
}

func (m MosquittoGatewayImpl) DeleteTopicFromAcl(username, name string) {
	go m.mosquitto.DeleteTopicFromAcl(username, name)
}
