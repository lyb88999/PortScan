client:
	cd cmd/client && go build -o portscan .

server:
	cd cmd/server && go run main.go

redis-monitor:
	redis-cli monitor

.PHONY: client server redis-monitor