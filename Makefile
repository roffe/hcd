.DEFAULT_GOAL := build
@PHONY: build build-nolicense setup

setup: build
	makensis -DVERSION=$$(sed -n 's/.*Version = "\(.*\)"/\1/p' FyneApp.toml) setup/setup.nsi

hcd.exe:
	go build -ldflags "-H windowsgui -s -w" -o hcd.exe .

build:
	rm -rf "dist/Saab NG9-3 Halogen Check Disabler.exe"
	rsrc -arch 386 -manifest manifest.xml
	CGO_ENABLED=1 GOOS=windows GOARCH=386 fyne package --release
# mv "Saab NG9-3 Halogen Check Disabler.exe" "dist/Saab NG9-3 Halogen Check Disabler.exe"
	rm -rf rsrc_windows_386.syso