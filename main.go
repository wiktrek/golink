package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {


    fmt.Println("Server is running at http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
