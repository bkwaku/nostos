# Nostos

Nostos is a fast adaptive flow service designed to ingest high-velocity event streams and apply dynamic rate limiting to protect your APIs. It functions as a high-performance API gateway specifically optimized for heavy write requests.

## How It Works

Nostos acts as an ingress layer between your clients and backend services:

1. **Request Ingestion**: Accepts incoming POST requests at the `/ingest` endpoint
2. **Event Processing**: Validates and processes the incoming data
3. **Stream Publishing**: Forwards events to a Kafka topic for reliable, asynchronous processing
4. **Rate Limiting**: (Coming soon) Applies dynamic rate limiting to protect downstream services
5. **Authentication**: (Coming soon) Validates requests before processing

This architecture decouples write operations from your core services, providing resilience, scalability, and the ability to handle traffic spikes without overwhelming backend systems.

## Project Structure

```
nostos/
├── docker-compose.yml       # Docker services configuration
├── kafka-topics.sh          # Kafka topic initialization script
├── go_ingress/              # Main ingress service
│   ├── main.go              # Application entry point
│   ├── config.go            # Configuration management
│   ├── server_injector.go   # Dependency injection
│   ├── kafka/               # Kafka producer implementation
│   │   └── producer.go
│   └── server/              # HTTP server and handlers
│       ├── server.go        # Server implementation
│       ├── types.go         # Request/response types
│       └── response_writer.go
```

## Getting Started Locally

### Prerequisites
- Docker and Docker Compose installed
- Go 1.25.3+ (for local development)

### Running the Service

1. **Build the ingress service**:
   ```bash
   docker-compose build ingress
   ```

2. **Start the service**:
   ```bash
   docker-compose up ingress
   ```
   Note: The startup process includes Kafka topic creation, which may take a moment.

3. **Make a request**:
   ```bash
   curl -X POST http://localhost:8080/ingest \
     -H "Content-Type: application/json" \
     -d '{"your": "data"}'
   ```

### Development

**Run tests**:
```bash
docker-compose run --rm test
```

**Monitor Kafka events**:
```bash
docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server kafka:9092 \
  --topic ingress-topic
```

## How to Contribute

We use [Fizzy](https://app.fizzy.do/6155831/public/boards/Jp19hqsTDqNbXYUFak75XFEZ) to manage our project board.

### Contribution Workflow

1. **Pick a ticket** from the [project board](https://app.fizzy.do/6155831/public/boards/Jp19hqsTDqNbXYUFak75XFEZ)
2. **Move it to "In Progress"** on the board
3. **Create a branch** and implement your changes
4. **Open a Pull Request** with:
   - Clear description of changes
   - Reference to the ticket number
   - Any relevant tests
5. **Wait for review** and address feedback

