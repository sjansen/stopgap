package storage

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"

	"github.com/sjansen/stopgap/internal/rqx"
)

// DefaultTableName is used when a more specific name isn't provided.
const DefaultTableName = "stopgap"

// ErrDeleteInProgress is returned when table creation fails because
// a table with the same name was recently deleted.
var ErrDeleteInProgress = errors.New("table deletion in progress")

// ErrCreateTimedOut is returned when table creation takes too long.
var ErrCreateTimedOut = errors.New("timed out waiting for table creation")

// DynamoStore stores mutex data in DynamoDB.
type DynamoStore struct {
	svc   *dynamodb.DynamoDB
	table *string
}

// Mutex can be used to coordinate access to shared resources.
type Mutex struct {
	Version     int64
	Description string
	Locked      bool
	LockedBy    string
	Message     string
}

// New creates a DynamoStore instance using default values.
func New(svc *dynamodb.DynamoDB) *DynamoStore {
	return NewWithTableName(svc, DefaultTableName)
}

// NewWithTableName create a DynamoStore instance, overriding the default
// table name.
func NewWithTableName(svc *dynamodb.DynamoDB, table string) *DynamoStore {
	return &DynamoStore{
		svc:   svc,
		table: aws.String(table),
	}
}

func mutexEntityID(name string) string {
	return "mutex:" + name
}

// CreateMutex adds the named mutex.
func (s *DynamoStore) CreateMutex(rqx *rqx.RequestContext, name, description string) error {
	id := mutexEntityID(name)

	t := &writeTransaction{}
	err := t.addPut(&dynamodb.Put{
		TableName: s.table,
		Item: map[string]*dynamodb.AttributeValue{
			"entity":      {S: aws.String(id)},
			"revision":    {N: aws.String("0")},
			"entity_type": {S: aws.String("mutex")},
			"version":     {N: aws.String("1")},
			"description": {S: aws.String(description)},
			"summary": {M: map[string]*dynamodb.AttributeValue{
				"locked": {BOOL: aws.Bool(false)},
			}},
		},
	}).addEvent(rqx, s.table, id, 1,
		"mutex-created",
		map[string]string{
			"description": description,
		},
	)
	if err != nil {
		return err
	}

	return t.exec(s.svc)
}

// GetMutex returns the data for a given mutex from the DynamoStore instance.
func (s *DynamoStore) GetMutex(name string, consistent bool) (*Mutex, error) {
	id := mutexEntityID(name)
	item, err := s.getMutex(id, "version, summary", consistent)
	if err != nil {
		return nil, err
	}

	m := &Mutex{
		Version:  item.Version,
		Locked:   item.Summary.Locked,
		LockedBy: item.Summary.LockedBy,
		Message:  item.Summary.Message,
	}
	return m, nil
}

// LockMutex locks the named mutex.
func (s *DynamoStore) LockMutex(rqx *rqx.RequestContext, name, message string) error {
	id := mutexEntityID(name)
	item, err := s.getMutex(id, "version", true)
	if err != nil {
		return err
	}

	t := &writeTransaction{}
	version := item.Version + 1
	err = t.addUpdate(&dynamodb.Update{
		TableName: s.table,
		Key: map[string]*dynamodb.AttributeValue{
			"entity":   {S: aws.String(id)},
			"revision": {N: aws.String("0")},
		},
		ConditionExpression: aws.String(
			"summary.locked <> :locked",
		),
		UpdateExpression: aws.String(`
			SET summary.locked = :locked,
			    summary.locked_by = :locked_by,
			    summary.message = :message,
			    version = :version
		`),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":locked":    {BOOL: aws.Bool(true)},
			":locked_by": {S: aws.String(rqx.EUser.SlackID)},
			":message":   {S: aws.String(message)},
			":version": {N: aws.String(
				strconv.FormatInt(version, 10),
			)},
		},
	}).addEvent(rqx, s.table, id, version,
		"mutex-locked",
		map[string]string{
			"message": message,
		},
	)
	if err != nil {
		return err
	}

	return t.exec(s.svc)
}

// UnlockMutex unlocks the named mutex.
func (s *DynamoStore) UnlockMutex(rqx *rqx.RequestContext, name string) error {
	id := mutexEntityID(name)
	item, err := s.getMutex(id, "version", true)
	if err != nil {
		return err
	}

	t := &writeTransaction{}
	version := item.Version + 1
	err = t.addUpdate(&dynamodb.Update{
		TableName: s.table,
		Key: map[string]*dynamodb.AttributeValue{
			"entity":   {S: aws.String(id)},
			"revision": {N: aws.String("0")},
		},
		ConditionExpression: aws.String(
			"summary.locked <> :locked",
		),
		UpdateExpression: aws.String(`
			SET summary.locked = :locked,
			    version = :version
			REMOVE summary.locked_by,
			       summary.message
		`),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":locked": {BOOL: aws.Bool(false)},
			":version": {N: aws.String(
				strconv.FormatInt(version, 10),
			)},
		},
	}).addEvent(rqx, s.table, id, version,
		"mutex-unlocked",
		map[string]string{},
	)
	if err != nil {
		return err
	}

	return t.exec(s.svc)
}

// CreateTable creates the DynamoStore table, if it doesn't already exist.
// This is only intended as a convenience function to make development and
// testing easier. It is not intended for use in production.
func (s *DynamoStore) CreateTable() error {
	if ok, err := s.checkForTable(); err != nil {
		return err
	} else if ok {
		return nil
	}
	if err := s.createTable(); err != nil {
		return err
	}
	if err := s.waitForTable(); err != nil {
		return err
	}
	return s.updateTTL()
}

func (s *DynamoStore) checkForTable() (bool, error) {
	describeTable := &dynamodb.DescribeTableInput{
		TableName: s.table,
	}
	result, err := s.svc.DescribeTable(describeTable)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return false, nil
			}
		}
		return false, err
	}
	switch status := aws.StringValue(result.Table.TableStatus); status {
	case "CREATING":
		return true, s.waitForTable()
	case "DELETING":
		return false, ErrDeleteInProgress
	case "ACTIVE", "UPDATING":
		return true, nil
	default:
		return false, errors.New("unrecognized table status: " + status)
	}
}

func (s *DynamoStore) createTable() error {
	createTable := &dynamodb.CreateTableInput{
		BillingMode: aws.String("PAY_PER_REQUEST"),
		TableName:   s.table,
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("entity"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("revision"),
				KeyType:       aws.String("RANGE"),
			},
		},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("entity"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("revision"),
				AttributeType: aws.String("N"),
			},
		},
	}
	_, err := s.svc.CreateTable(createTable)
	return err
}

func (s *DynamoStore) getMutex(id, projection string, consistent bool) (*mutex, error) {
	result, err := s.svc.GetItem(&dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(consistent),
		TableName:      s.table,
		Key: map[string]*dynamodb.AttributeValue{
			"entity":   {S: aws.String(id)},
			"revision": {N: aws.String("0")},
		},
		ProjectionExpression: aws.String(projection),
	})
	if err != nil {
		return nil, err
	}

	item := &mutex{}
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *DynamoStore) updateTTL() error {
	updateTTL := &dynamodb.UpdateTimeToLiveInput{
		TableName: s.table,
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String("ttl"),
			Enabled:       aws.Bool(true),
		},
	}
	_, err := s.svc.UpdateTimeToLive(updateTTL)
	return err
}

func (s *DynamoStore) waitForTable() error {
	describeTable := &dynamodb.DescribeTableInput{
		TableName: s.table,
	}
	for i := 0; i < 60; i++ {
		time.Sleep(1 * time.Second)
		result, err := s.svc.DescribeTable(describeTable)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() != dynamodb.ErrCodeResourceNotFoundException {
					return err
				}
			} else {
				return err
			}
		}
		switch aws.StringValue(result.Table.TableStatus) {
		case "CREATING":
			// continue loop
		case "DELETING":
			return ErrDeleteInProgress
		case "ACTIVE", "UPDATING":
			return nil
		}
	}
	return ErrCreateTimedOut
}

type writeTransaction struct {
	ops []*dynamodb.TransactWriteItem
}

func (t *writeTransaction) add(op *dynamodb.TransactWriteItem) *writeTransaction {
	t.ops = append(t.ops, op)
	return t
}

func (t *writeTransaction) addEvent(
	rqx *rqx.RequestContext,
	table *string,
	entity string,
	revision int64,
	typ string,
	data map[string]string,
) error {
	now := time.Now()
	event, err := dynamodbattribute.MarshalMap(&event{
		base: base{
			ID:       entity,
			Revision: revision,
		},
		Created: now,
		TTL:     now.Add(30 * 24 * time.Hour),
		Client: client{
			Type:       rqx.Client.Type,
			RemoteAddr: rqx.Client.RemoteAddr,
			UserAgent:  rqx.Client.UserAgent,
		},
		EUser: user{
			UID:     rqx.EUser.UID.String(),
			Name:    rqx.EUser.Name,
			SlackID: rqx.EUser.SlackID,
		},
		RUser: user{
			UID:     rqx.RUser.UID.String(),
			Name:    rqx.RUser.Name,
			SlackID: rqx.RUser.SlackID,
		},
		Type: typ,
		Data: data,
	})
	if err != nil {
		return err
	}
	t.addPut(&dynamodb.Put{
		TableName: table,
		Item:      event,
		ConditionExpression: aws.String(
			"attribute_not_exists(revision)",
		),
	})
	return nil
}

func (t *writeTransaction) addPut(op *dynamodb.Put) *writeTransaction {
	return t.add(&dynamodb.TransactWriteItem{
		Put: op,
	})
}

func (t *writeTransaction) addUpdate(op *dynamodb.Update) *writeTransaction {
	return t.add(&dynamodb.TransactWriteItem{
		Update: op,
	})
}

func (t *writeTransaction) exec(svc *dynamodb.DynamoDB) error {
	token := make([]byte, 20)
	if _, err := rand.Read(token); err != nil {
		return errors.Wrap(err, "unable to generate request token")
	}

	// TODO: add retry logic
	_, err := svc.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
		TransactItems:      t.ops,
		ClientRequestToken: aws.String(base64.RawURLEncoding.EncodeToString(token)),
	})
	return err
}

type base struct {
	ID       string `dynamodbav:"entity"`
	Revision int64  `dynamodbav:"revision"`
}

type entity struct {
	base
	EntityType  string `dynamodbav:"entity_type"`
	Version     int64  `dynamodbav:"version"`
	Description string `dynamodbav:"description"`
}

type client struct {
	Type       string `dynamodbav:"type,omitempty"`
	RemoteAddr string `dynamodbav:"remote_addr,omitempty"`
	UserAgent  string `dynamodbav:"user_agent,omitempty"`
}

type event struct {
	base
	Created time.Time `dynamodbav:"created,unixtime"`
	TTL     time.Time `dynamodbav:"ttl,unixtime"`
	Client  client    `dynamodbav:"client,omitemptyelem"`
	EUser   user      `dynamodbav:"euser,omitemptyelem"`
	RUser   user      `dynamodbav:"ruser"`

	Type string            `dynamodbav:"type"`
	Data map[string]string `dynamodbav:"data"`
}

type mutex struct {
	entity
	Summary mutexSummary `dynamodbav:"summary"`
}
type mutexSummary struct {
	Locked   bool   `dynamodbav:"locked"`
	LockedBy string `dynamodbav:"locked_by"`
	Message  string `dynamodbav:"message"`
}

type user struct {
	UID     string `dynamodbav:"uid,omitempty"`
	Name    string `dynamodbav:"name,omitempty"`
	SlackID string `dynamodbav:"slack_id,omitempty"`
}
