.DEFAULT_GOAL := build
@PHONY: build build-nolicense

keygen.exe:
	go build -ldflags "-H windowsgui -s -w" -o keygen.exe keygen/main.go

hcd.exe:
	go build -ldflags "-H windowsgui -s -w" -o hcd.exe .

build:
	rm -rf "dist/Saab NG9-3 Halogen Check Disabler.exe"
	CGO_ENABLED=1 GOOS=windows GOARCH=386 fyne package --release
	mv "Saab NG9-3 Halogen Check Disabler.exe" "dist/Saab NG9-3 Halogen Check Disabler.exe"
	