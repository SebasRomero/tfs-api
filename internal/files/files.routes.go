package files

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	fp "path/filepath"
)

func Push(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push")

	err := r.ParseMultipartForm(10 << 20) // Limit the size to 10 MB
	if err != nil {
		http.Error(w, "unable to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "unable to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		out, err := os.Create(fmt.Sprintf("./uploads/%s", fileHeader.Filename))
		if err != nil {
			http.Error(w, "unable to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			http.Error(w, "unable to save file", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "file %s uploaded successfully\n", fileHeader.Filename)
	}

}

func Pull(w http.ResponseWriter, r *http.Request) {
	fmt.Println("pull")

	dir := "../uploads"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "error getting the files", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		fileName := entry.Name()
		newFilePath := fp.Join(dir, fileName)
		file, err := os.Open(newFilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("error opening file %s: %v", newFilePath, err), http.StatusInternalServerError)
			return
		}

		part, err := writer.CreateFormFile("files", fileName)
		if err != nil {
			file.Close()
			http.Error(w, fmt.Sprintf("error creating form file for %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("error copying file %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}
	}

	writer.Close()
	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.Write(body.Bytes())
}
