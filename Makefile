test: ## Run unit tests
	go test -count=1 -short ./...

cover: dep
	go test $(shell go list ./... | grep -v /vendor/;) -cover -v

dep: ## Get the dependencies
	GO111MODULE=on go mod vendor