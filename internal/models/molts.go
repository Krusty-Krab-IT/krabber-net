package models

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"krabber.net/internal/models/validator"
	"time"
)

type MoltModel struct {
	SVC ItemService
}

type Molt struct {
	ID           string    `dynamodbav:"id"`
	Comments     []Comment // added now whenever retrieving molts will also get the comments associated with it
	Likes        []Like
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	GSI3PK       string `dynamodbav:"GSI3PK"`
	GSI3SK       string `dynamodbav:"GSI3SK"`
	GSI5PK       string `dynamodbav:"GSI5PK"`
	GSI5SK       string `dynamodbav:"GSI5SK"`
	Author       string `dynamodbav:"author"`
	CommentCount int    `dynamodbav:"comment_count"`
	Content      string `dynamodbav:"content"`
	Deleted      bool   `dynamodbav:"deleted"`
	LikeCount    int    `dynamodbav:"like_count"`
	Remolt       bool   `dynamodbav:"remolt"`
	RemoltCount  int    `dynamodbav:"remolt_count"`
	Url          string `dynamodbav:"url"`
}

// Insert a molt into the db
func (m MoltModel) Insert(molt *Molt) error {
	item, err := attributevalue.MarshalMap(
		&Molt{
			ID:      molt.ID,
			PK:      molt.PK,
			SK:      molt.SK,
			GSI3PK:  molt.GSI3PK,
			GSI3SK:  molt.GSI3SK,
			GSI5PK:  molt.GSI5PK,
			GSI5SK:  molt.GSI5SK,
			Author:  molt.Author,
			Content: molt.Content,
			Deleted: molt.Deleted,
		})
	if err != nil {
		fmt.Println("ERROR MARSHALLING: ", err)
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
	tItems = append(tItems, tw1)
	// Worried about this part
	_, err = m.SVC.ItemTable.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: tItems,
	})

	if err != nil {
		fmt.Printf("\nERROR INSERTING: %v", err)
	}
	return err
}

// By ID for individual viewing
func (m MoltModel) ByID(id string) (*Molt, error) {
	molt := make([]Molt, 0)
	ut, err := m.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI5"),
		FilterExpression:       aws.String("deleted <> :deleted"),
		KeyConditionExpression: aws.String("GSI5PK = :gsi5pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi5pk":  &types.AttributeValueMemberS{Value: fmt.Sprintf("M#%s", id)},
			":deleted": &types.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &molt)
	if err != nil {
		fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	return &molt[0], nil
}

// Get a molt by Author and ID for updates
func (m MoltModel) Get(author, id string) (*Molt, error) {
	molt := Molt{}
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf("M#%s", author),
		"SK": fmt.Sprintf("M#%s#%s", author, id),
	}

	key, err := attributevalue.MarshalMap(selectedKeys)

	data, err := m.SVC.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       key,
	})
	if err != nil {
		return &molt, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return &molt, fmt.Errorf("GetItem: Data not found.\n")
	}
	err = attributevalue.UnmarshalMap(data.Item, &molt)
	if err != nil {
		return &molt, fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	// if molt exists show it
	if !molt.Deleted {
		return &molt, nil
	}
	// if it is deleted then deletedError TODO: some point add this...
	return nil, nil
}

// Update a molt that the crab has molted
func (m MoltModel) Update(molt *Molt) error {
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: molt.PK},
			"SK": &types.AttributeValueMemberS{Value: molt.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "content", "content")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "content"): &types.AttributeValueMemberS{Value: molt.Content},
		},
	})

	if err != nil {
		panic(err)
	}

	return nil
}

// Show a crab's molts if not
func (m MoltModel) Show(id string) ([]Molt, error) {
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		KeyConditionExpression: aws.String("PK = :hashKey"),
		FilterExpression:       aws.String("deleted <> :deleted"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "M#" + id},
			":deleted": &types.AttributeValueMemberBOOL{Value: true},
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

// Delete a molt that a crab has molted
func (m MoltModel) Delete(molt *Molt) error {
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: molt.PK},
			"SK": &types.AttributeValueMemberS{Value: molt.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "deleted", "deleted")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "deleted"): &types.AttributeValueMemberBOOL{Value: molt.Deleted},
		},
	})

	if err != nil {
		panic(err)
	}

	return nil
}

// Latest - returns all the molts from past day
func (m MoltModel) latest() []Molt {
	now := time.Now()
	y, mnth, d := now.Date()
	p := dynamodb.NewQueryPaginator(m.SVC.ItemTable, &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		Limit:                  aws.Int32(PageSize),
		IndexName:              aws.String("GSI3"),
		KeyConditionExpression: aws.String("GSI3PK = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: "M#" + fmt.Sprintf("%d-%d-%d", y, int(mnth), d)},
		},
		ScanIndexForward: aws.Bool(false),
	})
	var items []Molt
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var pItems []Molt
		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		items = append(items, pItems...)

	}
	return items
}

func ValidateMolt(v *validator.Validator, molt *Molt) {
	v.Check(molt.Author != "", "author", "must be provided")
	v.Check(len(molt.Content) <= 200, "content", "must not be more than 200 bytes long")
}
