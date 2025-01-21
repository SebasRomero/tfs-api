package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Push(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push")

	err := r.ParseMultipartForm(10 << 20) // Limit the size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Unable to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		out, err := os.Create(fmt.Sprintf("./uploads/%s", fileHeader.Filename))
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "Unable to save file", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "File %s uploaded successfully\n", fileHeader.Filename)
	}

}

func Pull(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pull")
}
