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
	
	fileserver := http.FileServer(http.Dir("static"))
	http.Handle("/", fileserver)
    http.HandleFunc("/api/user/", userHandler)
	http.HandleFunc("/api/link/", linkHandler)
    fmt.Println("Server is running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
func userHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
    case "POST":
	case "DELETE":
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
func linkHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
    case "POST":
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
func create_link() {

}
func get_user() {

}
func delete_user() {

}
func create_user() {

}

