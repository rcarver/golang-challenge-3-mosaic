run:
	rm -f output.png
	go build .
	./golang-challenge-3-mosaic
	test -f output.png && open output.png

test:
	go test

