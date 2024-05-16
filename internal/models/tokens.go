package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"krabber.net/internal/models/validator"
	"time"
)

// Define constants for the token scope. For now we just define the scope "activation"
// but we'll add additional scopes later in the book.
const (
	ScopeActivation     = "ACTIVATION"
	ScopeAuthentication = "AUTHENTICATION"
	ScopePasswordReset  = "PASSWORD-RESET"
)

// Define a Token struct to hold the data for an individual token. This includes the
// plaintext and hashed versions of the token, associated user ID, expiry time and
// scope.
type Token struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	Plaintext    string `dynamodbav:"plaintext"`
	ByteHash     []byte `dynamodbav:"byte_hash"`
	CrabID       string `dynamodbav:"crab_id"`
	CrabEmail    string `dynamodbav:"crab_email"`
	CrabUserName string `dynamodbav:"crab_username"`
	CreatedAt    string `dynamodbav:"created_at"`
	ExpiresAt    string `dynamodbav:"expires_at"`
	TTL          string `dynamodbav:"ttl"`
	Scope        string `dynamodbav:"scope"`
}

type Auth struct {
	Plaintext string `json:"token"`
	ExpiresAt string `json:"expiry"`
}

// Define the TokenModel type.
type TokenModel struct {
	SVC ItemService
}

func (m TokenModel) Insert(token *Token) error {
	t := &Token{
		PK:           fmt.Sprintf("CT#%s", token.ByteHash),
		SK:           fmt.Sprintf("CT#%sTYPE#%s", token.ByteHash, token.Scope),
		Plaintext:    token.Plaintext,
		CrabID:       token.CrabID,
		CrabEmail:    token.CrabEmail,
		CrabUserName: token.CrabUserName,
		CreatedAt:    token.CreatedAt,
		ExpiresAt:    token.ExpiresAt,
		TTL:          token.TTL,
		Scope:        token.Scope,
	}

	item, err := attributevalue.MarshalMap(t)
	if err != nil {
		panic(err)
	}

	_, err = m.SVC.ItemTable.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(plaintext)"),
	})
	if err != nil {
		fmt.Printf("\n Insert Token ERROR : %v", err)
		return err
	}
	return nil
}

func (m TokenModel) New(c *Crab, activation string) (*Token, error) {
	token, err := generateToken(c, activation)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Get(tokenScope, tokenPlaintext string) (*Token, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	t := &Token{}
	fmt.Printf("CT HASH: %v", tokenHash)
	selectedKeys := map[string]string{
		"PK": fmt.Sprintf("CT#%s", tokenHash),
		"SK": fmt.Sprintf("CT#%sTYPE#%s", tokenHash, tokenScope),
	}
	key, err := attributevalue.MarshalMap(selectedKeys)
	data, err := m.SVC.ItemTable.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key:       key,
	})
	if err != nil {
		return t, fmt.Errorf("GetItem: %v\n", err)
	}

	if data.Item == nil {
		return t, fmt.Errorf("GetItem: Data not found.\n")
	}
	err = attributevalue.UnmarshalMap(data.Item, &t)
	if err != nil {
		return t, fmt.Errorf("UnmarshalMap: %v\n", err)
	}
	//fmt.Println("pretty get token", prettyPrint(t))

	return t, nil
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func generateToken(c *Crab, scope string) (*Token, error) {
	expiresAt := time.Now().Add(time.Hour * 24 * 3).Format(time.RFC3339)
	createdAt := time.Now().Format(time.RFC3339)
	token := &Token{
		CrabID:       c.ID,
		CrabEmail:    c.Email,
		CrabUserName: c.UserName,
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
		TTL:          fmt.Sprintf("%d", time.Now().Add(time.Hour*24*3).Unix()),
		Scope:        scope,
	}

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.ByteHash = hash[:]
	return token, nil
}

// Check that the plaintext token has been provided and is exactly 26 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
//func (m TokenModel) DeleteAllActivation(scope string, userID int64) error {
//	query := `
//       DELETE FROM tokens
//       WHERE scope = $1 AND user_id = $2`
//
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	defer cancel()
//
//	_, err := m.DB.ExecContext(ctx, query, scope, userID)
//	return err
//}
