PACKNAME=github-repo

release: package release bin
release-candidate: package release-candidate
binary: package build

release: release-nix

docker-build: release-nix
	@echo "  >  Building docker image ..."
	@docker build -t $(PACKNAME) .

docker-run:
	@echo "  >  Running docker image ..."
	docker run -it --rm -v ./config.yml:/.privateer/config.yml -v ./docker_output:/evaluation_results $(PACKNAME)

# Pushing to eddieknight docker hub namespace until a more preferred option exists
docker-build-release:
	@echo "  >  Building docker images for linux/amd64 and linux/arm64..."
	docker buildx build --platform linux/amd64,linux/arm64 -t eddieknight/pvtr-github-repo:latest .

docker-run-latest:
	@echo "  >  Running docker image ..."
	docker pull eddieknight/pvtr-github-repo:latest
	docker run -it --rm -v ./config.yml:/.privateer/config.yml -v ./docker_output:/evaluation_results eddieknight/pvtr-github-repo:latest

build:
	@echo "  >  Building binary ..."
	@go build -o $(PACKNAME) -ldflags="$(BUILD_FLAGS)"

package: tidy test
	@echo "  >  Packaging static files..."

test:
	@echo "  >  Validating code ..."
	@go vet ./...
	@go clean -testcache
	@go test ./...

tidy:
	@echo "  >  Tidying go.mod ..."
	@go mod tidy

test-cov:
	@echo "Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count
	@sleep 2 # Sleeping to allow for coverage.out file to get generated
	@echo "Current test coverage : $(shell go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') %"

release-candidate: tidy test
	@make release-nix

release-nix:
	@echo "  >  Building release for Linux..."
	goreleaser release --snapshot --clean

bin:
	@mv $(PACKNAME)* ~/privateer/bin
