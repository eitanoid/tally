build:
    CGO=0 go build -C backend -o=../tally

test:
    go test ./...

lint:
    golangci-lint run --working-dir backend ./...
    # buf lint

bind:
    CGO_ENABLED=1 gomobile bind -target=android -androidapi 21 -o tally.aar ./pkg/bridge
