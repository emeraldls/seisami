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

build-universal:
	@wails build -platform darwin/universal

installer: build-universal
	create-dmg \
  --volname "Seisami Installer" \
  --volicon "build/bin/Seisami.app/Contents/Resources/AppIcon.icns" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --icon "Seisami.app" 100 200 \
  --app-drop-link 400 200 \
  dist/Seisami.dmg \
  build/bin/Seisami.app
