package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Task struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type DBStore struct {
	db *gorm.DB
}

func NewDBStore(dsn string) (*DBStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DBStore{db: db}, nil
}

func (s *DBStore) Add(task *Task) error {
	return s.db.Create(task).Error
}

func (s *DBStore) Get(id uint) (*Task, error) {
	var task Task
	err := s.db.First(&task, id).Error
	return &task, err
}

func (s *DBStore) Update(task *Task) error {
	return s.db.Save(task).Error
}

func (s *DBStore) Delete(id uint) error {
	return s.db.Delete(&Task{}, id).Error
}

func (s *DBStore) GetAll() ([]Task, error) {
	var tasks []Task
	err := s.db.Find(&tasks).Error
	return tasks, err
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
	var task Task
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
	var task Task
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
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}

    err = store.db.AutoMigrate(&Task{})
    if err != nil {
        fmt.Printf("Failed to run auto migration: %v\n", err)
        return
    }

	mux := http.NewServeMux()
	tasks := TaskResource{s: store}

	mux.HandleFunc("GET /tasks", tasks.GetAll)
	mux.HandleFunc("POST /tasks", tasks.CreateOne)
	mux.HandleFunc("GET /tasks/{id}", tasks.GetOne)
	mux.HandleFunc("PUT /tasks/{id}", tasks.UpdateOne)
	mux.HandleFunc("DELETE /tasks/{id}", tasks.DeleteOne)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Failed to listen and serve: %v\n", err)
	}
}