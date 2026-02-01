# nostos
Nostos is fast adaptive flow service designed to ingest high-velocity event streams and apply dynamic rate limiting to protect your APIs, essentially a fast API gateway for heavy write requests

run tests `docker-compose run --rm test`

run `docker-compose build ingress` to build image

you could run `docker-compose up ingress` to just start the service, it might take a while since we attempt to create the Kafka topics before starting ingress, see `kafka-topic.sh`

To view the Kafka stream of events, use: ```docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server kafka:9092 \
  --topic ingress-topic``` 


### How to use this:
1. Run `docker-compose up ingress` to start app(Make sure you build the app first)
1. Make a POST request to `/ingress`. Only supports POST; the idea is to support writes
2. The server will receive your request and send it to Kafka


### TODO:
1. Rate Limiting
2. Authentication and Authorisation
