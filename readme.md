### Simple Key/val Store in Go

A simple TCP-based key-val store server written in Go.

#### Run program
```bash
go run main.go
```

#### Available Commands
- `GET <key>` : Retrieve the val for the specified key.
- `PUT <key> <val>` : Store a val for the specified key.
- `DEL <key>` : Delete the key-val pair.
- `LIST` : List all keys and vals.

#### Usage
After running the program, connect to the server using TELNET or nc on localhost:8080.

**Examples:**

- Using nc: `echo "GET example" | nc localhost 8080`
- Using telnet: `telnet localhost 8080` then type commands like `GET example` and press Enter.
