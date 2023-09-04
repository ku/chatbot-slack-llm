gen:
	mockgen  -source=./internal/chatbot/connection/slack.go  -destination=./internal/mock/chatbot/connection/slack.go
test:
	go test ./...
fmt:
	gofumpt -w .
