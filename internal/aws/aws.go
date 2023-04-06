package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/sirupsen/logrus"
)

// NewAwsV2 instantiates an AWS client
// The following environment variables are required:
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
func NewAwsV2(region string) aws.Config {
	awsClient, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("unable to create aws client")
	}

	return awsClient
}
