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

type FollowModel struct {
	SVC ItemService
}

type Follow struct {
	PK     string `dynamodbav:"PK"`
	SK     string `dynamodbav:"SK"`
	GSI6PK string `dynamodbav:"GSI6PK"`
	GSI6SK string `dynamodbav:"GSI6SK"`
}

func (m FollowModel) Insert(Follower, Followee *Crab) error {
	item, err := attributevalue.MarshalMap(
		&Follow{
			PK:     fmt.Sprintf("F#%s", Follower.ID),
			SK:     fmt.Sprintf("F#%s", Followee.ID),
			GSI6PK: fmt.Sprintf("F#%s", Followee.ID),
			GSI6SK: fmt.Sprintf("F#%s", Follower.ID),
		})
	if err != nil {
		fmt.Println("ERR: ", err)
		panic(err)
	}
	notification, err := attributevalue.MarshalMap(
		&Notification{
			PK:       fmt.Sprintf("N#%s", Followee.ID),
			SK:       fmt.Sprintf("N#%s#%s#%s", Follower.ID, "F", Followee.ID), // needs to be unique enough...
			UserName: Follower.UserName,
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
	// increments how many crabs the user is following
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: Follower.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: Follower.SK,
				},
			},
			ConditionExpression: aws.String("attribute_exists(PK)"),
			TableName:           aws.String(TableName),
			UpdateExpression:    aws.String("set #following_count = #following_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#following_count": "following_count",
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
					Value: Followee.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: Followee.SK,
				},
			},
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_exists(PK)"),
			UpdateExpression:    aws.String("set #follower_count = #follower_count + :value"),
			ExpressionAttributeNames: map[string]string{
				"#follower_count": "follower_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	// notify
	tw4 := types.TransactWriteItem{
		Put: &types.Put{
			Item:                notification,
			TableName:           aws.String(TableName),
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	tItems = append(tItems, tw3)
	tItems = append(tItems, tw4)

	_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		return err
	}
	return nil

}

// Show the crabs you are following
func (m FollowModel) Show(id string) []Crab {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "F#" + id},
		},
	})
	following := make([]Crab, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var follow []Crab
		err = attributevalue.UnmarshalListOfMaps(out.Items, &following)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		following = append(following, follow...)
	}
	return following
}

func (m FollowModel) Followers(id string) []Crab {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		IndexName:              aws.String("GSI6"),
		KeyConditionExpression: aws.String("GSI6PK = :gsi6pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi6pk": &types.AttributeValueMemberS{Value: "F#" + id},
		},
	})
	followers := make([]Crab, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var follow []Crab
		err = attributevalue.UnmarshalListOfMaps(out.Items, &followers)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		followers = append(followers, follow...)
	}
	return followers
}

func (m FollowModel) Delete(Follower, Followee *Crab) error {
	tItems := make([]types.TransactWriteItem, 0)
	// delete it from the main table
	tw1 := types.TransactWriteItem{
		Delete: &types.Delete{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("F#%s", Follower.ID),
				},
				"SK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("F#%s", Followee.ID),
				},
			},
			TableName: aws.String(TableName),
		},
	}
	tw2 := types.TransactWriteItem{
		Update: &types.Update{
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{
					Value: Follower.PK,
				},
				"SK": &types.AttributeValueMemberS{
					Value: Follower.SK,
				},
			},
			TableName:        aws.String(TableName),
			UpdateExpression: aws.String("set #following_count = #following_count - :value"),
			ExpressionAttributeNames: map[string]string{
				"#following_count": "following_count",
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
					Value: fmt.Sprintf("C#%s", Followee.Email),
				},
				"SK": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("CU#%s", Followee.UserName), // need the email...
				},
			},
			TableName:        aws.String(TableName),
			UpdateExpression: aws.String("set #follower_count = #follower_count - :value"),
			ExpressionAttributeNames: map[string]string{
				"#follower_count": "follower_count",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":value": &types.AttributeValueMemberN{Value: "1"},
			},
		},
	}
	tItems = append(tItems, tw1)
	tItems = append(tItems, tw2)
	tItems = append(tItems, tw3)

	_, err := m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nErr: %v\n", err)
	}
	return err
}
