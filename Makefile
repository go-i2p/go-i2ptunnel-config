fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

test:
	go test -race ./...

vet:
	go vet ./...