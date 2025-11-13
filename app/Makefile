build:
	@wails build
	@echo "Applying ad-hoc code signature..."
	@codesign --force --deep --sign - build/bin/Seisami.app
	@echo "Build signed successfully!"

run: 	
	@wails dev

sqlc-gen:
	@cd sqlc && sqlc generate

# build-windows_amd64:
#   @wails build --target windows/amd64 --nsis

# build-windows_arm64:
#   @wails build --target windows/arm64 --nsis

# installer-windows:
#   @makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe

build-darwin_amd64:
	@GOARCH=amd64 GOOS=darwin wails build
	@echo "Applying ad-hoc code signature..."
	@codesign --force --deep --sign - build/bin/Seisami.app
	@echo "Build signed successfully!"

build-darwin_arm64:
	@GOARCH=arm64 GOOS=darwin wails build
	@echo "Applying ad-hoc code signature..."
	@codesign --force --deep --sign - build/bin/Seisami.app
	@echo "Build signed successfully!"

build-universal:
	@wails build -platform darwin/universal
	@echo "Applying ad-hoc code signature..."
	@codesign --force --deep --sign - build/bin/Seisami.app
	@echo "Build signed successfully!"

installer-macos: build-universal
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
