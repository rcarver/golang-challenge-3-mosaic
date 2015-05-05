run:
	rm -f output.png
	go run cli/main.go
	test -f output.png && open output.png

serve:
	go run cli/main.go -serve

test:
	go test

