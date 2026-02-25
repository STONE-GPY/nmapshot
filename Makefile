.PHONY: help build dev deploy clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  make build  - Build the Docker image"
	@echo "  make dev    - Run Skaffold for local development"
	@echo "  make deploy - Deploy using Helm to the active Kubernetes context"
	@echo "  make clean  - Remove built binaries and clean up deployments"
	@echo "  make test   - Run Go unit tests"

build:
	docker build -t nmapshot:latest .

dev:
	skaffold dev --port-forward

deploy:
	helm upgrade --install nmapshot ./charts/nmapshot -n default

clean:
	helm uninstall nmapshot -n default || true
	rm -f nmapshot

test:
	go test -v ./...
