build:
	go install github.com/rcarver/golang-challenge-3-mosaic/mosaicly

gen: build
	rm -f output.jpg
	$$GOPATH/bin/mosaicly fetch -tag balloon
	$$GOPATH/bin/mosaicly gen -tag balloon -in fixtures/balloon.jpg -out output.jpg
	test -f output.jpg && open output.jpg

serve: build
	$$GOPATH/bin/mosaicly serve

test: test_unit test_cli test_service 

test_unit:
	go test ./...

test_cli: build
	./tests/cli.sh

test_service: build
	./tests/service.sh

lint:
	go fmt ./...
	go vet ./...
	$$GOPATH/bin/golint ./...

cov:
	go test -coverprofile=mosaic.coverage.out ./mosaic
	go test -coverprofile=instagram.coverage.out ./instagram
	go tool cover -html=mosaic.coverage.out
	go tool cover -html=instagram.coverage.out

