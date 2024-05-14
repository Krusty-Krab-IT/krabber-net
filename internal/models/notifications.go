package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	Scope  string `dynamodbav:"scope"`
	TTL    string `dynamodbav:"ttl"` // make them expire after 1 week so that the dynamodb table stays slim...
	Viewed bool   `dynamodbav:"viewed"`
}

// show all the notifications for the logged in crab
func (m NotificationModel) Show(crabID string) ([]Notification, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		FilterExpression:       aws.String("viewed <> :viewed"), // read true or false...?
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "N#" + crabID},
			":viewed":  &types.AttributeValueMemberBOOL{Value: true},
		},
		ScanIndexForward: aws.Bool(false),
	})
	molts := make([]Notification, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var molt []Notification
		err = attributevalue.UnmarshalListOfMaps(out.Items, &molt)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		molts = append(molts, molt...)
	}
	return molts, nil
}
