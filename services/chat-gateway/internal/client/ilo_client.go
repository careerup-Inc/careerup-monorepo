package client

import (
	context "context"

	careerupv1 "github.com/careerup-Inc/careerup-monorepo/proto/careerup/v1"
	"google.golang.org/grpc"
)

type IloClient struct {
	client careerupv1.IloServiceClient
}

func NewIloClient(conn *grpc.ClientConn) *IloClient {
	return &IloClient{
		client: careerupv1.NewIloServiceClient(conn),
	}
}

// GetLatestIloTestResult fetches the latest ILO test result for a user
func (c *IloClient) GetLatestIloTestResult(ctx context.Context, userID string) (*careerupv1.IloTestResult, error) {
	resp, err := c.client.GetIloTestResults(ctx, &careerupv1.GetIloTestResultsRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	results := resp.GetResults()
	if len(results) == 0 {
		return nil, nil // No results
	}
	// Assume the latest is the last one (by created_at)
	return results[len(results)-1], nil
}
