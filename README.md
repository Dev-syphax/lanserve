# LANserve

[![PyPI version](https://img.shields.io/pypi/v/lanserve.svg)](https://pypi.org/project/lanserve/)
[![Python versions](https://img.shields.io/pypi/pyversions/lanserve.svg)](https://pypi.org/project/lanserve/)
[![Downloads](https://img.shields.io/pypi/dm/lanserve.svg)](https://pypi.org/project/lanserve/)
[![License](https://img.shields.io/github/license/Dev-syphax/lanserve.svg)](LICENSE)

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
- **Choose upload folder** from a dropdown
- **File type icons** and human-readable file sizes
- **Threaded** — large uploads don't freeze browsing
- **Optional access code** — lock uploads/deletes behind a shared code, browsing stays open
- **Zero external dependencies** — Python 3.8+ standard library only

---

## Usage

```
lanserve [options]

Options:
  --port PORT, -p PORT    Port to listen on        (default: 8080)
  --dir DIR,  -d DIR      Directory to serve        (default: current directory)
  --host HOST             Address to bind to        (default: 0.0.0.0)
  --code CODE, -c CODE    Access code for uploads/deletes (default: none, disabled)
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

# Require a code before anyone can upload or delete
lanserve --code mySecret123
```

---

## Access codes

By default, anyone on the network can browse, upload, and delete — no login required.

If you want to lock down _uploads and deletes_ while keeping _browsing_ open, start LANserve with `--code`:

```bash
lanserve --code mySecret123
```

- Browsing files still works for anyone on the network, no code needed.
- The first upload or delete from a device will prompt for the code.
- Once entered correctly, that device gets a signed, `HttpOnly` session cookie (valid for the rest of the session) so it won't be asked again.
- The code and any active sessions reset every time you restart the server.

This isn't meant to be strong authentication for hostile networks — it's a lightweight gate against "oops, someone accidentally deleted my files" on a shared WiFi.

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

- Without `--code`, there is no authentication — anyone on the network can browse, upload, and delete files.
- With `--code`, uploads and deletes require the code; browsing is still open to everyone.
- DELETE requests are path-traversal protected — files outside the served directory cannot be deleted.
- Do **not** expose this server to the public internet.

---

## Requirements

- Python 3.8 or newer
- No dependencies (bundled with Python standard library)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE)
