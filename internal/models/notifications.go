package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

const (
	ScopeLike     = "L"
	ScopeRemolt   = "R"
	ScopeComment  = "C"
	ScopeMention  = "M"
	ScopeFollower = "F"
)

type NotificationModel struct {
	SVC ItemService
}

type Notification struct {
	PK    string `dynamodbav:"PK"`
	SK    string `dynamodbav:"SK"`
	Read  bool   `dynamodbav:"read"`
	Scope string `dynamodbav:"scope"`
	TTL   string `dynamodbav:"ttl"` // make them expire after 1 week so that the dynamodb table stays slim...
}

// add a notification for the other crab that the logged in crab interacted with
func (m NotificationModel) Insert(ownersID, moltID, otherID string) error {
	// if id contains M#, MC#, F# then strip the id from it
	// Remolt - like, remolt in general
	// Comment - like, remolt in general
	// Follower - new follower
	// Mentions -to be built
	// need two pieces of data (1) the Crab owner's ID, and the molt's ID
	n := &Notification{
		PK:    fmt.Sprintf("N#%s", ownersID),                      // ownersID -> get all of my notifications
		SK:    fmt.Sprintf("N#%s#%s#%s", moltID, "TYPE", otherID), // needs to be unique enough...
		Read:  false,
		TTL:   fmt.Sprintf("%d", time.Now().Add(time.Hour*24*7).Unix()), // delete notifs in a week to keep table smaller
		Scope: "LIKE",
	}

	notification, err := attributevalue.MarshalMap(n)
	if err != nil {
		fmt.Println("ERR marshalling: ", err)
		panic(err)
	}
	tItems := make([]types.TransactWriteItem, 0)
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                notification,
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		},
	}

	tItems = append(tItems, tw1)

	_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
		return err
	}
	return nil
}

// show all the notifications for the logged in crab
func (m NotificationModel) Show(crabID string) ([]Notification, error) {
	// get all the ones with PK: N#<my-crab-id> -> returns all of them and they'll have the PK,SK so updates should work
	// fine same as showMolts

	// for each item after showing change the field to true for read
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "mock"},
			"SK": &types.AttributeValueMemberS{Value: "mock"},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "read", "read")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "read"): &types.AttributeValueMemberBOOL{Value: true},
		},
	})

	if err != nil {
		panic(err)
	}
	return nil, nil
}
