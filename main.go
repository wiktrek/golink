package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

/*
	Create DB
	Add methods for Link
		Get types:
			redirect
			name
			owner
	Add methods for User
		Get  types:
			Links
*/
type User struct {
	id int 
	username string
	password string 
	email string
	createdAt string
} 
type Short struct {
	slug string
	url string
	ownerId int
	createdAt string
	expiresAt string
	isActive bool
	clicks int
}
type App struct {
	db *sql.DB
}
func (app *App) routes() {

	fileserver := http.FileServer(http.Dir("static"))
	http.Handle("/", fileserver)
	http.HandleFunc("/api/user", app.userHandler)
    http.HandleFunc("/api/link", app.linkHandler)
}

func main() {

	err := godotenv.Load()
 	if err != nil {
  		log.Fatal("Error loading .env file")
  	}
	DB := os.Getenv("DB")
	db, err := sql.Open("mysql", DB)
	
	if err != nil {
		panic(err)
	}
	
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	
	defer db.Close()

	err = db.Ping()
    
	if err != nil {
        panic(err.Error())
    }
	
	fmt.Println("Connected to db!")

	app := &App{db: db}
    app.routes()

    fmt.Println("Server is running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
func (app *App) linkHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
    case "POST":
		
		create_link(app.db, slug, url, ownerId)
	case "DELETE":
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
func get_link() {
	get_type := "name"
	
	switch get_type {
	case "owner":

	case "name":
	case "redirect":
	}
}
func delete_link() {

}
func create_link(db *sql.DB, slug, url string, ownerId int) error {
    query := "INSERT INTO Short (slug, url, ownerId) VALUES (?, ?, ?)"
    _, err := db.Exec(query, slug, url, ownerId)
    return err
}

func (app *App) userHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
    case "POST":
	case "DELETE":
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}


func get_user() {

}
func delete_user() {

}
func create_user() {

}

