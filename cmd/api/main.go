package main

import (
	"log"
	"net/http"
	"sql/internal/handlers"
	"sql/internal/store"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=taskdb port=5432 sslmode=disable"
	store, err := store.NewDBStore(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	mux := http.NewServeMux()
	tasks := handlers.TaskResource{Store: store}

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