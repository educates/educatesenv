APP_NAME=educatesenv
SRC=./main.go

.PHONY: all build clean help

all: build

build:
	go mod download
	go build -o dist/$(APP_NAME) $(SRC)

clean:
	rm -f $(APP_NAME)

help:
	@echo "Makefile for $(APP_NAME)"
	@echo "Available targets:"
	@echo "  build   - Build the binary"
	@echo "  clean   - Remove the binary"
	@echo "  help    - Show this help message" 