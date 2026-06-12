build:
	go build -o ./bin/flux ./cmd/flux

install:
	sudo cp ./bin/flux /usr/local/bin/flux