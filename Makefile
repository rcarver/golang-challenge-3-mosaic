run:
	rm -f output.png
	go build .
	./golang-challenge-3-mosaic
	test -f output.png && open output.png

serve:
	go build .
	go run service/main.go

test:
	go test

