package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	appconfig "github.com/nabhag8848/orchex/internal/config"
)

type SQS struct {
	client *sqs.Client
	url    string
}

type NodeJobMessage struct {
	RunID             uuid.UUID `json:"run_id"`
	WorkflowVersionID uuid.UUID `json:"workflow_version_id"`
	NodeID            uuid.UUID `json:"node_id"`
	Attempt           int32     `json:"attempt"`
}

func New(ctx context.Context, sqsConfig appconfig.SQSConfig) (*SQS, error) {
	loadOptions := make([]func(*config.LoadOptions) error, 0, 1)
	if region := sqsConfig.Region; region != "" {
		loadOptions = append(loadOptions, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}

	clientOptions := make([]func(*sqs.Options), 0, 1)
	// Use the local SQS endpoint when one is configured. Otherwise use AWS SQS.
	if endpoint := sqsConfig.EndpointURL; endpoint != "" {
		clientOptions = append(clientOptions, func(options *sqs.Options) {
			options.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &SQS{
		client: sqs.NewFromConfig(cfg, clientOptions...),
		url:    sqsConfig.QueueURL,
	}, nil
}

// Send adds a node job to the queue.
func (q *SQS) Send(ctx context.Context, message NodeJobMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal node job: %w", err)
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.url),
		MessageBody: aws.String(string(body)),
	})
	return err
}

// Receive waits up to 20 seconds for up to ten messages.
func (q *SQS) Receive(ctx context.Context) ([]types.Message, error) {
	output, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.url),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   30,
	})
	if err != nil {
		return nil, err
	}
	return output.Messages, nil
}

// Delete removes a received message using its receipt handle.
func (q *SQS) Delete(ctx context.Context, receiptHandle string) error {
	_, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.url),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}
