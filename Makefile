GOSRC = $(shell find . -type f -name '*.go')

build:
	CGO_ENABLED=0 GOOS=linux go build cmd/pg-ha/pg-ha.go

clean:
	rm -rf pg-ha

.PHONY: build
