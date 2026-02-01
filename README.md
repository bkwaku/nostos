# nostos
Nostos is fast adaptive flow service designed to ingest high-velocity event streams and apply dynamic rate limiting to protect your APIs, essentially a fast API gateway for heavy write requests

run tests `docker-compose run --rm test`

run `docker-compose build ingress` to build image

you could run `docker-compose up ingress` to just start the service, it might take a while since we attempt to create the Kafka topics before starting ingress


TODO:
1. Rate Limiting
2. Authentication and Authorisation
