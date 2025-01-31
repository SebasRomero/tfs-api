package files

import (
	"context"
	"log"
	"strings"

	aws3 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

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
