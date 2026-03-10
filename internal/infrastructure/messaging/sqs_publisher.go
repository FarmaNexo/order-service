package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/farmanexo/order-service/internal/domain/events"
	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/farmanexo/order-service/pkg/config"
	"go.uber.org/zap"
)

type SQSEventPublisher struct {
	sqsClient *sqs.Client
	queueURL  string
	logger    *zap.Logger
}

func NewSQSEventPublisher(awsCfg config.AWSConfig, sqsCfg config.SQSConfig, logger *zap.Logger) (*SQSEventPublisher, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	sqsOpts := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(awsCfg.Endpoint)
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOpts...)

	return &SQSEventPublisher{
		sqsClient: sqsClient,
		queueURL:  sqsCfg.OrderEventsQueueURL,
		logger:    logger,
	}, nil
}

func (p *SQSEventPublisher) Publish(ctx context.Context, event events.OrderEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error serializando evento: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
	}

	_, err = p.sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("error enviando mensaje a SQS: %w", err)
	}

	p.logger.Debug("Evento publicado en SQS",
		zap.String("event_type", event.EventType),
		zap.String("order_id", event.OrderID),
	)

	return nil
}

var _ services.EventPublisher = (*SQSEventPublisher)(nil)
