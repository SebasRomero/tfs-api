package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitAWS() *s3.Client {
	fmt.Println("Initializing AWS")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithDefaultRegion("us-east-2"))
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	return client
	/*
		file, err := os.Open("/home/sebastian/Desktop/sebastian/programming/projects/tempfish/tfs-api/README.md")

		if err != nil {
			log.Fatal(err)
		}
	*/
	/* 	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
	   		Bucket: aws.String("tfs-f-test"),
	   		Body:   file,
	   		Key:    aws.String("ooo/README.md"),
	   	})

	   	if err != nil {
	   		log.Fatal(err)
	   	} */

	// Get the first page of results for ListObjectsV2 for a bucket
	/* 	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
	   		Bucket: aws.String("tfs-f-test"),
	   	})
	   	if err != nil {
	   		log.Fatal(err)
	   	}

	   	log.Println("first page results")
	   	for _, object := range output.Contents {
	   		log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
	   	} */
}
