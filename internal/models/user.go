package models

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

type UserHTTP struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        Role   `json:"role"`
	FullName    string `json:"full_name"`
	MosquittoOn bool   `json:"mosquitto_on"`
}

type UserCore struct {
	ID          uint `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Email       string         `gorm:"not null;"`
	Password    string         `gorm:"not null;"`
	Role        Role           `gorm:"not null;"`
	FullName    string         `gorm:"not null;"`
	MosquittoOn bool           `gorm:"not null;default:false"`
}

func (u *UserHTTP) ToCore() UserCore {
	id, _ := strconv.ParseUint(u.ID, 10, 64)
	return UserCore{
		ID:       uint(id),
		Email:    u.Email,
		Password: u.Password,
		Role:     u.Role,
		FullName: u.FullName,
	}
}

func (u *UserHTTP) FromCore(userCore UserCore) {
	u.ID = strconv.Itoa(int(userCore.ID))
	u.CreatedAt = userCore.CreatedAt.Format(time.DateTime)
	u.UpdatedAt = userCore.UpdatedAt.Format(time.DateTime)
	u.Email = userCore.Email
	u.FullName = userCore.FullName
	u.Role = userCore.Role
	u.MosquittoOn = userCore.MosquittoOn
}

func FromUsersCore(usersCore []UserCore) (usersHttp []*UserHTTP) {
	for _, userCore := range usersCore {
		var tmpUserHttp UserHTTP
		tmpUserHttp.FromCore(userCore)
		usersHttp = append(usersHttp, &tmpUserHttp)
	}
	return
}
