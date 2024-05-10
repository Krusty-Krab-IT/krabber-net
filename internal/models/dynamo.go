package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ItemService struct {
	ItemTable *dynamodb.Client
}
