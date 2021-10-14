.PHONY: all test testv clean fmt

GOBIN = ./build/bin
GOCMD = env GOPROXY=https://goproxy.io,direct go

all:
	$(GOCMD) build -v ./...
	$(GOCMD) run build/ci.go install ./cmd/...
	@echo "Done building."
	@echo "Find binaries in \"$(GOBIN)\" directory."

test: all
	$(GOCMD) test ./...

testv: all
	$(GOCMD) test -v ./...

clean:
	$(GOCMD) clean -cache
	rm -fr $(GOBIN)/*

fmt:
	./gofmt.sh
