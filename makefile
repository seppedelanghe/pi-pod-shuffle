dev:
	go run cmd/player/main.go

build-pi:
	GOOS=linux GOARCH=arm64 go build -o bin/player cmd/player/main.go

