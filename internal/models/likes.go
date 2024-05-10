package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Like struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI7PK string `dynamodbav:"GSI7PK"`
	GSI7SK string `dynamodbav:"GSI7SK"`
}

type LikesModel struct {
	SVC ItemService
}

// Insert a Like record
func (m LikesModel) Insert(cid string, molt *Molt) error {
	item, err := attributevalue.MarshalMap(
		&Like{
			PK:     fmt.Sprintf("L#%s", cid),
			SK:     fmt.Sprintf("L#%s", molt.ID),
			GSI7PK: fmt.Sprintf("L#%s", molt.ID),
			GSI7SK: fmt.Sprintf("L#%s", cid),
		})
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	tItems := make([]types.TransactWriteItem, 0)
	tw1 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                item,
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		},
	}
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: molt.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: molt.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #like_count = #like_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#like_count": "like_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)

	_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v", err)
	}
	return err
}

// Show a crab's likes by id
func (m LikesModel) Show(id string) []Molt {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "L#" + id},
		},
	})
	molts := make([]Molt, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var molt []Molt
		err = attributevalue.UnmarshalListOfMaps(out.Items, &molts)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		molts = append(molts, molt...)
	}
	return molts
}

// Delete a like record based on current logged in crab
func (m LikesModel) Delete(c *Crab, molt *Molt) error {
	tItems := make([]types.TransactWriteItem, 0)
	// delete it from the main table
	tw1 := types.TransactWriteItem{
		Delete: &types.Delete{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("L#%s", c.ID),
				},
				"SK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("L#%s", molt.ID),
				},
			},
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_exists(PK)"),
		},
	}
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: molt.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: molt.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #like_count = #like_count - :value"),
			ExpressionAttributeNames: map[string]string{
				"#like_count": "like_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	_, err := m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v\n", err)
	}
	return err
}

// GET - Crabs that have liked a specific molt
func (m LikesModel) On(moltID string) ([]Like, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI7"),
		Limit:                  aws.Int32(5),
		KeyConditionExpression: aws.String("GSI7PK = :gsi7pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi7pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("L#%s", moltID)},
		},
	})
	// update this for pagination
	likes := make([]Like, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var like []Like
		err = attributevalue.UnmarshalListOfMaps(out.Items, &like)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		likes = append(likes, like...)
	}
	return likes, nil
}
