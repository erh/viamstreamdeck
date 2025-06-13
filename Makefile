
BINARY = bin/viamstreamdeck

$(BINARY): bin *.go cmd/module/*.go *.mod assets/*.jpg
	go build -o $(BINARY) cmd/module/cmd.go

test:
	go test

lint:
	gofmt -w -s .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

module: $(BINARY) meta.json
	tar czf module.tar.gz $(BINARY) meta.json

bin:
	-mkdir bin
