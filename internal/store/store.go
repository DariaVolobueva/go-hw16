package store

import (
	"fmt"
	"sql/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBStore struct {
	db *gorm.DB
}

func NewDBStore(dsn string) (*DBStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &DBStore{db: db}, nil
}

func (s *DBStore) Add(task *models.Task) error {
	if err := s.db.Create(task).Error; err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	return nil
}

func (s *DBStore) Get(id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get task with id %d: %w", id, err)
	}
	return &task, nil
}

func (s *DBStore) Update(task *models.Task) error {
	if err := s.db.Save(task).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (s *DBStore) Delete(id uint) error {
	if err := s.db.Delete(&models.Task{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (s *DBStore) GetAll() ([]models.Task, error) {
	var tasks []models.Task
	if err := s.db.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get all tasks: %w", err)
	}
	return tasks, nil
}