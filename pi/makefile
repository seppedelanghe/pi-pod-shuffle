IMAGE := pi-go-alsa:1.25

docker:
	docker build -t $(IMAGE) .

pi: docker
	docker run --rm -v "$(PWD)":/app -w /app \
		-e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=arm64 \
		$(IMAGE) \
		go build -o ./dist/pi-pod-shuffle ./cmd/avrcp/main.go

pi-debug: docker
	docker run --rm -v "$(PWD)":/app -w /app \
		-e CGO_ENABLED=1 -e GOOS=linux -e GOARCH=arm64 -e GOMAXPROCS=1 -e GOGC=200 \
		$(IMAGE) \
		go build -o ./dist/pi-pod-shuffle ./cmd/debug/main.go

pi-deploy: pi
	scp ./dist/pi-pod-shuffle pipod@pi-pod-shuffle.local:/home/pipod/

clean:
	rm -rf ./dist
