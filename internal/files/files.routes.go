package files

import (
	"fmt"
	"net/http"
)

var arr []string

func Push(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push")

}

func Pull(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pull")
}
