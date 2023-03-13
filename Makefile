.DEFAULT_GOAL := build
@PHONY: build build-nolicense

keygen.exe:
	go build -ldflags "-H windowsgui -s -w" -o keygen.exe keygen/main.go

build:
	CGO_ENABLED=1 GOOS=windows GOARCH=386 fyne package

	