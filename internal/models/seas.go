package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"math/rand"
	"time"
)

type Cache struct {
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Molts []Molt `dynamodbav:"molts"`
}

// Sea the latest molts from all the trenches
func (m MoltModel) Sea() ([]Molt, error) {
	rand.Seed(time.Now().UnixNano())
	cache := rand.Intn(ShardSize)
	out, err := m.SVC.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("MS#%d", cache)},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("MS#%d", cache)},
		},
	})
	log.Printf("\nReading from Cache #: %d", cache)
	if err != nil {
		fmt.Errorf("ERR: %s", err)
	}
	molt := make([]Molt, 0)
	molts := out.Item["molts"]
	err = attributevalue.Unmarshal(molts, &molt)
	if err != nil {
		return nil, err
	}
	return molt, err
}

// FillSea -> Creates shards of the latest molts
func (m MoltModel) FillSea() error {
	// retrieve all deals from past day
	l := m.latest()
	fmt.Printf("Length of latest is: %d", len(l))
	for i := 0; i < ShardSize; i++ {
		c := &Cache{
			PK:    fmt.Sprintf("MS#%d", i),
			SK:    fmt.Sprintf("MS#%d", i),
			Molts: l,
		}
		item, err := attributevalue.MarshalMap(c)
		if err != nil {
			fmt.Println("ERR: ", err)
			return err
		}
		_, err = m.SVC.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(TableName),
			Item:      item,
		})
		if err != nil {
			return err
		}

	}
	return nil
}
