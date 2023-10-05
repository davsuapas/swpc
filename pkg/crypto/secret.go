package crypto

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/pkg/errors"
)

const (
	errSValue = "Getting aws secret"
)

// AWSSecret manages the secrets
type AWSSecret struct {
	secret *secretsmanager.Client
}

// NewAWSSecret creates a secret handler
func NewAWSSecret(cnf aws.Config) *AWSSecret {
	return &AWSSecret{
		secret: secretsmanager.NewFromConfig(cnf),
	}
}

// Get gets the secret from aws manager
func (s *AWSSecret) Get(secretName string) (map[string]string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	secretValues := make(map[string]string)

	result, err := s.secret.GetSecretValue(context.TODO(), input)
	if err != nil {
		return secretValues, errors.Wrap(err, errSValue)
	}

	if err := json.Unmarshal([]byte(*result.SecretString), &secretValues); err != nil {
		return secretValues, errors.Wrap(err, errSValue)
	}

	return secretValues, nil
}
