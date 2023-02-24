.DEFAULT_GOAL := build
@PHONY: build

build:
	CGO_ENABLED=1 GOOS=windows GOARCH=386 fyne package --icon ./window/hk_white.png
