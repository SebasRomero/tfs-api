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
}
