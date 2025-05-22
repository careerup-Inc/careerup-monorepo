package client

import (
	context "context"

	llmpb "github.com/careerup-Inc/careerup-monorepo/proto/llm/v1"
	"google.golang.org/grpc"
)

type LLMClient struct {
	client llmpb.LLMServiceClient
}

func NewLLMClient(conn *grpc.ClientConn) *LLMClient {
	return &LLMClient{
		client: llmpb.NewLLMServiceClient(conn),
	}
}

type LLMAnalysisRequest struct {
	Prompt string
	UserID string
}

type LLMAnalysisResponse struct {
	Completion string
}

func (c *LLMClient) AnalyzeILOResult(ctx context.Context, req *LLMAnalysisRequest) (string, error) {
	stream, err := c.client.GenerateStream(ctx, &llmpb.GenerateStreamRequest{
		Prompt: req.Prompt,
		UserId: req.UserID,
	})
	if err != nil {
		return "", err
	}
	var result string
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		result += resp.GetToken()
	}
	return result, nil
}
