package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sebasromero/tfs-api/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Listen in port:", port)
	err := http.ListenAndServe(":"+port, internal.MainHandler())
	if err != nil {
		log.Panic(err)
	}
}
