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
	"encoding/csv"
	"encoding/json"
	"os"

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
	errOpenSample    = "Opening sample data: "
	errWritingFile   = "Writing sample data: "
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
	Temp string `json:"temp"`
	PH   string `json:"ph"`
	ORP  string `json:"orp"`
	// Quality is judged by the expert
	Quality string `json:"quality"`
	// Chlorine is judged by the expert
	Chlorine string `json:"chlorine"`
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
				dynamoDBTableTemp: &types.AttributeValueMemberS{
					Value: data.Temp},
				dynamoDBTablePH: &types.AttributeValueMemberS{
					Value: data.PH},
				dynamoDBTableORP: &types.AttributeValueMemberS{
					Value: data.ORP},
				dynamoDBTableQuality: &types.AttributeValueMemberS{
					Value: data.Quality},
				dynamoDBTableChlorine: &types.AttributeValueMemberS{
					Value: data.Chlorine},
			},
		})

	if err != nil {
		return errors.Wrap(err, strings.Concat(errAWSSaveSample, s.tableName))
	}

	return nil
}

// SampleFileRepo defines the file repository
type SampleFileRepo struct {
	Log      *zap.Logger
	FileName string
}

// save saves the samples data into AWS dynamo repository
func (s *SampleFileRepo) Save(data SampleData) error {
	s.Log.Info(
		infSaving,
		zap.String("sample", data.String()),
		zap.String("file", s.FileName))

	file, err := os.OpenFile(
		s.FileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644)

	if err != nil {
		return errors.Wrap(err, strings.Concat(errOpenSample, s.FileName))
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{data.Temp, data.PH, data.ORP, data.Chlorine, data.Quality}
	if err := writer.Write(record); err != nil {
		return errors.Wrap(err, strings.Concat(errWritingFile, s.FileName))
	}

	return nil
}
