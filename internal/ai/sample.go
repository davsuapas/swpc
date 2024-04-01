/*
 *   Copyright (c) 2022 ELIPCERO
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/swpoolcontroller/pkg/strings"
	"go.uber.org/zap"
)

const (
	errAWSSaveSample = "Saving sample data in AWS dynamo repository"
)

const (
	infSaving = "Saving data sample"
)

// AWS dynamodb field
const (
	dynamoDBTableKeyName  = "id"
	dynamoDBTableTemp     = "temp"
	dynamoDBTablePH       = "ph"
	dynamoDBTableORP      = "orp"
	dynamoDBTableQuality  = "quality"
	dynamoDBTableChlorine = "chlorine"
)

// SampleData is a sample of the state of the water
type SampleData struct {
	Temp float32 `json:"temp"`
	PH   float32 `json:"ph"`
	ORP  float32 `json:"orp"`
	// Quality is judged by the expert
	Quality int `json:"quality"`
	// Chlorine is judged by the expert
	Chlorine float32 `json:"chlorine"`
}

func (s *SampleData) String() string {
	m, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	return string(m)
}

// SampleRepo defines the data repository
type SampleRepo interface {
	// Save saves the samples in the db
	Save(data SampleData) error
}

// SampleAWSDynamoRepo defines the AWS dynamo repository
type SampleAWSDynamoRepo struct {
	log       *zap.Logger
	client    *dynamodb.Client
	tableName string
}

// NewSampleAWSDynamoRepo creates the AWS dymano repository
func NewSampleAWSDynamoRepo(
	cfg aws.Config,
	log *zap.Logger,
	tableName string) *SampleAWSDynamoRepo {
	//
	return &SampleAWSDynamoRepo{
		log:       log,
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}
}

// save saves the samples data into AWS dynamo repository
func (s *SampleAWSDynamoRepo) Save(data SampleData) error {
	s.log.Info(infSaving, zap.String("sample", data.String()))

	_, err := s.client.PutItem(
		context.TODO(),
		&dynamodb.PutItemInput{
			TableName: aws.String(s.tableName),
			Item: map[string]types.AttributeValue{
				dynamoDBTableKeyName: &types.AttributeValueMemberS{
					Value: xid.New().String()},
				dynamoDBTableTemp: &types.AttributeValueMemberN{
					Value: fmt.Sprintf("%f", data.Temp)},
				dynamoDBTablePH: &types.AttributeValueMemberN{
					Value: fmt.Sprintf("%f", data.PH)},
				dynamoDBTableORP: &types.AttributeValueMemberN{
					Value: fmt.Sprintf("%f", data.ORP)},
				dynamoDBTableQuality: &types.AttributeValueMemberN{
					Value: strconv.Itoa(data.Quality)},
				dynamoDBTableChlorine: &types.AttributeValueMemberN{
					Value: fmt.Sprintf("%f", data.Chlorine)},
			},
		})

	if err != nil {
		return errors.Wrap(err, strings.Concat(errAWSSaveSample, s.tableName))
	}

	return nil
}
