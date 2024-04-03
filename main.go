package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type APiServer struct {
	addr   string
	db     *sql.DB
	dbConn string
}

func NewApiServer(addr, dbConn string) (*APiServer, error) {
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		return nil, err
	}

	return &APiServer{
		addr:   addr,
		db:     db,
		dbConn: dbConn,
	}, nil
}

type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

func (S *APiServer) Run() error {
	err := S.autoMigrate()
	if err != nil {
		return err
	}

	router := http.NewServeMux()
	router.HandleFunc("/CreateUser", S.CreateUser)
	router.HandleFunc("/GetUser", S.GetUser)
	router.HandleFunc("/DeleteUser", S.DeleteUser)

	server := http.Server{
		Addr:    S.addr,
		Handler: RequestMiddleWare(router),
	}

	log.Printf("Server Started In Port:%s", S.addr)

	return server.ListenAndServe()
}

func (S *APiServer) autoMigrate() error {
	// Define your database schema migration here
	// For example:
	_, err := S.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT,
		age INT,
		gender TEXT
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (S *APiServer) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = S.db.Exec("INSERT INTO users (name, age, gender) VALUES ($1, $2, $3)",
		user.Name, user.Age, user.Gender)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func RequestMiddleWare(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method:%s,Path:%s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func (S *APiServer) GetUser(w http.ResponseWriter, r *http.Request) {
	rows, err := S.db.Query("SELECT name, age, gender FROM users LIMIT 1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var user User
	for rows.Next() {
		err := rows.Scan(&user.Name, &user.Age, &user.Gender)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (S *APiServer) DeleteUser(w http.ResponseWriter, r *http.Request) {
	_, err := S.db.Exec("DELETE FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Response := "Work Completed"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(Response))
}

func main() {
	dbConn := "postgres://user:password@localhost/dbname?sslmode=disable"
	server, err := NewApiServer(":8080", dbConn)
	if err != nil {
		log.Fatal("Error initializing server:", err)
	}

	if err := server.Run(); err != nil {
		log.Fatal("Server error:", err)
	}
}
