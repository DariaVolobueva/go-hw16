package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

type TaskResource struct {
	s *DBStore
}

func (t *TaskResource) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := t.s.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func (t *TaskResource) CreateOne(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = t.s.Add(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (t *TaskResource) GetOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	task, err := t.s.Get(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (t *TaskResource) UpdateOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var task models.Task
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task.ID = uint(id)
	err = t.s.Update(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (t *TaskResource) DeleteOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	err = t.s.Delete(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=taskdb port=5432 sslmode=disable"
	store, err := NewDBStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	mux := http.NewServeMux()
	tasks := TaskResource{s: store}

	mux.HandleFunc("GET /tasks", tasks.GetAll)
	mux.HandleFunc("POST /tasks", tasks.CreateOne)
	mux.HandleFunc("GET /tasks/{id}", tasks.GetOne)
	mux.HandleFunc("PUT /tasks/{id}", tasks.UpdateOne)
	mux.HandleFunc("DELETE /tasks/{id}", tasks.DeleteOne)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}