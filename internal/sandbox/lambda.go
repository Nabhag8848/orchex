package sandbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	appconfig "github.com/nabhag8848/orchex/internal/config"
)

type Lambda struct {
	client   *lambda.Client
	function string
}

func NewLambda(ctx context.Context, lambdaConfig appconfig.LambdaConfig) (*Lambda, error) {
	loadOptions := make([]func(*config.LoadOptions) error, 0, 1)
	if lambdaConfig.Region != "" {
		loadOptions = append(loadOptions, config.WithRegion(lambdaConfig.Region))
	}

	awsConfig, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	clientOptions := make([]func(*lambda.Options), 0, 1)
	if lambdaConfig.EndpointURL != "" {
		clientOptions = append(clientOptions, func(options *lambda.Options) {
			options.BaseEndpoint = aws.String(lambdaConfig.EndpointURL)
		})
	}

	return &Lambda{
		client:   lambda.NewFromConfig(awsConfig, clientOptions...),
		function: lambdaConfig.FunctionSandboxARN,
	}, nil
}

func (l *Lambda) Invoke(ctx context.Context, source string, input json.RawMessage, timeoutMS int) (json.RawMessage, error) {
	payload, err := json.Marshal(struct {
		Source    string          `json:"source"`
		Input     json.RawMessage `json:"input"`
		TimeoutMS int             `json:"timeout_ms"`
	}{
		Source: source, Input: input, TimeoutMS: timeoutMS,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal sandbox request: %w", err)
	}

	result, err := l.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(l.function),
		InvocationType: "RequestResponse",
		Payload:        payload,
	})
	if err != nil {
		return nil, fmt.Errorf("invoke sandbox: %w", err)
	}
	if result.FunctionError != nil {
		return result.Payload, fmt.Errorf("sandbox function error: %s", *result.FunctionError)
	}
	return result.Payload, nil
}
