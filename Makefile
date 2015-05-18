build:
	go install github.com/rcarver/golang-challenge-3-mosaic/mosaicly

run: build
	rm -f output.jpg
	$$GOPATH/bin/mosaicly -run fetch -tag balloon
	$$GOPATH/bin/mosaicly -run gen -tag balloon -in fixtures/balloon.jpg -out output.jpg
	test -f output.jpg && open output.jpg

serve: build
	$$GOPATH/bin/mosaicly -run serve

test: test_unit test_cli test_service 

test_unit:
	go test ./...

test_cli: build
	./tests/cli.sh

test_service: build
	./tests/service.sh

