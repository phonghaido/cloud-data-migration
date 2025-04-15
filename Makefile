export AWS_ACCESS_KEY_ID :=
export AWS_SECRET_ACCESS_KEY :=
export AWS_REGION :=
export AWS_S3_BUCKET :=

export GCP_PROJECT_ID :=
export GCP_BUCKET_NAME :=
export GCP_CREDENTIALS := ./k8s/dev/pubsub/secret/svc_account.json
export GCP_PUBSUB_HOST := http://localhost:8085
export GCP_SUBSCRIPTION_ID :=
export GCP_TOPIC_ID :=
export SYSTEM_MAX_WORKERS := 5


all: build run

build:
	go build -o ./bin/sub ./cmd/subscribe_to_topic
	chmod +x ./bin/sub

run:
	./bin/app

docker-build:
	docker build -t data-migration-app:latest .

compose-up:
	docker compose -f docker-compose-dev.yaml up -d --build
compose-down:
	docker compose -f docker-compose-dev.yaml down

dev-pubsub:
	kubectl apply -k ./k8s/dev/pubsub
	kubectl rollout restart -n gcloud deployment/pubsub-emulator

prod-pubsub:
	kubectl apply -k ./k8s/production/pubsub
	kubectl rollout restart -n gcloud deployment/pubsub-emulator

dev-migration-mk:
	docker build -f Dockerfile.sub -t data-migration-app:latest .
	kubectl apply -k ./k8s/dev/migration_service
	kubectl rollout restart -n data-migration deployment/data-migration-app

dev-publisher-mk:
	docker build -f Dockerfile.pub -t publisher:latest .
	kubectl apply -k ./k8s/dev/publisher
	kubectl rollout restart -n data-migration deployment/publisher

prod-migration-mk:
	docker build -f Dockerfile.sub -t data-migration-app:latest .
	kubectl apply -k ./k8s/production/migration_service
	kubectl rollout restart -n data-migration deployment/data-migration-app

dev-redis:
	kubectl apply -k ./k8s/dev/redis
	kubectl rollout restart -n database statefulset/redis

