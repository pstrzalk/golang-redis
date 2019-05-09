package lib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"sync"
)

type dynamodbStore struct {
	client *dynamodb.DynamoDB
	lock   *sync.RWMutex
}

type tableItem struct {
	Key   string
	Value string
}

// NewInMemoryStore returns an in-memory implementation of Store.
func NewDynamodbStore() Store {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &dynamodbStore{
		client: dynamodb.New(sess),
		lock:   new(sync.RWMutex),
	}
}

func (s *dynamodbStore) Get(key string) (value string, found bool, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	input := &dynamodb.GetItemInput{
		TableName: aws.String("Redis"),
		Key: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
		},
	}

	result, err := s.client.GetItem(input)
	if err != nil {
		return "", false, err
	}
	if result.Item == nil {
		return "", false, nil
	}

	item := tableItem{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	if err != nil {
		return "", false, err
	}

	return item.Value, true, nil
}

func (s *dynamodbStore) Set(key string, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	input := &dynamodb.PutItemInput{
		TableName: aws.String("Redis"),
		Item: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
			"Value": {
				S: aws.String(value),
			},
		},
	}

	_, err := s.client.PutItem(input)
	return err
}

func (s *dynamodbStore) Delete(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String("Redis"),
		Key: map[string]*dynamodb.AttributeValue{
			"Key": {
				S: aws.String(key),
			},
		},
	}

	_, err := s.client.DeleteItem(input)
	return err
}
