package main

import (
	"fmt"
	"log"
	"net/http"
)

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
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}
func linkHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
    case "POST":
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}