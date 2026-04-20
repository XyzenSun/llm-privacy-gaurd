package repository

import (
	"errors"

	"llm-privacy-gaurd/internal/domain"

	"gorm.io/gorm"
)

type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

func (r *SystemConfigRepository) GetAll() ([]domain.SystemConfig, error) {
	var items []domain.SystemConfig
	err := r.db.Order("config_key asc").Find(&items).Error
	return items, err
}

func (r *SystemConfigRepository) GetByKey(key string) (*domain.SystemConfig, error) {
	var item domain.SystemConfig
	if err := r.db.First(&item, "config_key = ?", key).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SystemConfigRepository) Set(key, value string) error {
	var item domain.SystemConfig
	err := r.db.First(&item, "config_key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.Create(&domain.SystemConfig{
				ID:        newRepoID(),
				ConfigKey: key,
				Value:     value,
			}).Error
		}
		return err
	}
	item.Value = value
	return r.db.Save(&item).Error
}

func (r *SystemConfigRepository) SetMany(values map[string]string) error {
	for key, value := range values {
		if err := r.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func newRepoID() string {
	return newSimpleID()
}
