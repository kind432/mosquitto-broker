package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"net/mail"

	"golang.org/x/crypto/bcrypt"

	"github.com/robboworld/mosquitto-broker/internal/models"
)

func HashPassword(s string) string {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(hashed)
}

func ComparePassword(hashed string, normal string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(normal))
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func GetOffsetAndLimit(page, pageSize *int) (offset, limit int) {
	if page == nil || pageSize == nil {
		limit = -1
		offset = 0
	} else {
		offset = (*page - 1) * *pageSize
		limit = *pageSize
	}
	return
}

func DoesHaveRole(clientRole models.Role, roles []models.Role) bool {
	for _, role := range roles {
		if role.String() == clientRole.String() {
			return true
		}
	}
	return false
}

func GetHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func StringPointerToString(p *string) string {
	var s string
	if p != nil {
		s = *p
	}
	return s
}

func BoolPointerToBool(p *bool) bool {
	var b bool
	if p != nil {
		b = *p
	}
	return b
}
