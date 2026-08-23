package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Message struct {
	Body          string
	MessageID     string
	ReceiptHandle string
}

type Client struct {
	sqs      *sqs.Client
	queueURL string
}

func New(ctx context.Context, queueURL, region, endpoint string) (*Client, error) {
	if queueURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL is required")
	}

	opts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}
	// Local ElasticMQ (and similar) need dummy static creds. Production leaves
	// AWS_ENDPOINT_URL unset so the default chain (ECS task role / SSO) is used.
	if endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	var sqsOpts []func(*sqs.Options)
	if endpoint != "" {
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &Client{
		sqs:      sqs.NewFromConfig(cfg, sqsOpts...),
		queueURL: queueURL,
	}, nil
}

func (c *Client) Send(ctx context.Context, body string) (string, error) {
	out, err := c.sqs.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(body),
	})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.MessageId), nil
}

func (c *Client) Receive(ctx context.Context) ([]Message, error) {
	out, err := c.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, err
	}

	msgs := make([]Message, 0, len(out.Messages))
	for _, m := range out.Messages {
		msgs = append(msgs, Message{
			Body:          aws.ToString(m.Body),
			MessageID:     aws.ToString(m.MessageId),
			ReceiptHandle: aws.ToString(m.ReceiptHandle),
		})
	}
	return msgs, nil
}

func (c *Client) Delete(ctx context.Context, receiptHandle string) error {
	_, err := c.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}
