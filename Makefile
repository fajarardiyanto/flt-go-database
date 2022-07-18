help:
	@echo "See Makefile"
tidy:
	@ go mod tidy
run-elasticsearch:
	@go run example/elasticsearch/main.go