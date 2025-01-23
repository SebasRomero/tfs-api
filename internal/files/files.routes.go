package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	aws3 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sebasromero/tfs-api/internal/aws"
)

var awsClient = aws.InitAWS()

func Push(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push")
	folderName := generateRandomID()
	bucketName := os.Getenv("AWS_BUCKET_NAME")
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

		_, err = awsClient.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws3.String(bucketName),
			Body:   file,
			Key:    aws3.String(fmt.Sprintf("%s/%s", folderName, fileHeader.Filename))})

		if err != nil {
			log.Fatal(err)
		}
		go func() {
			time.Sleep(5 * time.Minute)

			output, err := awsClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
				Bucket: &bucketName,
				Prefix: aws3.String(folderName + "/"),
			})
			if err != nil {
				log.Printf("error listing objects in folder %s: %v", folderName, err)
				return
			}

			for _, object := range output.Contents {
				awsClient.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
					Bucket: &bucketName,
					Key:    object.Key,
				})
			}

			_, err = awsClient.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: &bucketName,
				Key:    aws3.String(folderName + "/"),
			})
			if err != nil {
				log.Printf("error deleting folder %s: %v", folderName, err)
			}
		}()
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

func generateRandomID() string {
	id := strings.ToLower(uuid.New().String())

	return strings.Split(id, "-")[0]
}
