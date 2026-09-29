mod app "./apps/tally/app.just"

build:
    CGO=0 go build -C backend -o=../tally

test:
    cd backend && go test ./... -v

lint:
    cd backend && golangci-lint run 
    buf lint

generate:
    buf generate && cd backend && go mod tidy

bind: bind-android

# build the Go backend into a .aar package and copy it into the react-native module
bind-android:
    cd backend && gomobile bind -v -target=android -androidapi 24 \
        -o {{ justfile_dir() }}/apps/tally/modules/tally-backend/android/src/libs/tally.aar ./pkg/bridge
