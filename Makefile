client:
	cd cmd/client && go build -o portscan . && mv portscan ../..

server:
	cd cmd/server && go build -o portscan-server . && ./portscan-server

redis-monitor:
	redis-cli monitor

clean:
	rm -rf cmd/server/portscan-server && rm -rf portscan

.PHONY: client server redis-monitor