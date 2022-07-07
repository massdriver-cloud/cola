.PHONY: test
test:
	go test ./cmd
	go test ./pkg/...
	go build

.PHONY: docker.build
docker.build:
	DOCKER_BUILDKIT=1 docker build -t 005022811284.dkr.ecr.us-west-2.amazonaws.com/massdriver-cloud/cola .

hack.build-to-massdriver:
	GOOS=linux GOARCH=amd64 go build && cp ./cola ../massdriver/cola-amd64

local.build-to-m1:
	GOOS=darwin GOARCH=arm64 go build
