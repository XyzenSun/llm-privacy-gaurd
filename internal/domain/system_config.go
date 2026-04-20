package domain

import "time"

type SystemConfig struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	ConfigKey string    `gorm:"size:100;uniqueIndex;not null" json:"configKey"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
