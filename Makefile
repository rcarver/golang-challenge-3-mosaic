build:
	go install github.com/rcarver/golang-challenge-3-mosaic/mosaicly

sample.jpg: build
	$$GOPATH/bin/mosaicly fetch -tag balloon -num 2000
	$$GOPATH/bin/mosaicly gen -tag balloon -in fixtures/balloon-square.jpg -out $@
	test -f $@ && open $@

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

cov_packages=mosaic instagram
cov_files=$(addsuffix .coverage.out,$(cov_packages))

cov: clean_cov $(cov_files)

clean_cov: 
	rm -f $(cov_files)

%.coverage.out:
	go test -coverprofile=$@ ./$(firstword $(subst ., ,$@))
	go tool cover -html=$<

