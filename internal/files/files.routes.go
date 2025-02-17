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
	"time"

	aws3 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
		encryptedFile := encryptFile(fileHeader)
		file, err := os.Open(encryptedFile)
		if err != nil {
			http.Error(w, "unable to open uploaded file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		defer os.Remove(encryptedFile)
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
		fmt.Fprintf(w, "files uploaded successfully: \nGet them with this directory code: %s\n", folderName)
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
			Key:    &fileName,
		})

		if err != nil {
			http.Error(w, fmt.Sprintf("error getting object %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}

		fileDecrypted := decryptFile(obj, fileName, directoryId)
		openedFile, err := os.Open(fileDecrypted)
		if err != nil {
			http.Error(w, fmt.Sprintf("error opening file %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}
		defer os.Remove(uploadsDir + directoryId)
		defer os.Remove(fileDecrypted)
		defer openedFile.Close()
		_, err = io.Copy(part, openedFile)

		if err != nil {
			http.Error(w, fmt.Sprintf("error copying file %s: %v", fileName, err), http.StatusInternalServerError)
			return
		}
	}

	writer.Close()
	w.Header().Set("Content-Type", writer.FormDataContentType())
	w.Write(body.Bytes())
}
