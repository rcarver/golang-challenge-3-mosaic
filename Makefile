run:
	rm -f output.jpg
	go install github.com/rcarver/golang-challenge-3-mosaic/mosaicly
	$$GOPATH/bin/mosaicly -run fetch -tag balloon
	$$GOPATH/bin/mosaicly -run gen -tag balloon -in fixtures/balloon.jpg -out output.jpg
	test -f output.jpg && open output.jpg

serve:
	$$GOPATH/bin/mosaicly -run serve

service_test:
	./service_test.sh

test: test_unit test_serve

test_unit:
	go test ./...

test_serve: 
	$(MAKE) -j serve service_test

