### Simple Key/Value Store in Go

A simple TCP-based key-value store server written in Go.

#### Run program
```bash
go run main.go
```

#### Available Commands
- `GET <key>` : Retrieve the value for the specified key.
- `PUT <key> <value>` : Store a value for the specified key.
- `DEL <key>` : Delete the key-value pair.
- `LIST` : List all keys and values.

#### Usage
After running the program, connect to the server using TELNET or nc on localhost:8080.

**Examples:**

- Using nc: `echo "GET example" | nc localhost 8080`
- Using telnet: `telnet localhost 8080` then type commands like `GET example` and press Enter.
