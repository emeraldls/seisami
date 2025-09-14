build:
	@wails build

run: 	
	@wails dev

sqlc-gen:
	cd sqlc && sqlc generate

build-darwin_amd64:
	@GOARCH=amd64 GOOS=darwin wails build

build-darwin_arm64:
	@GOARCH=arm64 GOOS=darwin wails build