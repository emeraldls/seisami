build:
	@wails build

run: 	
	@wails dev

sqlc-gen:
	cd sqlc && sqlc generate