package msclient

import (
	"github.com/google/uuid"
)

type Client struct {
	path string
}

func New(path string) *Client {
	return &Client{
		path: path,
	}
}

func (c *Client) CreateMultiImageUrl() string {
	return uuid.New().String()
}
