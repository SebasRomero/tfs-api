package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sebasromero/tfs-api/internal"
)

func ensureUploadsDir() {
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		err = os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			panic("Unable to create uploads directory")
		}
	}
}

func main() {
	ensureUploadsDir()
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
