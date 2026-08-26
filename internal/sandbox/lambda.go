package sandbox

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// pingPayload is a Function-node body that passes input_schema ({"data": {}}).
var pingPayload = []byte(`{"source":"return { ping: true };","input":{"data":{}},"timeout_ms":5000}`)

type Client struct {
	lambda *lambda.Client
	arn    string
}

func New(ctx context.Context, arn, region, endpoint string) (*Client, error) {
	if arn == "" {
		return nil, fmt.Errorf("FUNCTION_SANDBOX_ARN is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	// Local SAM (and similar) need dummy static creds. Production leaves
	// LAMBDA_ENDPOINT_URL unset so the default chain (ECS task role / SSO) is used.
	if endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	var lambdaOpts []func(*lambda.Options)
	if endpoint != "" {
		lambdaOpts = append(lambdaOpts, func(o *lambda.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &Client{
		lambda: lambda.NewFromConfig(cfg, lambdaOpts...),
		arn:    arn,
	}, nil
}

// Ping sync-Invokes the sandbox with a tiny valid Function payload.
func (c *Client) Ping(ctx context.Context) ([]byte, error) {
	out, err := c.lambda.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(c.arn),
		Payload:      pingPayload,
	})
	if err != nil {
		return nil, err
	}
	if out.FunctionError != nil && *out.FunctionError != "" {
		return out.Payload, fmt.Errorf("function error %s: %s", *out.FunctionError, out.Payload)
	}
	return out.Payload, nil
}
