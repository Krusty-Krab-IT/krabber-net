package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TrenchModel struct {
	SVC ItemService
}

// For the Id of the crab get the followers -> maybe once crab authenticates call the trench method...?
// then for each follower POST to their Trench -> each time I molt -> post to my followers
// type trench struct {
// PK: T%s, OtherCrabID
// SK: T%s, MyMoltID
// Fanned: false
// }
// Naive algorithm for a hobby project
// do POST /v1/trench with my special key
// returns all of my molts that haven't been fanned out
// then for each one do fan out -> write to each of my followers trench
// update the molt to fanned:true when done
type Trench struct {
	PK string `dynamodbav:"PK"` // PK: T%s, OtherCrabID
	SK string `dynamodbav:"SK"` // //SK: T%s, MyMoltID
}

// called after a molt is successfully inserted
func (m TrenchModel) Insert(crabs []Crab, molt *Molt) error {
	for _, c := range crabs {
		//fmt.Printf("%+v\n", c)
		//fmt.Printf("follower %v", c.PK[2:])
		item, err := attributevalue.MarshalMap(
			&Trench{
				PK: fmt.Sprintf("T#%s", c.PK[2:]), // one crab will have many -> sort by the SK
				SK: fmt.Sprintf("T#%s", molt.ID),
			})
		if err != nil {
			fmt.Println("ERR: ", err)
			panic(err)
		}
		tItems := make([]types.TransactWriteItem, 0)
		tw1 := types.TransactWriteItem{
			Put: &types.Put{
				Item:      item,
				TableName: aws.String(TableName),
			},
		}
		tItems = append(tItems, tw1)
		// Worried about this part
		_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
			TransactItems: tItems,
		})
		if err != nil {
			fmt.Printf("\nErr: %v", err)
			return err
		}
	}
	return nil
}

// Show - Get the crab trench feed for current crab
func (m TrenchModel) Get(id string) ([]Trench, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "T#" + id},
		},
		ScanIndexForward: aws.Bool(false),
	})
	trenches := make([]Trench, 0)
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var trench []Trench
		err = attributevalue.UnmarshalListOfMaps(out.Items, &trench)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		trenches = append(trenches, trench...)
	}
	return trenches, nil
}

func (m MoltModel) GetTrenchMolts(trench []Trench) []Molt {
	molts := make([]Molt, 0)
	for _, t := range trench {
		molt, err := m.ByID(t.SK[2:]) // delete the time stamp not needed in insert
		if err != nil {
			fmt.Printf("err %s", err)
		}
		molts = append(molts, *molt)
	}
	return molts
}

// Build - For each SK[:2] -> do get by id
// return slice of molts
