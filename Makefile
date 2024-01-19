all:
	go mod tidy
	go build -o ./build/blow-handles ./cmd/blow-handles
	go build -o ./build/echo-handles ./cmd/echo-handles
