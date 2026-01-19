IMAGE := pi-go-alsa:1.25

dev:
	go run ./cmd/player/main.go

docker:
	docker build -t $(IMAGE) .

pi: docker
	docker run --rm -v "$(PWD)":/app -w /app \
		-e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=arm64 \
		$(IMAGE) \
		go build -o ./dist/pi-pod-shuffle ./cmd/player/main.go

pi-debug: docker
	docker run --rm -v "$(PWD)":/app -w /app \
		-e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=arm64 \
		$(IMAGE) \
		go build -o ./dist/pi-pod-shuffle ./cmd/debug/main.go

clean:
	rm -rf ./dist
