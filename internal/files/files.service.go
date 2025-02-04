package files

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"

	aws3 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var key = []byte(os.Getenv("ENCRYPTION_KEY"))
var uploadsDir = "./uploads/"

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

func encryptFile(fileHeader *multipart.FileHeader) string {
	file, err := fileHeader.Open()
	if err != nil {
		log.Fatalf("open file err: %v ", err.Error())
	}

	defer file.Close()
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Fatalf("copy err: %v ", err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("gcm err: %v ", err.Error())
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatalf("nonce err: %v ", err.Error())
	}

	cipherText := gcm.Seal(nil, nonce, buf.Bytes(), nil)

	createDir(uploadsDir)

	newFile := uploadsDir + fileHeader.Filename

	err = os.WriteFile(newFile, append(nonce, cipherText...), 0777)
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
	}

	return newFile
}

func decryptFile(obj *s3.GetObjectOutput, fileName string, directoryId string) string {
	file := uploadsDir + fileName
	var buf = bytes.NewBuffer(nil)
	_, err := io.Copy(buf, obj.Body)
	if err != nil {
		log.Fatalf("copy err: %v", err.Error())
	}
	cipherText := buf.Bytes()
	if err != nil {
		log.Fatalf("read file err: %v", err.Error())
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("cipher err: %v", err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("cipher GCM err: %v", err.Error())
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		log.Fatalf("ciphertext too short")
	}

	nonce := cipherText[:nonceSize]
	cipherText = cipherText[nonceSize:]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		log.Fatalf("decrypt file err: %v", err.Error())
	}

	createDir(uploadsDir + directoryId)
	if err != nil {
		log.Fatalf("remove file err: %v", err.Error())
	}

	err = os.WriteFile(file, plainText, 0777)
	if err != nil {
		log.Fatalf("write file err: %v", err.Error())
	}

	return file
}

func createDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			panic("Unable to create uploads directory")
		}
	}
}
