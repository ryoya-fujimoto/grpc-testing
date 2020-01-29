setup:
	go get github.com/pilu/fresh

run:
	fresh -c runner.conf

lint:
	go vet ./...
	golint -set_exit_status ./...

fmt: lint
	goimports -w .