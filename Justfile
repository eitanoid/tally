build:
    CGO=0 go build -C backend -o=../tally

test:
    cd backend && go test ./...

lint:
    cd backend && golangci-lint run 
    # buf lint

bind:
    CGO_ENABLED=1 gomobile bind -target=android -androidapi 21 -o tally.aar ./pkg/bridge
