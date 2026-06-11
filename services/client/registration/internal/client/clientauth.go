package client

import (
	"context"

	clientauthv1 "github.com/dealer/dealer/pkg/pb/clientauth/v1"
	"github.com/dealer/dealer/pkg/grpclient"
	"google.golang.org/grpc"
)

type ClientAuthClient struct {
	conn *grpc.ClientConn
	api  clientauthv1.ClientAuthServiceClient
}

func NewClientAuthClient(ctx context.Context, addr string) (*ClientAuthClient, error) {
	conn, err := grpclient.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &ClientAuthClient{conn: conn, api: clientauthv1.NewClientAuthServiceClient(conn)}, nil
}

func (c *ClientAuthClient) Close() error {
	return c.conn.Close()
}

func (c *ClientAuthClient) IssueTokens(ctx context.Context, userID string) (access, refresh string, expiresAt int64, err error) {
	resp, err := c.api.IssueTokens(ctx, &clientauthv1.IssueTokensRequest{UserId: userID})
	if err != nil {
		return "", "", 0, err
	}
	return resp.AccessToken, resp.RefreshToken, resp.ExpiresAt, nil
}
