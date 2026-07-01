# Go TCP Chat Room

A minimal chat room server built from scratch in Go using raw TCP sockets — no websockets, no frameworks, just `net.Listen` and goroutines.

## Features

- Password-protected room join
- Broadcasts messages to all connected clients
- One goroutine pair (reader/writer) per client, non-blocking broadcast so a slow client can't stall the room
- Thread-safe room membership via a mutex

## Run it

```bash
go run main.go
```

The server listens on `:8090`. Connect with any raw TCP client, e.g.:

```bash
nc localhost 8090
```

On connect, you'll be prompted to join:

```
To join a room use "JOIN <room-id> <room-password> <your-name>"
JOIN roomid roompass yourname
You Joined the Room
```

Once joined, anything you type is broadcast to every other client in the room.

## How it works

- `main.go` accepts incoming connections and spins up a `handleClient` goroutine per connection.
- Each client first completes the join handshake synchronously (so only one goroutine ever reads the socket at a time), then a dedicated `clientReader`/`clientWriter` pair takes over.
- Incoming messages go through a shared `messages` channel to a single broadcaster goroutine, which fans them out to all room members.

## Notes

This is a learning project exploring raw socket handling and goroutine coordination in Go. Currently supports a single hardcoded room, no authentication beyond the shared room password, and no message history or reconnect handling.
