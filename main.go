package main

import (
	"fmt"
	"log"
	"net/http"
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

