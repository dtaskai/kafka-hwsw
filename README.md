# Kafka HWSW Meetup Demo

This project provides a complete Kafka setup with Docker Compose and Go applications for producing and consuming messages.

## Features

- **Kafka Cluster**: 3-broker Kafka cluster with Zookeeper
- **Kafka UI**: Web interface for monitoring topics and messages
- **Go Producer**: Application to send messages to Kafka topics
- **Go Consumer**: Application to consume messages from Kafka topics
- **Makefile**: Convenient commands for managing the entire setup

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later

## Quick Start

1. **Start the Kafka cluster:**
   ```bash
   make up
   ```

2. **Create a topic:**
   ```bash
   make bootstrap-topic
   ```

3. **Build the Go applications:**
   ```bash
   make build
   ```

4. **Run the producer (in one terminal):**
   ```bash
   make run-producer
   ```

5. **Run the consumer (in another terminal):**
   ```bash
   make run-consumer
   ```

## Configuration

### Environment Variables

Copy `env.example` to `.env` and customize the settings:

```bash
cp env.example .env
```

**Kafka Configuration:**
- `KAFKA_BROKERS`: Comma-separated list of Kafka broker addresses
- `KAFKA_TOPIC`: Topic name to produce/consume from
- `KAFKA_GROUP_ID`: Consumer group ID

**Producer Configuration:**
- `MESSAGE_COUNT`: Number of messages to send (default: 10)
- `MESSAGE_INTERVAL_MS`: Interval between messages in milliseconds (default: 1000)

**Consumer Configuration:**
- `MAX_MESSAGES`: Maximum messages to consume (0 = unlimited, default: 0)

### Default Values

- **Brokers**: `localhost:9092,localhost:9094,localhost:9096`
- **Topic**: `test-topic`
- **Consumer Group**: `test-consumer-group`
- **Message Count**: 10
- **Message Interval**: 1000ms (1 second)

## Available Commands

### Kafka Management
- `make up` - Start all services
- `make down` - Stop all services
- `make restart` - Restart all services
- `make logs` - View logs
- `make bootstrap-topic` - Create a topic
- `make list-topics` - List all topics
- `make describe-topic TOPIC_NAME=my-topic` - Describe a topic
- `make delete-topic TOPIC_NAME=my-topic` - Delete a topic
- `make cluster-health` - Check cluster health
- `make clean` - Stop services and clean up volumes

### Go Applications
- `make build-producer` - Build the producer
- `make build-consumer` - Build the consumer
- `make build` - Build both applications
- `make run-producer` - Run the producer
- `make run-consumer` - Run the consumer

## Architecture

### Kafka Cluster
- **3 Brokers**: High availability setup
- **Zookeeper**: Coordination service
- **Kafka UI**: Web interface at http://localhost:7777

### Go Applications

#### Producer (`cmd/producer/main.go`)
- Sends user event messages to Kafka topics
- **Partition Routing Demo**: Uses user IDs as keys to demonstrate consistent partition routing
- Simulates real user events (page views, purchases, logins, etc.)
- Shows how the same user ID always routes to the same partition
- Configurable message count and interval
- Graceful shutdown with Ctrl+C
- Logs partition and offset information with partition distribution summary

#### Consumer (`cmd/consumer/main.go`)
- Consumes user event messages from Kafka topics
- **Partition Routing Demo**: Shows how messages with the same keys come from the same partitions
- Uses consumer groups for scalability
- Auto-commits offsets
- Graceful shutdown with Ctrl+C
- Displays partition distribution summary

## Ports

- **Broker 1**: localhost:9092 (external), localhost:9093 (internal)
- **Broker 2**: localhost:9094 (external), localhost:9095 (internal)
- **Broker 3**: localhost:9096 (external), localhost:9097 (internal)
- **Kafka UI**: http://localhost:7777

## Example Usage

1. **Start everything:**
   ```bash
   make up
   make bootstrap-topic
   make build
   ```

2. **Run producer with custom settings:**
   ```bash
   MESSAGE_COUNT=5 MESSAGE_INTERVAL_MS=500 make run-producer
   ```

3. **Run consumer:**
   ```bash
   make run-consumer
   ```

4. **Monitor in Kafka UI:**
   - Open http://localhost:7777
   - Navigate to Topics → test-topic
   - View messages and consumer groups

## Partition Routing Demo

This project includes a demonstration of Kafka's partition routing behavior using a real-life example:

### **User Events Example**
The producer simulates user activity events (page views, purchases, logins, etc.) for specific users:
- `user-123`, `user-456`, `user-789`

### **Key-Based Partitioning**
- **User ID as Key**: Each message uses the user ID as the key
- **Consistent Routing**: The same user ID always routes to the same partition
- **Event Types**: page_view, purchase, login, logout, search, add_to_cart

### **What You'll See**
1. **Producer Output**: Shows which partition each message goes to
2. **Consumer Output**: Shows which partition each message comes from
3. **Summary**: Both producer and consumer show partition distribution per user

### **Example Output**
```
=== Partition Distribution Summary ===
User user-123: 4 messages all went to partition(s) [1]
User user-456: 4 messages all went to partition(s) [2]
User user-789: 4 messages all went to partition(s) [0]
User user-101: 4 messages all went to partition(s) [1]
User user-202: 4 messages all went to partition(s) [2]
=====================================
```

This demonstrates Kafka's guarantee that messages with the same key always go to the same partition, ensuring order and enabling efficient processing per user.

## Development

### Project Structure
```
kafka-hwsw/
├── cmd/
│   ├── producer/
│   │   └── main.go
│   └── consumer/
│       └── main.go
├── docker-compose.yml
├── Makefile
├── go.mod
├── env.example
└── README.md
```

### Dependencies
- `github.com/Shopify/sarama` - Kafka client library
- `github.com/joho/godotenv` - Environment variable loading

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 9092-9097 and 7777 are available
2. **Connection refused**: Wait for Kafka brokers to fully start (may take 30-60 seconds)
3. **Topic not found**: Run `make bootstrap-topic` to create the topic
4. **Consumer not receiving messages**: Check that the topic exists and has messages

### Logs
- View all logs: `make logs`
- View specific service logs: `docker-compose logs -f broker-1`

## Cleanup

To completely clean up the environment:
```bash
make clean
```

This will stop all containers and remove volumes.