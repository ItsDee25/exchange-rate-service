package builders

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type env struct {
	HttpClient   *http.Client
	DynamoClient *dynamodb.Client
}

func NewEnv() *env {
	return &env{}
}

func (e *env) WithDynamoClient(c *dynamodb.Client) *env {
	e.DynamoClient = c
	return e
}

func (e *env) WithHTTPClient(c *http.Client) *env {
	e.HttpClient = c
	return e
}
