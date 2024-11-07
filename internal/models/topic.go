package models

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

type TopicHTTP struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	UserId    string `json:"user_id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	CanRead   bool   `json:"can_read"`
	CanWrite  bool   `json:"can_write"`
}

type TopicCore struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserId    uint
	User      UserCore `gorm:"foreignKey:UserId"`
	Name      string   `gorm:"not null"`
	Password  string   `gorm:"not null"`
	CanRead   bool     `gorm:"not null;default:false"`
	CanWrite  bool     `gorm:"not null;default:false"`
}

func (t *TopicHTTP) ToCore() TopicCore {
	id, _ := strconv.ParseUint(t.ID, 10, 64)
	return TopicCore{
		ID:       uint(id),
		Name:     t.Name,
		Password: t.Password,
	}
}

func (t *TopicHTTP) FromCore(topicCore TopicCore) {
	t.ID = strconv.Itoa(int(topicCore.ID))
	t.CreatedAt = topicCore.CreatedAt.Format(time.DateTime)
	t.UpdatedAt = topicCore.UpdatedAt.Format(time.DateTime)
	t.UserId = strconv.Itoa(int(topicCore.UserId))
	t.Name = topicCore.Name
	t.Password = topicCore.Password
	t.CanRead = topicCore.CanRead
	t.CanWrite = topicCore.CanWrite
}

func FromTopicsCore(topicsCore []TopicCore) (topicsHttp []*TopicHTTP) {
	for _, topicCore := range topicsCore {
		var tmpTopicHttp TopicHTTP
		tmpTopicHttp.FromCore(topicCore)
		topicsHttp = append(topicsHttp, &tmpTopicHttp)
	}
	return
}
