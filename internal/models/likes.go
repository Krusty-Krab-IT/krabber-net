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

type Like struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI7PK string `dynamodbav:"GSI7PK"`
	GSI7SK string `dynamodbav:"GSI7SK"`
}

type LikesModel struct {
	SVC ItemService
}

// Change this to be by Email -> Add e-mail to param store ->
func (m LikesModel) ByID(id string) (*Crab, error) {
	c := make([]Crab, 0)
	ut, err := m.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("ID#%s", id)},
		},
	})
	if err != nil {
		fmt.Printf("BYID ERROR %v", err)
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &c)
	if err != nil {
		fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	fmt.Println("HERES BY ID", prettyPrint(&c[0]))
	return &c[0], nil
}

// Insert a Like record
func (m LikesModel) Insert(cid string, molt *Molt) error {
	c, err := m.ByID(cid)
	if err != nil {
		fmt.Printf("ERR getting crab that liked the molt... %v", err)
	}
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
	ownerID := molt.PK[2:]
	fmt.Println("OWNER ID", ownerID)
	notification, err := attributevalue.MarshalMap(
		&Notification{
			PK:       fmt.Sprintf("N#%s", ownerID),                     // slice the crabs id
			SK:       fmt.Sprintf("N#%s#%s#%s", ownerID, "L", molt.ID), // needs to be unique enough...
			UserName: c.UserName,                                       // need to fetch crab
			Viewed:   false,
			TTL:      fmt.Sprintf("%d", time.Now().Add(time.Hour*24*7).Unix()), // delete notifs in a week to keep table smaller
		})
	if err != nil {
		fmt.Println("Notification ERR: ", err)
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
	tw3 := types.TransactWriteItem{
		Put: &types.Put{
			Item:      notification,
			TableName: aws.String(TableName),
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	tItems = append(tItems, tw3)

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
