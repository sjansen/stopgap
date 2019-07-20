package rqx

import (
	"context"

	"github.com/oklog/ulid/v2"
)

type RequestContext struct {
	Ctx    context.Context
	Client Client
	EUser  User
	RUser  User
}

type User struct {
	UID   ulid.ULID
	Name  string
	Email string
}

type Client struct {
	Type       string
	RemoteAddr string
	UserAgent  string
}
