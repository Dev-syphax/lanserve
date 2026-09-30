# LANserve

A lightweight local file server written in Go.

## Installation

### Using Go

If you have Go installed:

```bash
go install github.com/Dev-syphax/lanserve@latest
```

Then run:

```bash
lanserve
```

### Using the install script

On Linux or macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/Dev-syphax/lanserve/main/install.sh | bash
```

Then:

```bash
lanserve
```

---

## Clone & Build

Clone the repository:

```bash
git clone https://github.com/Dev-syphax/lanserve.git
cd lanserve
```

Build the application:

```bash
go build -o lanserve .
```

Run it:

```bash
./lanserve
```

---

## Usage

Start LANserve in the current directory:

```bash
lanserve
```

By default, the server runs on port `8080`.

That's it. Open the printed URL in any browser on your network:

```
┌──────────────────────────────────────────────────────────┐
│    LANserve — Local File Server                          │
├──────────────────────────────────────────────────────────┤
│   Local:   http://localhost:8080                         │
│   Network: http://192.168.x.x:8080                       │
│   Serving: /your/current/directory                       │
│   Auth:    Disabled                                      │
└──────────────────────────────────────────────────────────┘
```

To access the server from another device on the same network, use the network address displayed by LANserve.

### Serve a specific directory

```bash
lanserve --dir ./files
```

### Change the port

```bash
lanserve --port 3000
```

Then open:

```text
http://localhost:3000
```

### Specify the host

```bash
lanserve --host 0.0.0.0
```

### Enable access protection

You can require an access code for uploads and deletes:

```bash
lanserve --code my-secret-code
```

### Combine options

```bash
lanserve  --dir ./Downloads  --port 8080  --host 0.0.0.0  --code my-secret-code
```

---

## Options

| Option   | Short | Default   | Description                     |
| -------- | ----- | --------- | ------------------------------- |
| `--port` | `-p`  | `8080`    | Port to listen on               |
| `--host` | —     | `0.0.0.0` | Address to bind to              |
| `--dir`  | `-d`  | `.`       | Directory to serve              |
| `--code` | `-c`  | —         | Access code for uploads/deletes |

---

## Examples

Serve your current directory:

```bash
lanserve
```

Serve a shared folder:

```bash
lanserve -d ~/shared
```

Run on port `3000`:

```bash
lanserve -p 3000
```

Serve a folder with an access code:

```bash
lanserve -d ./files -c my-secret-code
```

---

## License

[MIT](LICENSE)
