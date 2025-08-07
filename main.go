package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
/*
	all of this code isn't tested because I'm too lazy to selfhost a db
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
		get_link(app.db,"owner")
    case "POST":
		slug := ""
		url := ""
		ownerId := 0
		create_link(app.db, slug, url, ownerId)
	case "DELETE":
		slug := ""
		delete_link(app.db, slug)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
func get_link(db *sql.DB, get_type string) string {
		switch get_type {
		case "owner":
			slug := "random_slug"
			var owner int
			query := "SELECT ownerId FROM Short WHERE slug = ?"
			err := db.QueryRow(query, slug).Scan(&owner)
			if err != nil {
				if err == sql.ErrNoRows {
					panic(fmt.Errorf("no ownerId found for slug: %s", slug))
				}
				panic(err)
			}
			return strconv.Itoa(owner)
		case "clicks":
			slug := "random_slug"
			var clicks int
			query := "SELECT clicks FROM Short WHERE slug = ?"
			err := db.QueryRow(query, slug).Scan(&clicks)
			if err != nil {
				if err == sql.ErrNoRows {
					panic(fmt.Errorf("no Owner found for slug: %s", slug))
				}
				panic(err)
			}
			return strconv.Itoa(clicks)
		case "redirect":
			slug := "random_slug"
			var url string
			query := "SELECT url FROM Short WHERE slug = ?"
			err := db.QueryRow(query, slug).Scan(&url)
			if err != nil {
				if err == sql.ErrNoRows {
					panic(fmt.Errorf("no URL found for slug: %s", slug))
				}
				panic(err)
			}
			return url
		}
	return ""
}
func delete_link(db *sql.DB, slug string) error {
	query := "DELETE FROM Short WHERE slug = ?"
	result, err := db.Exec(query, slug)
	if err != nil {
		return fmt.Errorf("failed to deelte link: %v", err)
	}
	rows_affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rows_affected == 0 {
		return fmt.Errorf("no link found with slug: %s", slug)
	}
	return nil
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

