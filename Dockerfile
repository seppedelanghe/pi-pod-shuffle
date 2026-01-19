FROM arm64v8/golang:1.25
RUN apt-get update && apt-get install -y --no-install-recommends libasound2-dev && rm -rf /var/lib/apt/lists/*
