.PHONY: up down restart logs bootstrap-topic list-topics clean build-producer build-consumer run-producer run-consumer

# Default topic configuration
TOPIC_NAME ?= test-topic
PARTITIONS ?= 3
REPLICATION_FACTOR ?= 3

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Restart all services
restart: down up

# View logs
logs:
	docker-compose logs -f

# Bootstrap a topic in Kafka
bootstrap-topic:
	@echo "Creating topic: $(TOPIC_NAME) with $(PARTITIONS) partitions and replication factor $(REPLICATION_FACTOR)"
	docker exec broker-1 kafka-topics --create \
		--topic $(TOPIC_NAME) \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097 \
		--partitions $(PARTITIONS) \
		--replication-factor $(REPLICATION_FACTOR) \
		--if-not-exists

# List all topics
list-topics:
	docker exec broker-1 kafka-topics --list \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097

# Describe a specific topic
describe-topic:
	@if [ -z "$(TOPIC_NAME)" ]; then \
		echo "Error: TOPIC_NAME is required. Usage: make describe-topic TOPIC_NAME=your-topic"; \
		exit 1; \
	fi
	docker exec broker-1 kafka-topics --describe \
		--topic $(TOPIC_NAME) \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097

# Delete a topic
delete-topic:
	@if [ -z "$(TOPIC_NAME)" ]; then \
		echo "Error: TOPIC_NAME is required. Usage: make delete-topic TOPIC_NAME=your-topic"; \
		exit 1; \
	fi
	@echo "Deleting topic: $(TOPIC_NAME)"
	docker exec broker-1 kafka-topics --delete \
		--topic $(TOPIC_NAME) \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097

# Check cluster health
cluster-health:
	@echo "Checking cluster health..."
	docker exec broker-1 kafka-broker-api-versions \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097

# Show broker details
broker-details:
	@echo "Broker details:"
	docker exec broker-1 kafka-topics --describe \
		--bootstrap-server broker-1:9093,broker-2:9095,broker-3:9097

# Clean up everything (stop containers and remove volumes)
clean:
	docker-compose down -v
	docker system prune -f

# Build Go applications
build-producer:
	go build -o bin/producer cmd/producer/main.go

build-consumer:
	go build -o bin/consumer cmd/consumer/main.go

build: build-producer build-consumer

# Run Go applications
run-producer: build-producer
	./bin/producer

run-consumer: build-consumer
	./bin/consumer

# Show help
help:
	@echo "Available commands:"
	@echo "  up              - Start all services (3 brokers + zookeeper + kafka-ui)"
	@echo "  down            - Stop all services"
	@echo "  restart         - Restart all services"
	@echo "  logs            - View logs"
	@echo "  bootstrap-topic - Create a topic (default: test-topic with 3 partitions, RF=3)"
	@echo "  list-topics     - List all topics"
	@echo "  describe-topic  - Describe a specific topic (requires TOPIC_NAME)"
	@echo "  delete-topic    - Delete a topic (requires TOPIC_NAME)"
	@echo "  cluster-health  - Check cluster health"
	@echo "  broker-details  - Show broker details"
	@echo "  clean           - Stop services and clean up volumes"
	@echo ""
	@echo "Go Application Commands:"
	@echo "  build-producer  - Build the Kafka producer"
	@echo "  build-consumer  - Build the Kafka consumer"
	@echo "  build           - Build both producer and consumer"
	@echo "  run-producer    - Run the Kafka producer"
	@echo "  run-consumer    - Run the Kafka consumer"
	@echo ""
	@echo "Examples:"
	@echo "  make bootstrap-topic TOPIC_NAME=my-topic PARTITIONS=5 REPLICATION_FACTOR=3"
	@echo "  make describe-topic TOPIC_NAME=my-topic"
	@echo "  make delete-topic TOPIC_NAME=my-topic"
	@echo "  make run-producer"
	@echo "  make run-consumer"
	@echo ""
	@echo "Ports:"
	@echo "  Broker 1: localhost:9092 (external), localhost:9093 (internal)"
	@echo "  Broker 2: localhost:9094 (external), localhost:9095 (internal)"
	@echo "  Broker 3: localhost:9096 (external), localhost:9097 (internal)"
	@echo "  Kafka UI: http://localhost:7777" 