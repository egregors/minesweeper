# minesweeper

Multiplayer Minesweeper game, just for recreation programming purpose.

> **Warning**
> Development in process

## Screenshots

<img width="1346" alt="Screenshot 2022-10-26 at 12 42 04" src="https://user-images.githubusercontent.com/2153895/198007007-5d54d9ad-4a44-4c7b-80f2-52226bc9b361.png">


## How to run

### Using Makefile

Start server: `make server` (or `make s`)

Start client: `make client` (or `make c`)

### Using Go directly

Start server:
```bash
go run main.go --server
```

Start client:
```bash
go run main.go --client
```

Or simply run as client (default mode):
```bash
go run main.go
```

### Command-line options

```
Usage:
  main [OPTIONS]

Application Options:
  -s, --server  Run as server
  -c, --client  Run as client
  -a, --addr=   Server address (for client mode) or bind address (for server mode) (default: 127.0.0.1:8080)
      --debug   Enable debug mode

Help Options:
  -h, --help    Show this help message
```

## Controls

### Keyboard Controls

- **Arrow Keys** or **WASD**: Move cursor
- **Space**: Open cell
- **Enter**: Cycle flag markers (Flag → Guess → Hidden)
- **Ctrl+C**: Quit game
