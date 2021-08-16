package db

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type DeletedModel struct {
	Model
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
}
