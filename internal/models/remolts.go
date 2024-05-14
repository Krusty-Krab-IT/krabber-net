package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// show a crab's remolts
func (m MoltModel) ShowReMolts(id string) ([]Molt, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		FilterExpression:       aws.String("remolt <> :remolt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "M#" + id},
			":remolt":  &types.AttributeValueMemberBOOL{Value: false},
		},
		ScanIndexForward: aws.Bool(false),
	})
	molts := make([]Molt, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var molt []Molt
		err = attributevalue.UnmarshalListOfMaps(out.Items, &molt)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		molts = append(molts, molt...)
	}
	return molts, nil
}

// This crab remolts another molt
func (m MoltModel) ReMolt(c *Crab, other, molt *Molt) error {
	newMolt := &Molt{
		ID:      molt.ID,
		PK:      molt.PK,
		SK:      molt.SK,
		GSI3PK:  molt.GSI3PK,
		GSI3SK:  molt.GSI3SK,
		GSI5PK:  molt.GSI5PK,
		GSI5SK:  molt.GSI5SK,
		Author:  molt.Author,
		Content: molt.Content,
		Remolt:  true,
		Deleted: molt.Deleted,
	}

	item, err := attributevalue.MarshalMap(newMolt)

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
					Value: c.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: c.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #molt_count = #molt_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#molt_count": "molt_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tw3 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: other.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: other.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #remolt_count = #remolt_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#remolt_count": "remolt_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	// notify

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

func (m MoltModel) DeleteReMolt(molt *Molt) error {
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: molt.PK},
			"SK": &types.AttributeValueMemberS{Value: molt.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "remolt", "remolt")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "remolt"): &types.AttributeValueMemberBOOL{Value: molt.Remolt},
		},
	})

	if err != nil {
		panic(err)
	}

	return nil
}
