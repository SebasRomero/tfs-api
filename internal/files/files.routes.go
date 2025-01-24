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
	"strings"
	"time"

	aws3 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/sebasromero/tfs-api/internal/aws"
)

var awsClient = aws.InitAWS()

func Push(w http.ResponseWriter, r *http.Request) {
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
		link := "http://localhost:8080/pull/" + folderName
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
		fmt.Fprintf(w, "files uploaded successfully: get them with this link: %s\n", link)
	}

}

func Pull(w http.ResponseWriter, r *http.Request) {
	directoryId := r.PathValue("id")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	bucketName := os.Getenv("AWS_BUCKET_NAME")

	output := listFiles(bucketName, directoryId)

	if len(output.Contents) == 0 {
		http.Error(w, "no files found", http.StatusNotFound)
		return
	}

	for _, entry := range output.Contents {
		fileName := *entry.Key

		part, err := writer.CreateFormFile("files", fileName)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating form file for %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}

		obj, err := awsClient.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    entry.Key,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("error getting object %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(part, obj.Body)

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

func listFiles(bucketName string, folderName string) *s3.ListObjectsV2Output {
	output, err := awsClient.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
		Prefix: aws3.String(folderName + "/"),
	})
	if err != nil {
		log.Printf("error listing objects in folder %s: %v", folderName, err)
		return nil
	}

	return output
}
