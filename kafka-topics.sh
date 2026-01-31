#!/bin/bash
# Script to create Kafka topics because Kafka doesn't create topics automatically
echo "Waiting for Kafka to be ready..."
cub kafka-ready -b kafka:9092 1 120

echo "Creating topics..."

kafka-topics --bootstrap-server kafka:9092 \
  --create \
  --if-not-exists \
  --topic ingress-topic \
  --partitions 3 \
  --replication-factor 1

echo "Topics created successfully!"
echo "Listing all topics:"
kafka-topics --bootstrap-server kafka:9092 --list
