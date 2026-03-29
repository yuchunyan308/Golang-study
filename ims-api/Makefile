.PHONY: run build tidy test clean

run:
	go run cmd/main.go

build:
	CGO_ENABLED=1 go build -o bin/ims-api cmd/main.go

tidy:
	go mod tidy

test:
	go test ./... -v

clean:
	rm -f bin/ims-api ims.db logs/*.log
