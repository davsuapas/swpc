package crypto_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swpoolcontroller/pkg/crypto"
)

func TestAWSSecret_Get(t *testing.T) {
	t.Parallel()

	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		assert.Fail(t, "LoadDefaultConfig")
	}

	s := crypto.NewAWSSecret(config)

	_, err = s.Get("secretName")

	require.Error(t, err)
}
