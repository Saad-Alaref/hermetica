APP=hermetica

.PHONY: tools build run doctor

tools:
	@echo "Checking external tools..."
	@$(APP) doctor --dry-run || (echo "Doctor failed" && exit 1)

build:
	@echo "Building $(APP)"
	@go build -o bin/$(APP) ./cmd/hermetica

run:
	@go run ./cmd/hermetica -- run -c configs/hermetica.yaml

doctor:
	@go run ./cmd/hermetica -- doctor --dry-run -c configs/hermetica.yaml

