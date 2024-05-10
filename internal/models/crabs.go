package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"krabber.net/internal/models/validator"
	"strings"
	"time"
)

// Define a custom ErrDuplicateEmail error.
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type CrabModel struct {
	SVC ItemService
}

// Declare a new AnonymousUser variable.
var AnonymousCrab = &Crab{}

type Crab struct {
	ID             string   `dynamodbav:"id"`
	PK             string   `dynamodbav:"PK"`
	SK             string   `dynamodbav:"SK"`
	GSI1PK         string   `dynamodbav:"GSI1PK"`
	GSI1SK         string   `dynamodbav:"GSI1SK"`
	GSI2PK         string   `dynamodbav:"GSI2PK"`
	GSI2SK         string   `dynamodbav:"GSI2SK"`
	Activated      bool     `dynamodbav:"activated"`
	Avatar         string   `dynamodbav:"avatar"`
	Banned         bool     `dynamodbav:"banned"`
	Banner         string   `dynamodbav:"banner"`
	Created        string   `dynamodbav:"created"`
	Description    string   `dynamodbav:"description"`
	Display        string   `dynamodbav:"display"`
	Deleted        bool     `dynamodbav:"deleted"`
	Email          string   `dynamodbav:"email"`
	FollowerCount  int      `dynamodbav:"follower_count"`
	FollowingCount int      `dynamodbav:"following_count"`
	LikeCount      int      `dynamodbav:"like_count"`
	Moderator      bool     `dynamodbav:"moderator"`
	MoltCount      int      `dynamodbav:"molt_count"`
	UserName       string   `dynamodbav:"user_name"`
	Password       password `dynamodbav:"password"`
	PasswordHash   []byte   `dynamodbav:"password_hash"`
	Website        string   `dynamodbav:"website"`
	Verified       bool     `dynamodbav:"verified"`
}

// Check if a User instance is the AnonymousUser.
func (c *Crab) IsAnonymous() bool {
	return c == AnonymousCrab
}

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user. The plaintext field is a *pointer* to a string,
// so that we're able to distinguish between a plaintext password not being present in
// the struct at all, versus a plaintext password which is the empty string "".
type password struct {
	plaintext *string
	hash      []byte
}

func (m CrabModel) Show() ([]Crab, error) {
	p := dynamodb.NewScanPaginator(m.SVC.ItemTable, &dynamodb.ScanInput{
		TableName: aws.String(TableName),
		Limit:     aws.Int32(PageSize),
		IndexName: aws.String("GSI2"),
	})
	var items []Crab
	for p.HasMorePages() {
		out, err := p.NextPage(context.TODO())
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		var pItems []Crab
		err = attributevalue.UnmarshalListOfMaps(out.Items, &pItems)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			panic(err)
		}
		items = append(items, pItems...)

	}
	return items, nil
}

// Insert - creates user record in table
func (m CrabModel) Insert(crab *Crab) (*Crab, error) {
	id := uuid.New().String()
	c := &Crab{
		ID:           id,
		PK:           fmt.Sprintf("C#%s", crab.Email),
		SK:           fmt.Sprintf("CU#%s", strings.ToLower(crab.UserName)), // all will be stored as lower case to avoid dupes like Test and test
		GSI1PK:       fmt.Sprintf("C#%s", crab.Email),
		GSI1SK:       fmt.Sprintf("C#%s", crab.Email),
		GSI2PK:       fmt.Sprintf("ID#%s", id),
		GSI2SK:       fmt.Sprintf("ID#%s", id),
		Created:      fmt.Sprintf(time.Now().Format(time.RFC3339)),
		Email:        crab.Email,
		PasswordHash: crab.Password.hash,
		UserName:     crab.UserName, // it is now permanent...
	}

	item, err := attributevalue.MarshalMap(c)
	if err != nil {
		panic(err)
	}
	_, err = m.SVC.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"), // one email to one user...
	})
	if err != nil {
		fmt.Printf("\n\nERROR Inserting Crab Item: %v", item)
		fmt.Printf("\n\nERROR Inserting Crab: %v", err)
	}
	return c, err
}

func (m CrabModel) Activate(crab *Crab) error {
	fmt.Println("update crab ", prettyPrint(crab))
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: crab.PK},
			"SK": &types.AttributeValueMemberS{Value: crab.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "activated", "activated")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "activated"): &types.AttributeValueMemberBOOL{Value: crab.Activated},
		},
	})

	if err != nil {
		panic(err)
	}

	return nil
}

func (m CrabModel) ResetPassword(crab *Crab) error {
	_, err := m.SVC.ItemTable.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: crab.PK},
			"SK": &types.AttributeValueMemberS{Value: crab.SK},
		},
		UpdateExpression: aws.String(fmt.Sprintf("set %s = :%s", "password_hash", "password_hash")),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", "password_hash"): &types.AttributeValueMemberB{Value: crab.Password.hash},
		},
	})

	if err != nil {
		panic(err)
	}

	return nil
}

func (m CrabModel) Exists(id string) (bool, error) {
	c := make([]Crab, 0)
	ut, err := m.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("C#%s", id)},
		},
	})
	if err != nil {
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &c)
	if err != nil {
		return false, err
	}
	return true, nil

}

// Change this to be by Email -> Add e-mail to param store ->
func (m CrabModel) ByID(id string) (*Crab, error) {
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

// Change this to be by Email -> Add e-mail to param store ->
func (m CrabModel) ByEmailForID(email string) (*Crab, error) {
	c := make([]Crab, 0)
	ut, err := m.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi2pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("C#%s", email)},
		},
	})
	if err != nil {
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &c)
	if err != nil {
		fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	fmt.Println("HERES BY ID", prettyPrint(&c[0]))
	return &c[0], nil
}

func (m CrabModel) ByEmail(email string) (*Crab, error) {
	c := make([]Crab, 0)
	ut, err := m.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(TableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi1pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("C#%s", email)},
		},
	})
	if err != nil {
		panic(err)
	}
	err = attributevalue.UnmarshalListOfMaps(ut.Items, &c)
	if err != nil {
		fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	fmt.Println("heres crab", prettyPrint(&c[0]))
	return &c[0], nil
}

func (m CrabModel) ByToken(t *Token) (*Crab, error) {
	c := &Crab{}
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf("C#%s", t.CrabEmail),
		"SK": fmt.Sprintf("CU#%s", t.CrabUserName),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)
	data, err := m.SVC.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       key,
	})
	if err != nil {
		return c, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return c, fmt.Errorf("GetItem: Data not found.\n")
	}
	err = attributevalue.UnmarshalMap(data.Item, &c)
	if err != nil {
		return c, fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	return c, nil
}

// The Set() method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func Equal(password string, passwordHash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateCrab(v *validator.Validator, crab *Crab) {
	v.Check(crab.UserName != "", "name", "must be provided")
	v.Check(len(crab.UserName) <= 500, "name", "must not be more than 500 bytes long")

	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, crab.Email)

	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if crab.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *crab.Password.plaintext)
	}

	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase (probably because we forgot to set a password for the user). It's a
	// useful sanity check to include here, but it's not a problem with the data
	// provided by the client. So rather than adding an error to the validation map we
	// raise a panic instead.
	if crab.Password.hash == nil {
		panic("missing password hash for user")
	}
}

//
//// Delete - removes a user from Lobsterer DB & Cognito
//func (u *User) Delete(svc ItemService, tablename string) error {
//	return nil
//}

// Exists - checks if username is already taken
//func Exists(name string, svc ItemService, tablename string) (bool, error) {
//	selectedKeys := map[string]string{
//		"PK": fmt.Sprintf(PKFormat, name),
//		"SK": fmt.Sprintf(SKFormat, name),
//	}
//	key, err := attributevalue.MarshalMap(selectedKeys)
//
//	data, err := svc.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
//		TableName: aws.String(tablename),
//		Key:       key,
//	},
//	)
//	if err != nil {
//		return false, fmt.Errorf("GetItem: %v\n", err)
//	}

//	if models.Item == nil {
//		return false, fmt.Errorf("GetItem: Data not found.\n")
//	}
//
//	return true, nil
//}

//func (c CrabModel) GetByName(name string) (Crab, error) {
//	crab := make([]Crab, 0)
//	ut, err := c.SVC.ItemTable.Query(context.TODO(), &dynamodb.QueryInput{
//		TableName:              aws.String(TableName),
//		IndexName:              aws.String("GSI1"),
//		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
//		ExpressionAttributeValues: map[string]types.AttributeValue{
//			":gsi1pk": &types.AttributeValueMemberS{Value: name},
//		},
//	})
//	if err != nil {
//		panic(err)
//	}
//	err = attributevalue.UnmarshalListOfMaps(ut.Items, &crab)
//	if err != nil {
//		fmt.Errorf("UnmarshalMap: %v\n", err)
//	}
//
//	return crab[0], nil
//}
