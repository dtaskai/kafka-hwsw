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

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendMessage(key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message sent successfully - Topic: %s, Partition: %d, Offset: %d, Key: %s, Value: %s",
		p.topic, partition, offset, key, value)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

// UserEvent represents a user activity event
type UserEvent struct {
	UserID    string
	EventType string
	Timestamp time.Time
	Data      map[string]interface{}
}

func generateUserEvents(count int) []UserEvent {
	users := []string{"user-123", "user-456", "user-789"}
	eventTypes := []string{"page_view", "purchase", "login", "logout", "search", "add_to_cart"}

	var events []UserEvent

	for i := 0; i < count; i++ {
		userID := users[i%len(users)]
		eventType := eventTypes[i%len(eventTypes)]

		event := UserEvent{
			UserID:    userID,
			EventType: eventType,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Data: map[string]interface{}{
				"session_id": fmt.Sprintf("session-%d", i),
				"ip_address": fmt.Sprintf("192.168.1.%d", (i%254)+1),
				"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			},
		}

		if eventType == "purchase" {
			event.Data["amount"] = fmt.Sprintf("%.2f", float64((i%1000)+1)/100)
			event.Data["product_id"] = fmt.Sprintf("prod-%d", (i%100)+1)
		} else if eventType == "search" {
			event.Data["query"] = fmt.Sprintf("search term %d", i+1)
		}

		events = append(events, event)
	}

	return events
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	brokers := getBrokers()
	topic := getEnv("KAFKA_TOPIC", "test-topic")
	messageCount := getEnvAsInt("MESSAGE_COUNT", 20)
	messageInterval := getEnvAsInt("MESSAGE_INTERVAL_MS", 500)

	log.Printf("Starting Kafka Producer - Partition Routing Demo")
	log.Printf("Brokers: %v", brokers)
	log.Printf("Topic: %s", topic)
	log.Printf("Message Count: %d", messageCount)
	log.Printf("Message Interval: %dms", messageInterval)
	log.Printf("")

	producer, err := NewProducer(brokers, topic)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping producer...")
		cancel()
	}()

	events := generateUserEvents(messageCount)

	partitionMap := make(map[string][]int32)

	ticker := time.NewTicker(time.Duration(messageInterval) * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Producer stopped")
			return
		case <-ticker.C:
			if count >= messageCount || count >= len(events) {
				log.Printf("Sent %d messages, stopping producer", count)

				showProducerPartitionSummary(partitionMap)
				return
			}

			event := events[count]

			key := event.UserID

			value := fmt.Sprintf(`{"user_id":"%s","event_type":"%s","timestamp":"%s","data":%v}`,
				event.UserID, event.EventType, event.Timestamp.Format(time.RFC3339), event.Data)

			msg := &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(key),
				Value: sarama.StringEncoder(value),
			}

			partition, offset, err := producer.producer.SendMessage(msg)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			} else {
				log.Printf("Message sent - Partition: %d, Offset: %d, Key: %s, Event: %s",
					partition, offset, key, event.EventType)

				partitionMap[key] = append(partitionMap[key], partition)
			}

			count++
		}
	}
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

func showProducerPartitionSummary(partitionMap map[string][]int32) {
	log.Printf("")
	log.Printf("=== Partition Distribution Summary ===")
	for userID, partitions := range partitionMap {
		uniquePartitions := make(map[int32]bool)
		for _, p := range partitions {
			uniquePartitions[p] = true
		}

		var uniqueParts []int32
		for p := range uniquePartitions {
			uniqueParts = append(uniqueParts, p)
		}

		log.Printf("User %s: %d messages all went to partition(s) %v",
			userID, len(partitions), uniqueParts)
	}
	log.Printf("======================================")
}
