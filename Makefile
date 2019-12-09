test:
	go test -race -v ./...

bench:
	go test -run="^$$" -bench=.

cover:
	go test $(shell go list ./... | grep -v /vendor/;) -cover -v

dep:
	GO111MODULE=on go mod vendor
