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

// IloDomain represents one of the 5 domains assessed in the ILO test
type IloDomain struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// IloLevel represents the evaluation level for a domain score
type IloLevel struct {
	MinPercent int32  `json:"min_percent"`
	MaxPercent int32  `json:"max_percent"`
	LevelName  string `json:"level_name"`
	Suggestion string `json:"suggestion"`
}

// IloDomainScore represents a scored domain for a user
type IloDomainScore struct {
	DomainCode string  `json:"domain_code"`
	RawScore   int32   `json:"raw_score"`
	Percent    float32 `json:"percent"`
	Level      string  `json:"level"`
	Rank       int32   `json:"rank"`
}

// IloAnswer represents a single answer to an ILO test question
type IloAnswer struct {
	QuestionID     string `json:"question_id"`
	QuestionNumber int32  `json:"question_number"`
	SelectedOption int32  `json:"selected_option"`
}

type SubmitILOTestResultRequest struct {
	UserID        string
	Answers       []IloAnswer
	RawResultData string
}

type SubmitILOTestResultResponse struct {
	ID               string
	UserID           string
	ResultData       string
	CreatedAt        string
	Scores           []IloDomainScore
	TopDomains       []string
	SuggestedCareers []string
}

// IloTestQuestion represents a question in the ILO test
type IloTestQuestion struct {
	ID             string   `json:"id"`
	QuestionNumber int32    `json:"question_number"`
	Text           string   `json:"text"`
	DomainCode     string   `json:"domain_code"`
	Options        []string `json:"options"`
}

// GetIloTestResponse represents the response for ILO test questions
type GetIloTestResponse struct {
	Questions []IloTestQuestion `json:"questions"`
	Domains   []IloDomain       `json:"domains"`
	Levels    []IloLevel        `json:"levels"`
}

func (c *IloClient) SubmitILOTestResult(ctx context.Context, req *SubmitILOTestResultRequest) (*SubmitILOTestResultResponse, error) {
	// Convert client answers to proto answers
	protoAnswers := make([]*careerupv1.IloAnswer, len(req.Answers))
	for i, answer := range req.Answers {
		protoAnswers[i] = &careerupv1.IloAnswer{
			QuestionId:     answer.QuestionID,
			QuestionNumber: answer.QuestionNumber,
			SelectedOption: answer.SelectedOption,
		}
	}

	resp, err := c.client.SubmitIloTestResult(ctx, &careerupv1.SubmitIloTestResultRequest{
		UserId:        req.UserID,
		Answers:       protoAnswers,
		RawResultData: req.RawResultData,
	})

	if err != nil {
		return nil, err
	}

	result := resp.GetResult()

	// Convert proto domain scores to client domain scores
	scores := make([]IloDomainScore, len(result.GetScores()))
	for i, score := range result.GetScores() {
		scores[i] = IloDomainScore{
			DomainCode: score.GetDomainCode(),
			RawScore:   score.GetRawScore(),
			Percent:    score.GetPercent(),
			Level:      score.GetLevel(),
			Rank:       score.GetRank(),
		}
	}

	return &SubmitILOTestResultResponse{
		ID:               result.GetId(),
		UserID:           result.GetUserId(),
		ResultData:       result.GetResultData(),
		CreatedAt:        result.GetCreatedAt(),
		Scores:           scores,
		TopDomains:       result.GetTopDomains(),
		SuggestedCareers: result.GetSuggestedCareers(),
	}, nil
}

// GetIloTest retrieves the ILO test questions from the backend service
func (c *IloClient) GetIloTest(ctx context.Context) (*GetIloTestResponse, error) {
	resp, err := c.client.GetIloTest(ctx, &careerupv1.GetIloTestRequest{})
	if err != nil {
		return nil, err
	}

	questions := make([]IloTestQuestion, 0, len(resp.GetQuestions()))
	for _, q := range resp.GetQuestions() {
		questions = append(questions, IloTestQuestion{
			ID:             q.GetId(),
			QuestionNumber: q.GetQuestionNumber(),
			Text:           q.GetText(),
			DomainCode:     q.GetDomainCode(),
			Options:        q.GetOptions(),
		})
	}

	domains := make([]IloDomain, 0, len(resp.GetDomains()))
	for _, d := range resp.GetDomains() {
		domains = append(domains, IloDomain{
			Code:        d.GetCode(),
			Name:        d.GetName(),
			Description: d.GetDescription(),
		})
	}

	levels := make([]IloLevel, 0, len(resp.GetLevels()))
	for _, l := range resp.GetLevels() {
		levels = append(levels, IloLevel{
			MinPercent: l.GetMinPercent(),
			MaxPercent: l.GetMaxPercent(),
			LevelName:  l.GetLevelName(),
			Suggestion: l.GetSuggestion(),
		})
	}

	return &GetIloTestResponse{
		Questions: questions,
		Domains:   domains,
		Levels:    levels,
	}, nil
}

// GetIloCareerSuggestions retrieves career suggestions based on domain scores
func (c *IloClient) GetIloCareerSuggestions(ctx context.Context, domainCodes []string, limit int32) ([]string, error) {
	resp, err := c.client.GetIloCareerSuggestions(ctx, &careerupv1.GetIloCareerSuggestionsRequest{
		DomainCodes: domainCodes,
		Limit:       limit,
	})

	if err != nil {
		return nil, err
	}

	careers := make([]string, 0, len(resp.GetSuggestions()))
	for _, suggestion := range resp.GetSuggestions() {
		careers = append(careers, suggestion.GetCareerField())
	}

	return careers, nil
}

// GetIloTestResults retrieves all ILO test results for a user
func (c *IloClient) GetIloTestResults(ctx context.Context, userID string) ([]*SubmitILOTestResultResponse, error) {
	resp, err := c.client.GetIloTestResults(ctx, &careerupv1.GetIloTestResultsRequest{
		UserId: userID,
	})

	if err != nil {
		return nil, err
	}

	results := make([]*SubmitILOTestResultResponse, 0, len(resp.GetResults()))
	for _, protoResult := range resp.GetResults() {
		// Convert proto domain scores to client domain scores
		scores := make([]IloDomainScore, len(protoResult.GetScores()))
		for i, score := range protoResult.GetScores() {
			scores[i] = IloDomainScore{
				DomainCode: score.GetDomainCode(),
				RawScore:   score.GetRawScore(),
				Percent:    score.GetPercent(),
				Level:      score.GetLevel(),
				Rank:       score.GetRank(),
			}
		}

		results = append(results, &SubmitILOTestResultResponse{
			ID:               protoResult.GetId(),
			UserID:           protoResult.GetUserId(),
			ResultData:       protoResult.GetResultData(),
			CreatedAt:        protoResult.GetCreatedAt(),
			Scores:           scores,
			TopDomains:       protoResult.GetTopDomains(),
			SuggestedCareers: protoResult.GetSuggestedCareers(),
		})
	}

	return results, nil
}

// GetIloTestResultById retrieves a specific ILO test result by ID
func (c *IloClient) GetIloTestResultById(ctx context.Context, resultID string) (*SubmitILOTestResultResponse, error) {
    // Call the proper gRPC method
    resp, err := c.client.GetIloTestResult(ctx, &careerupv1.GetIloTestResultRequest{
        ResultId: resultID,
    })
    
    if err != nil {
        return nil, err
    }
    
    result := resp.GetResult()
    
    // Convert the proto result to client result (like in your other methods)
    scores := make([]IloDomainScore, len(result.GetScores()))
    for i, score := range result.GetScores() {
        scores[i] = IloDomainScore{
            DomainCode: score.GetDomainCode(),
            RawScore:   score.GetRawScore(),
            Percent:    score.GetPercent(),
            Level:      score.GetLevel(),
            Rank:       score.GetRank(),
        }
    }
    
    return &SubmitILOTestResultResponse{
        ID:               result.GetId(),
        UserID:           result.GetUserId(),
        ResultData:       result.GetResultData(),
        CreatedAt:        result.GetCreatedAt(),
        Scores:           scores,
        TopDomains:       result.GetTopDomains(),
        SuggestedCareers: result.GetSuggestedCareers(),
    }, nil
}