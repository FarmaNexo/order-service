package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/farmanexo/order-service/internal/domain/events"
	"github.com/farmanexo/order-service/internal/domain/repositories"
	"github.com/farmanexo/order-service/pkg/config"
	"go.uber.org/zap"
)

type SQSConsumer struct {
	sqsClient    *sqs.Client
	queueURLs    []string
	cartRepo     repositories.CartRepository
	logger       *zap.Logger
	pollInterval time.Duration
	maxMessages  int32
}

func NewSQSConsumer(
	awsCfg config.AWSConfig,
	sqsCfg config.SQSConfig,
	consumerCfg config.ConsumerConfig,
	cartRepo repositories.CartRepository,
	logger *zap.Logger,
) (*SQSConsumer, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "test")),
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	sqsOpts := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOpts = append(sqsOpts, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(awsCfg.Endpoint)
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOpts...)

	queueURLs := []string{
		sqsCfg.CatalogEventsQueueURL,
		sqsCfg.PharmacyEventsQueueURL,
	}

	return &SQSConsumer{
		sqsClient:    sqsClient,
		queueURLs:    queueURLs,
		cartRepo:     cartRepo,
		logger:       logger,
		pollInterval: consumerCfg.PollInterval,
		maxMessages:  consumerCfg.MaxMessages,
	}, nil
}

func (c *SQSConsumer) Start(ctx context.Context) {
	for _, queueURL := range c.queueURLs {
		go func(url string) {
			ticker := time.NewTicker(c.pollInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					c.pollMessages(ctx, url)
				}
			}
		}(queueURL)
	}
}

func (c *SQSConsumer) pollMessages(ctx context.Context, queueURL string) {
	output, err := c.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: c.maxMessages,
		WaitTimeSeconds:     5,
	})
	if err != nil {
		c.logger.Debug("Error recibiendo mensajes SQS", zap.Error(err))
		return
	}

	for _, msg := range output.Messages {
		c.processMessage(ctx, queueURL, msg)
	}
}

func (c *SQSConsumer) processMessage(ctx context.Context, queueURL string, msg sqstypes.Message) {
	if msg.Body == nil {
		c.deleteMessage(ctx, queueURL, msg)
		return
	}

	var baseEvent struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal([]byte(*msg.Body), &baseEvent); err != nil {
		c.logger.Error("Error deserializando evento", zap.Error(err))
		c.deleteMessage(ctx, queueURL, msg)
		return
	}

	switch baseEvent.EventType {
	case "INVENTORY_UPDATED":
		c.handleInventoryUpdated(ctx, *msg.Body)
	case "PRODUCT_UPDATED":
		c.handleProductUpdated(ctx, *msg.Body)
	default:
		c.logger.Debug("Evento ignorado", zap.String("event_type", baseEvent.EventType))
	}

	c.deleteMessage(ctx, queueURL, msg)
}

func (c *SQSConsumer) handleInventoryUpdated(ctx context.Context, body string) {
	var event events.InventoryUpdatedEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		c.logger.Error("Error deserializando INVENTORY_UPDATED", zap.Error(err))
		return
	}

	c.logger.Info("Procesando INVENTORY_UPDATED",
		zap.String("product_id", event.ProductID),
		zap.String("pharmacy_id", event.PharmacyID),
		zap.Int("stock", event.Stock),
	)
	// Could update cart items with new prices or remove items with 0 stock
}

func (c *SQSConsumer) handleProductUpdated(ctx context.Context, body string) {
	var event events.ProductUpdatedEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		c.logger.Error("Error deserializando PRODUCT_UPDATED", zap.Error(err))
		return
	}

	c.logger.Info("Procesando PRODUCT_UPDATED",
		zap.String("product_id", event.ProductID),
		zap.String("name", event.Name),
	)
}

func (c *SQSConsumer) deleteMessage(ctx context.Context, queueURL string, msg sqstypes.Message) {
	c.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
}
