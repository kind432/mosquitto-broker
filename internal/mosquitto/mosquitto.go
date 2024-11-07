package mosquitto

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const defaultFailedCode = 1

type Mosquitto interface {
	RunCommand(name string, args ...string) (stdout, stderr string, exitCode int)
	WriteNewUserToAcl(username string)
	WriteUpdatedTopicToAcl(username, name string, canRead, canWrite bool)
	DeleteTopicFromAcl(username, name string)
	WriteNewTopicToAcl(username, name string, canRead, canWrite bool)
}

type MosquittoImpl struct {
	loggers logger.Loggers
}

func InitMosquitto(loggers logger.Loggers) Mosquitto {
	return &MosquittoImpl{
		loggers: loggers,
	}
}

func (m MosquittoImpl) RunCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	m.loggers.Info.Println("run command:", name, args)

	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
			} else {
				exitCode = defaultFailedCode
			}
		} else {
			m.loggers.Err.Printf("Could not get exit code for failed program: %v, %v", name, args)
			exitCode = defaultFailedCode
			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
			exitCode = ws.ExitStatus()
		} else {
			exitCode = defaultFailedCode
		}
	}
	m.loggers.Info.Println("command result")
	m.loggers.Info.Printf("stdout: %v", stdout)
	m.loggers.Err.Printf("stderr: %v", stderr)
	m.loggers.Info.Printf("exitCode: %v", exitCode)
	return
}

func (m MosquittoImpl) WriteNewUserToAcl(username string) {
	f, err := os.OpenFile(viper.GetString("mosquitto_dir_file")+"mosquitto.acl",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		m.loggers.Err.Printf("Could not open mosquitto.acl: %v", err)
	}
	defer f.Close()

	data := "user " + username + "\r\n"
	if _, err = f.WriteString(data); err != nil {
		m.loggers.Err.Printf("Could not write to mosquitto.acl: %v", err)
	}
}

func (m MosquittoImpl) WriteNewTopicToAcl(username, name string, canRead, canWrite bool) {
	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	file, err := os.Open(aclPath)
	if err != nil {
		m.loggers.Err.Printf("Could not open mosquitto.acl for reading: %v", err)
		return
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	userFound := false

	var permissions string
	if canRead && canWrite {
		permissions = "readwrite"
	} else if canRead {
		permissions = "read"
	} else if canWrite {
		permissions = "write"
	}

	for scanner.Scan() {
		line := scanner.Text()
		buffer.WriteString(line + "\n")

		if line == "user "+username {
			userFound = true
			buffer.WriteString("topic " + permissions + " " + name + "\n")
		}
	}

	if err = scanner.Err(); err != nil {
		m.loggers.Err.Printf("Error reading mosquitto.acl: %v", err)
		return
	}

	if !userFound {
		m.loggers.Err.Printf("Could not find user %s", username)
		return
	}

	if err = os.WriteFile(aclPath, buffer.Bytes(), 0644); err != nil {
		m.loggers.Err.Printf("Could not write updated data to mosquitto.acl: %v", err)
	}
}

func (m MosquittoImpl) WriteUpdatedTopicToAcl(username, name string, canRead, canWrite bool) {
	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	file, err := os.Open(aclPath)
	if err != nil {
		m.loggers.Err.Printf("Could not open mosquitto.acl for reading: %v", err)
		return
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	userFound := false
	topicUpdated := false

	var permissions string
	if canRead && canWrite {
		permissions = "readwrite"
	} else if canRead {
		permissions = "read"
	} else if canWrite {
		permissions = "write"
	}

	var topicLine string
	if permissions == "" {
		topicLine = "topic " + name
	} else {
		topicLine = "topic " + permissions + " " + name
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "user "+username {
			userFound = true
			buffer.WriteString(line + "\n")

			for scanner.Scan() {
				nextLine := scanner.Text()
				if strings.HasPrefix(nextLine, "topic") && strings.Contains(nextLine, name) {
					buffer.WriteString(topicLine + "\n")
					topicUpdated = true
				} else {
					buffer.WriteString(nextLine + "\n")
				}

				if nextLine == "" {
					break
				}
			}
			if !topicUpdated {
				m.loggers.Err.Printf("Could not find topic %s", topicLine)
				return
			}
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	if err = scanner.Err(); err != nil {
		m.loggers.Err.Printf("Error reading mosquitto.acl: %v", err)
		return
	}

	if !userFound {
		m.loggers.Err.Printf("Could not find user %s", username)
		return
	}

	if err = os.WriteFile(aclPath, buffer.Bytes(), 0644); err != nil {
		m.loggers.Err.Printf("Could not write updated data to mosquitto.acl: %v", err)
	}
}

func (m MosquittoImpl) DeleteTopicFromAcl(username, name string) {
	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	file, err := os.Open(aclPath)
	if err != nil {
		m.loggers.Err.Printf("Could not open mosquitto.acl for reading: %v", err)
		return
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	userFound := false
	topicDeleted := false

	readTopic := "topic read " + name
	writeTopic := "topic write " + name
	readWriteTopic := "topic readwrite " + name
	topic := "topic " + name

	for scanner.Scan() {
		line := scanner.Text()
		if line == "user "+username {
			userFound = true
			buffer.WriteString(line + "\n")

			for scanner.Scan() {
				nextLine := scanner.Text()
				if nextLine == readTopic || nextLine == writeTopic ||
					nextLine == readWriteTopic || nextLine == topic {
					topicDeleted = true
					continue
				}

				buffer.WriteString(nextLine + "\n")
				if nextLine == "" {
					break
				}
			}
		} else {
			buffer.WriteString(line + "\n")
		}
	}

	if err = scanner.Err(); err != nil {
		m.loggers.Err.Printf("Error reading mosquitto.acl: %v", err)
		return
	}

	if !userFound {
		m.loggers.Err.Printf("Could not find user %s", username)
		return
	}
	if !topicDeleted {
		m.loggers.Err.Printf("Could not find topic %s for user %s with the specified permissions", name, username)
		return
	}

	if err = os.WriteFile(aclPath, buffer.Bytes(), 0644); err != nil {
		m.loggers.Err.Printf("Could not write updated data to mosquitto.acl: %v", err)
	}
}
