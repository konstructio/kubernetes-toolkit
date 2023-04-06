package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	log "github.com/sirupsen/logrus"
)

// GetECRAuthToken
func (conf *AWSConfiguration) GetECRAuthToken() (string, error) {
	log.Info("getting ecr auth token")
	ecrClient := ecr.NewFromConfig(conf.Config)

	token, err := ecrClient.GetAuthorizationToken(context.Background(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	return *token.AuthorizationData[0].AuthorizationToken, nil
}
