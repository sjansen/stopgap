// +build integration

package storage_test

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/require"

	"github.com/sjansen/stopgap/internal/rqx"
	"github.com/sjansen/stopgap/internal/storage"
)

func createClient() *dynamodb.DynamoDB {
	endpoint := os.Getenv("DYNAMOSTORE_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	creds := credentials.NewStaticCredentials("id", "secret", "token")
	sess := session.Must(session.NewSession())
	return dynamodb.New(
		sess,
		aws.NewConfig().
			WithCredentials(creds).
			WithRegion("us-west-2").
			WithEndpoint(endpoint),
	)
}

func randomString() string {
	rand.Seed(time.Now().Unix())
	bytes := make([]byte, 10)
	for i := range bytes {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}

func TestDynamoDBLocal(t *testing.T) {
	require := require.New(t)

	svc := createClient()
	require.NotNil(svc)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := svc.ListTablesWithContext(ctx, &dynamodb.ListTablesInput{})
	require.NoError(err)
}

func TestCreateTable(t *testing.T) {
	require := require.New(t)

	svc := createClient()
	require.NotNil(svc)

	store := storage.New(svc)

	// first time: created
	err := store.CreateTable()
	require.NoError(err)

	// second time: noop
	err = store.CreateTable()
	require.NoError(err)
}

func TestDynamoStore(t *testing.T) {
	require := require.New(t)

	svc := createClient()
	require.NotNil(svc)

	store := storage.New(svc)
	require.NotNil(store)

	name := randomString()
	user := rqx.User{
		Name:    "Test User",
		SlackID: "UFoo42",
	}
	rqx := &rqx.RequestContext{
		Ctx: context.TODO(),
		Client: rqx.Client{
			Type: "test case",
		},
		EUser: user,
		RUser: user,
	}

	err := store.CreateTable()
	require.NoError(err)

	err = store.CreateMutex(rqx, name, "a test mutex")
	require.NoError(err)

	m, err := store.GetMutex(true, name)
	require.NoError(err)
	require.False(m.Locked)

	err = store.LockMutex(rqx, name, "first attempt")
	require.NoError(err)

	m, err = store.GetMutex(true, name)
	require.NoError(err)
	require.True(m.Locked)
	require.Equal(user.SlackID, m.LockedBy)

	err = store.LockMutex(rqx, name, "second attempt")
	require.Error(err)

	err = store.UnlockMutex(rqx, name)
	require.NoError(err)

	m, err = store.GetMutex(true, name)
	require.NoError(err)
	require.False(m.Locked)
	require.Empty(m.LockedBy)
}
