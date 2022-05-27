package domain

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn          *grpc.ClientConn
	productClient ProductServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:          conn,
		productClient: NewProductServiceClient(conn),
	}, nil
}

func (c *Client) CloseConnection() error {
	return c.conn.Close()
}

func (c *Client) Fetch(ctx context.Context, req *FetchRequest) error {
	_, err := c.productClient.Fetch(ctx, req)

	return err
}

func (c *Client) List(ctx context.Context, req *Filters) (*ListResponse, error) {
	return c.productClient.List(ctx, req)
}
