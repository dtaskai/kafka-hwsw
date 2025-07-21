package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/joho/godotenv"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	groupID  string
}

func NewConsumer(brokers []string, topic, groupID string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		groupID:  groupID,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	topics := []string{c.topic}

	for {
		err := c.consumer.Consume(ctx, topics, c)
		if err != nil {
			return fmt.Errorf("error from consumer: %w", err)
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer setup completed for topic: %s, group: %s", c.topic, c.groupID)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer cleanup completed for topic: %s, group: %s", c.topic, c.groupID)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Track partition assignments for demonstration
	partitionMap := make(map[string][]int32)
	messageCount := 0
	summaryShown := false

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				if !summaryShown && len(partitionMap) > 0 {
					showPartitionSummary(partitionMap)
					summaryShown = true
				}
				return nil
			}

			messageCount++
			userID := string(message.Key)

			// Track partition assignments
			partitionMap[userID] = append(partitionMap[userID], message.Partition)

			log.Printf("Message #%d received - Partition: %d, Offset: %d, Key: %s, Value: %s",
				messageCount, message.Partition, message.Offset, userID, string(message.Value))

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			if !summaryShown && len(partitionMap) > 0 {
				showPartitionSummary(partitionMap)
				summaryShown = true
			}
			return nil
		}
	}
}

func showPartitionSummary(partitionMap map[string][]int32) {
	log.Printf("")
	log.Printf("=== Partition Distribution Summary ===")
	for userID, partitions := range partitionMap {
		// Get unique partitions for this user
		uniquePartitions := make(map[int32]bool)
		for _, p := range partitions {
			uniquePartitions[p] = true
		}

		// Convert back to slice for display
		var uniqueParts []int32
		for p := range uniquePartitions {
			uniqueParts = append(uniqueParts, p)
		}

		log.Printf("User %s: %d messages all came from partition(s) %v",
			userID, len(partitions), uniqueParts)
	}
	log.Printf("=====================================")
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	brokers := getBrokers()
	topic := getEnv("KAFKA_TOPIC", "test-topic")
	groupID := getEnv("KAFKA_GROUP_ID", "test-consumer-group")
	maxMessages := getEnvAsInt("MAX_MESSAGES", 0)

	log.Printf("Starting Kafka Consumer - Partition Routing Demo")
	log.Printf("Brokers: %v", brokers)
	log.Printf("Topic: %s", topic)
	log.Printf("Group ID: %s", groupID)
	if maxMessages > 0 {
		log.Printf("Max Messages: %d", maxMessages)
	} else {
		log.Printf("Max Messages: Unlimited")
	}
	log.Printf("")
	log.Printf("This demo shows how messages with the same keys (user IDs) come from the same partitions:")
	log.Printf("- All messages for user-123 will come from the same partition")
	log.Printf("- All messages for user-456 will come from the same partition")
	log.Printf("- All messages for user-789 will come from the same partition")
	log.Printf("- etc.")
	log.Printf("")

	consumer, err := NewConsumer(brokers, topic, groupID)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping consumer...")
		cancel()
	}()

	log.Println("Starting to consume messages...")
	if err := consumer.Consume(ctx); err != nil {
		log.Fatalf("Error consuming messages: %v", err)
	}

	log.Println("Consumer stopped")
}

func getBrokers() []string {
	brokersStr := getEnv("KAFKA_BROKERS", "localhost:9092,localhost:9094,localhost:9096")
	return strings.Split(brokersStr, ",")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
