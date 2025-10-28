package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;" json:"id"`
	ServiceName string     `gorm:"type:varchar(200);not null;index" json:"service_name"`
	Price       int        `gorm:"not null" json:"price"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartDate   time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate     *time.Time `gorm:"type:date" json:"end_date,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (s *Subscription) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
