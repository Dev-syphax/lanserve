# LANserve

A lightweight local file server with a clean browser UI — browse, upload, and delete files over your WiFi network from any device.

No dependencies. No config. Just run it.

---

## Install

```bash
pip install lanserve
```

---

## Quick start

```bash
lanserve
```

That's it. Open the printed URL in any browser on your network:

```
 LANserve running!
   Local:   http://localhost:8080
   Network: http://192.168.x.x:8080
   Serving: /your/current/directory
```

On mobile, open the **Network** URL. On desktop, use either.

---

## Features

- **Browse** your file system from any device on the same network
- **Upload** files via drag-and-drop or file picker, with a real-time progress bar
- **Delete** files directly from the UI
- **Optional access code** to restrict uploads/deletes while browsing stays open
- **Choose upload folder** from a dropdown
- **File type icons** and human-readable file sizes
- **Threaded** — large uploads don't freeze browsing
- **Zero external dependencies** — Python 3.11+ standard library only

---

## Usage

```
lanserve [options]

Options:
  --port PORT, -p PORT    Port to listen on        (default: 8080)
  --dir DIR,  -d DIR      Directory to serve        (default: current directory)
  --host HOST             Address to bind to        (default: 0.0.0.0)
  --code CODE, -c CODE    Access code required for uploads/deletes (default: none, disabled)
  --version, -v           Show version and exit
```

### Examples

```bash
# Serve a specific folder
lanserve --dir ~/Downloads

# Use a different port
lanserve --port 9000

# Serve Downloads on port 9000
lanserve --dir ~/Downloads --port 9000

# Require an access code for uploads/deletes (browsing stays open to everyone)
lanserve --code mySecret123

```

---

## Run without installing

```bash
git clone https://github.com/Dev-syphax/lanserve.git
cd lanserve
python -m lanserve
```

---

## Security

LANserve is designed for **trusted local networks only** (home, office LAN, dev WiFi).

- By default there is no authentication — anyone on the network can browse, upload, and delete files.
- To restrict uploads and deletes, start the server with `--code`:
  ```bash
  lanserve --code mySecret123
  ```
  Browsing and downloading files stays open to everyone; only uploads and deletes require the code.
- When the code is required, the browser prompts for it the first time you upload or delete something, then remembers you via a session cookie — no need to re-enter it on every action or page reload. The session is cleared whenever the server restarts.
- The code is sent as a plain HTTP header, not encrypted — this is meant to keep casual users on your LAN from uploading/deleting, not as strong authentication. Don't reuse a sensitive password as your access code.
- DELETE requests are path-traversal protected — files outside the served directory cannot be deleted.
- Do **not** expose this server to the public internet.

---

## Requirements

- Python 3.11 or newer
- No dependencies (bundled with Python standard library)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE)
