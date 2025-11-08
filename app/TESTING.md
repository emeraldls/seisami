# Testing Multiple Instances

You can now run multiple instances of Seisami with separate databases for testing purposes.

## Development Mode (Wails Dev)

When running in development mode, use the `SEISAMI_DB_PATH` environment variable:

```bash
# First instance (default database)
wails dev

# Second instance (custom database) - in a different terminal
SEISAMI_DB_PATH=./seisami-test.db wails dev

# Third instance (another custom database) - in yet another terminal
SEISAMI_DB_PATH=/tmp/seisami-test2.db wails dev
```

## Production Build

For built applications, you can use the `-db` command-line flag:

```bash
# First instance (default database)
./Seisami.app/Contents/MacOS/seisami

# Second instance (custom database)
./Seisami.app/Contents/MacOS/seisami -db=/tmp/seisami-test1.db

# Third instance (another custom database)
./Seisami.app/Contents/MacOS/seisami -db=/tmp/seisami-test2.db
```

## Building and Testing

```bash
# Build the app
wails build

# Run first instance
./build/bin/Seisami.app/Contents/MacOS/seisami

# Run second instance in another terminal
./build/bin/Seisami.app/Contents/MacOS/seisami -db=/tmp/seisami-test.db
```

## Custom Database Locations

You can specify any path for the database:

- **Absolute path**: `-db=/Users/yourname/Desktop/test.db`
- **Relative path**: `-db=./test-data/instance1.db`
- **Temp directory**: `-db=/tmp/seisami-test.db`

The directory will be created automatically if it doesn't exist.

## Use Cases

This feature is particularly useful for:

1. **Testing collaboration features** - Run two instances and test real-time sync
2. **Testing sync behavior** - Create different states in separate instances
3. **Development testing** - Keep production data separate from test data
4. **Demo purposes** - Show multiple users simultaneously
