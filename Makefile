NAME ?= testkube-karate-executor
BIN_DIR ?= $(HOME)/bin
NAMESPACE ?= "default"

build:
	go build -o $(BIN_DIR)/$(NAME) cmd/agent/main.go 

.PHONY: test cover build

run: 
	EXECUTOR_PORT=8082 go run cmd/agent/main.go

mongo-dev: 
	docker run -p 27017:27017 mongo

docker-build: 
	docker build -t lreimer/$(NAME) -f build/agent/Dockerfile .

install-jbang:
	curl -Ls https://sh.jbang.dev | bash -s - app setup
	export PATH="$$HOME/.jbang/bin:$$PATH"

install-jbang-mac:
	brew install jbangdev/tap/jbang

install-karate:
	jbang app install --name karate com.intuit.karate:karate-core:1.1.0

install-swagger-codegen-mac: 
	brew install swagger-codegen

test: 
	go test ./... -cover

test-e2e:
	go test --tags=e2e -v ./test/e2e

test-e2e-namespace:
	NAMESPACE=$(NAMESPACE) go test --tags=e2e -v  ./test/e2e 

cover: 
	@go test -failfast -count=1 -v -tags test  -coverprofile=./testCoverage.txt ./... && go tool cover -html=./testCoverage.txt -o testCoverage.html && rm ./testCoverage.txt 
	open testCoverage.html


version-bump: version-bump-patch

version-bump-patch:
	go run cmd/tools/main.go bump -k patch

version-bump-minor:
	go run cmd/tools/main.go bump -k minor

version-bump-major:
	go run cmd/tools/main.go bump -k major

version-bump-dev:
	go run cmd/tools/main.go bump --dev
