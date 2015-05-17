run:
	rm -f output.png
	go run cli/main.go
	test -f output.png && open output.png

test:
	go test


serve:
	go run cli/main.go -run serve

test_serve: 
	$(MAKE) -j serve service_test

service_test:
	./service_test.sh
