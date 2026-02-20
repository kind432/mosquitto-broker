package mosquitto

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/spf13/viper"
)

const defaultFailedCode = 1

type Mosquitto interface {
	RunCommand(name string, args ...string) (stdout, stderr string, exitCode int)
	RunCommandBackground(name string, args ...string) error
	WriteNewUserToAcl(username string)
	WriteUpdatedTopicToAcl(username, name string, canRead, canWrite bool)
	DeleteTopicFromAcl(username, name string)
	WriteNewTopicToAcl(username, name string, canRead, canWrite bool)
}

type mosquitto struct {
	loggers logger.Loggers
	mu      sync.Mutex
}

func NewMosquitto(loggers logger.Loggers) Mosquitto {
	dir := viper.GetString("mosquitto_dir_file")

	if err := os.MkdirAll(dir, 0755); err != nil {
		loggers.Err.Fatalf("cannot create mosquitto dir: %v", err)
	}
	return &mosquitto{
		loggers: loggers,
	}
}

func (m *mosquitto) RunCommand(name string, args ...string) (stdout string, stderr string, exitCode int) {
	m.loggers.Info.Println("run command:", name, args)

	var outBuf, errBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Start()
	if err != nil {
		m.loggers.Err.Printf("start failed: %v", err)
		return "", err.Error(), defaultFailedCode
	}

	err = cmd.Wait()

	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = defaultFailedCode
		}
	} else {
		exitCode = cmd.ProcessState.ExitCode()
	}

	m.loggers.Info.Println("command result")
	m.loggers.Info.Printf("stdout: %s", stdout)
	m.loggers.Err.Printf("stderr: %s", stderr)
	m.loggers.Info.Printf("exitCode: %d", exitCode)
	return
}

func (m *mosquitto) RunCommandBackground(name string, args ...string) error {
	m.loggers.Info.Println("run command in background:", name, args)

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		m.loggers.Err.Printf("start failed: %v", err)
		return err
	}

	m.loggers.Info.Printf("process started with PID %d", cmd.Process.Pid)
	return nil
}

func (m *mosquitto) WriteNewUserToAcl(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	lines, err := m.readAcl(aclPath)
	if err != nil && !os.IsNotExist(err) {
		m.loggers.Err.Println(err)
		return
	}

	for _, l := range lines {
		if l == "user "+username {
			return
		}
	}

	lines = append(lines, "", "user "+username)
	if err = m.writeAclAtomic(aclPath, lines); err != nil {
		m.loggers.Err.Println(err)
	}
}

func (m *mosquitto) WriteNewTopicToAcl(username, name string, canRead, canWrite bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	lines, err := m.readAcl(aclPath)
	if err != nil {
		m.loggers.Err.Println(err)
		return
	}

	perm := permission(canRead, canWrite)
	if perm == "" {
		return
	}

	topicLine := "topic " + perm + " " + name

	var result []string
	userFound := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		result = append(result, line)

		if line == "user "+username {
			userFound = true

			result = append(result, topicLine)
		}
	}

	if !userFound {
		m.loggers.Err.Printf("user %s not found", username)
		return
	}

	if err = m.writeAclAtomic(aclPath, result); err != nil {
		m.loggers.Err.Println(err)
	}
}

func (m *mosquitto) WriteUpdatedTopicToAcl(username, name string, canRead, canWrite bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	file, err := os.Open(aclPath)
	if err != nil {
		m.loggers.Err.Printf("open acl error: %v", err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		m.loggers.Err.Println(err)
		return
	}

	perm := permission(canRead, canWrite)
	if perm == "" {
		return
	}

	newTopic := "topic " + perm + " " + name

	userFound := false
	topicUpdated := false
	inUser := false

	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "user ") {
			inUser = line == "user "+username
			if inUser {
				userFound = true
			}
			result = append(result, line)
			continue
		}
		if inUser && strings.HasPrefix(line, "topic ") {
			fields := strings.Fields(line)
			if len(fields) > 0 &&
				fields[len(fields)-1] == name {

				result = append(result, newTopic)
				topicUpdated = true
				continue
			}
		}

		result = append(result, line)
	}

	if !userFound {
		m.loggers.Err.Printf("user %s not found", username)
		return
	}

	if !topicUpdated {
		m.loggers.Err.Printf("topic %s not found", name)
		return
	}

	tmp := aclPath + ".tmp"
	data := strings.Join(result, "\n") + "\n"
	if err = os.WriteFile(tmp, []byte(data), 0644); err != nil {
		m.loggers.Err.Println(err)
		return
	}

	if err = os.Rename(tmp, aclPath); err != nil {
		m.loggers.Err.Println(err)
	}
}

func (m *mosquitto) DeleteTopicFromAcl(username, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aclPath := viper.GetString("mosquitto_dir_file") + "mosquitto.acl"

	lines, err := m.readAcl(aclPath)
	if err != nil {
		m.loggers.Err.Println(err)
		return
	}

	var result []string
	inUser := false

	for _, line := range lines {
		if strings.HasPrefix(line, "user ") {
			inUser = line == "user "+username
			result = append(result, line)
			continue
		}
		if inUser && strings.HasPrefix(line, "topic ") {
			fields := strings.Fields(line)
			if len(fields) > 1 &&
				fields[len(fields)-1] == name {
				continue
			}
		}

		result = append(result, line)
	}

	if err = m.writeAclAtomic(aclPath, result); err != nil {
		m.loggers.Err.Println(err)
	}
}

func (m *mosquitto) readAcl(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func (m *mosquitto) writeAclAtomic(path string, lines []string) error {
	tmp := path + ".tmp"

	data := strings.Join(lines, "\n") + "\n"

	if err := os.WriteFile(tmp, []byte(data), 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func permission(canRead, canWrite bool) string {
	if canRead && canWrite {
		return "readwrite"
	}
	if canRead {
		return "read"
	}
	if canWrite {
		return "write"
	}
	return ""
}
