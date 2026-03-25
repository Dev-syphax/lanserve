"""
LANserve — HTTP file server for local network use.

Usage:
    python server.py                      # serves current directory on port 8080
    python server.py --port 9000          # custom port
    python server.py --dir ~/Downloads    # custom directory
    python server.py --host 0.0.0.0       # explicit bind address
"""

import email
import http.server
import io
import os
import socket
import sys
import urllib.parse
from socketserver import ThreadingMixIn

# importlib.resources loads package data correctly after pip install,
# whether the package is installed, run from source, or zipped.
try:
    from importlib.resources import files as _res_files          # Python 3.9+
    def _read_static(filename: str) -> bytes:
        return _res_files("lanserve").joinpath("static").joinpath(filename).read_bytes()
except ImportError:
    import importlib.resources as _ir                             # Python 3.8
    def _read_static(filename: str) -> bytes:
        with _ir.open_binary("lanserve.static", filename) as f:
            return f.read()

def _read_template() -> str:
    return _read_static("template.html").decode("utf-8")

# ── Config (overridden by CLI args in run()) ───────────────────────────────────

DEFAULT_PORT = 8080
DEFAULT_HOST = "0.0.0.0"
DIRECTORY    = os.path.abspath(".")   # overridden in run()


# ── Helpers ───────────────────────────────────────────────────────────────────

def get_local_ip() -> str:
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    try:
        s.connect(("8.8.8.8", 80))
        return s.getsockname()[0]
    except Exception:
        return "localhost"
    finally:
        s.close()


def human_size(size: int) -> str:
    for unit in ("B", "KB", "MB", "GB"):
        if size < 1024:
            return f"{size:.1f} {unit}"
        size /= 1024
    return f"{size:.1f} TB"


def file_icon(name: str, is_dir: bool) -> str:
    if is_dir:
        return "📁"
    ext = name.rsplit(".", 1)[-1].lower() if "." in name else ""
    return {
        "py":   "🐍", "js": "📜", "ts":  "📜", "jsx": "📜", "tsx": "📜",
        "json": "📋", "md": "📝", "txt": "📝", "html":"🌐", "css": "🎨",
        "png":  "🖼️", "jpg":"🖼️", "jpeg":"🖼️", "gif": "🖼️", "svg": "🖼️", "webp":"🖼️",
        "mp4":  "🎬", "mov":"🎬", "mkv": "🎬", "mp3": "🎵", "wav": "🎵",
        "zip":  "📦", "tar":"📦", "gz":  "📦", "pdf": "📄",
        "sh":   "⚙️", "env":"🔑",
    }.get(ext, "📄")


def load_template() -> str:
    """Load template.html from the installed package data."""
    try:
        return _read_template()
    except Exception as e:
        raise FileNotFoundError(
            f"Could not load template.html from the lanserve package: {e}\n"
            "Try reinstalling: pip install --force-reinstall lanserve"
        ) from e


def render_template(slots: dict) -> str:
    """Replace <!-- SLOT:name --> markers with dynamic values."""
    html = load_template()
    for key, value in slots.items():
        html = html.replace(f"<!-- SLOT:{key} -->", value)
    return html


# ── Multipart parser (no cgi module) ──────────────────────────────────────────

def parse_multipart(headers, body: bytes) -> tuple:
    """
    Parse multipart/form-data without the deprecated cgi module.
    Returns:
        fields : {name: str}
        files  : {name: (filename, bytes)}
    """
    content_type = headers.get("Content-Type", "")
    boundary = None
    for part in content_type.split(";"):
        part = part.strip()
        if part.startswith("boundary="):
            boundary = part[len("boundary="):].strip('"')
            break
    if not boundary:
        return {}, {}

    fields: dict = {}
    files:  dict = {}
    boundary_bytes = ("--" + boundary).encode()

    for raw_part in body.split(boundary_bytes)[1:]:   # skip preamble
        if raw_part in (b"--\r\n", b"--", b""):
            continue
        if raw_part.startswith(b"\r\n"):
            raw_part = raw_part[2:]
        if raw_part.endswith(b"\r\n"):
            raw_part = raw_part[:-2]
        if b"\r\n\r\n" not in raw_part:
            continue

        raw_headers, part_body = raw_part.split(b"\r\n\r\n", 1)
        msg         = email.message_from_bytes(raw_headers + b"\r\n\r\n")
        disposition = msg.get("Content-Disposition", "")

        name = filename = None
        for item in disposition.split(";"):
            item = item.strip()
            if item.startswith("name="):
                name = item[5:].strip('"')
            elif item.startswith("filename="):
                filename = item[9:].strip('"')

        if name is None:
            continue
        if filename is not None:
            files[name] = (filename, part_body)
        else:
            fields[name] = part_body.decode("utf-8", errors="replace")

    return fields, files


# ── Request handler ────────────────────────────────────────────────────────────

_STATIC_MIME = {
    ".css": "text/css; charset=utf-8",
    ".js":  "application/javascript; charset=utf-8",
}

class LANserveHandler(http.server.SimpleHTTPRequestHandler):

    def do_GET(self):
        """Intercept /_static/* requests and serve from package data."""
        if self.path.startswith("/_static/"):
            filename = self.path[len("/_static/"):].split("?")[0]
            ext  = os.path.splitext(filename)[1].lower()
            mime = _STATIC_MIME.get(ext, "application/octet-stream")
            try:
                data = _read_static(filename)
                self.send_response(200)
                self.send_header("Content-Type", mime)
                self.send_header("Content-Length", str(len(data)))
                self.send_header("Cache-Control", "public, max-age=3600")
                self.end_headers()
                self.wfile.write(data)
            except Exception:
                self.send_error(404, f"Static file not found: {filename}")
            return
        super().do_GET()

    def list_directory(self, path):
        try:
            items = sorted(
                os.listdir(path),
                key=lambda a: (not os.path.isdir(os.path.join(path, a)), a.lower()),
            )
        except OSError:
            self.send_error(404, "Cannot list directory")
            return None

        displaypath = urllib.parse.unquote(self.path)

        # Folder dropdown — top-level subdirs only
        subdirs = sorted(
            d for d in os.listdir(DIRECTORY)
            if os.path.isdir(os.path.join(DIRECTORY, d)) and not d.startswith(".")
        )
        folder_options = "".join(
            f'<option value="{d}">{d}/</option>' for d in subdirs
        )

        # File list rows
        rows = []
        for name in items:
            fullname = os.path.join(path, name)
            is_dir   = os.path.isdir(fullname)
            link     = urllib.parse.quote(name) + ("/" if is_dir else "")
            icon     = file_icon(name, is_dir)
            size_str = "" if is_dir else human_size(os.path.getsize(fullname))
            css      = "file-item dir" if is_dir else "file-item"

            # Build the relative path for DELETE requests
            rel_path = (self.path.rstrip("/") + "/" + urllib.parse.quote(name)).lstrip("/")
            delete_btn = (
                f'<button class="btn-delete" '
                f'onclick="event.preventDefault();deleteFile(\'/{rel_path}\',\'{name}\')" '
                f'title="Delete">✕</button>'
                if not is_dir else ""
            )

            rows.append(
                f'<li><a class="{css}" href="{link}">'
                f'<span class="file-icon">{icon}</span>'
                f'<span class="file-name">{name}{"/" if is_dir else ""}</span>'
                f'<span class="file-size">{size_str}</span>'
                f'</a>{delete_btn}</li>'
            )

        html    = render_template({
            "displaypath":    displaypath,
            "folder_options": folder_options,
            "file_items":     "\n    ".join(rows),
        })
        encoded = html.encode("utf-8", "surrogateescape")
        buf     = io.BytesIO(encoded)

        self.send_response(200)
        self.send_header("Content-Type",   "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(encoded)))
        self.end_headers()
        return buf

    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length)

        fields, files = parse_multipart(self.headers, body)
        target_dir    = fields.get("target_folder", ".")

        if "file" in files:
            filename, data = files["file"]
            if filename:
                fn          = os.path.basename(filename)
                upload_path = os.path.join(DIRECTORY, target_dir, fn)
                os.makedirs(os.path.dirname(upload_path), exist_ok=True)
                with open(upload_path, "wb") as fp:
                    fp.write(data)

        # 204 → XHR upload handler detects success without a page reload
        self.send_response(204)
        self.end_headers()

    def do_DELETE(self):
        """DELETE /path/to/file — triggered by the UI's delete button."""
        rel  = urllib.parse.unquote(self.path.lstrip("/"))
        full = os.path.realpath(os.path.join(DIRECTORY, rel))

        # Safety: refuse to delete anything outside DIRECTORY
        if not full.startswith(os.path.realpath(DIRECTORY) + os.sep):
            self.send_response(403)
            self.end_headers()
            return

        try:
            if os.path.isfile(full):
                os.remove(full)
                self.send_response(204)
            else:
                self.send_response(404)
        except OSError as e:
            self.send_response(500)
            self.send_header("Content-Type", "text/plain")
            self.end_headers()
            self.wfile.write(str(e).encode())
            return

        self.end_headers()

    def end_headers(self):
        self.send_header("Access-Control-Allow-Origin",  "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        self.send_header("Accept-Ranges", "bytes")
        super().end_headers()

    def do_OPTIONS(self):
        self.send_response(204)
        self.end_headers()

    def log_message(self, fmt, *args):
        # Suppress noisy favicon / asset requests.
        # Guard: args[0] may be an HTTPStatus enum (not a string) when
        # called from log_error, so check type before doing `in`.
        if args and isinstance(args[0], str):
            if any(ext in args[0] for ext in (".ico", ".png", ".css")):
                return
        super().log_message(fmt, *args)


# ── Threaded HTTP server ───────────────────────────────────────────────────────

class ThreadedHTTPServer(ThreadingMixIn, http.server.HTTPServer):
    """One thread per request so large uploads don't block browsing."""
    daemon_threads = True

    def handle_error(self, request, client_address):
        exctype, value = sys.exc_info()[:2]
        # Silently drop client-side disconnects — these are normal when:
        if exctype in (BrokenPipeError, ConnectionResetError, ConnectionAbortedError):
            return
        msg = str(value)
        if any(s in msg for s in ("Broken pipe", "Connection reset",
                                  "Connection aborted", "WinError 10053",
                                  "WinError 10054")):
            return
        super().handle_error(request, client_address)


# ── Entry point ────────────────────────────────────────────────────────────────

def run(host: str = DEFAULT_HOST, port: int = DEFAULT_PORT, directory: str = "."):
    """
    Start the HTTP file server.

    Args:
        host:      Bind address   (default: 0.0.0.0)
        port:      Port number    (default: 8080)
        directory: Directory to serve (default: current directory)

    Called by __main__.py (the `lanserve` CLI) or directly in scripts:
        from lanserve.server import run
        run(port=9000, directory="/tmp/share")
    """
    global DIRECTORY
    DIRECTORY = os.path.abspath(directory)

    def handler_factory(*a, **kw):
        return LANserveHandler(*a, directory=DIRECTORY, **kw)

    server = ThreadedHTTPServer((host, port), handler_factory)
    ip     = get_local_ip()

    print(f"\n LANserve running!")
    print(f"   Warning: Only use on trusted networks!")
    print(f"   Local:   http://localhost:{port}")
    print(f"   Network: http://{ip}:{port}")
    print(f"   Serving: {DIRECTORY}")
    print(f"   (Ctrl+C to stop)\n")

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n Stopped.")


if __name__ == "__main__":
    # Standalone usage: python server.py [--port] [--dir] [--host]
    import argparse
    parser = argparse.ArgumentParser(description="LANserve HTTP server")
    parser.add_argument("--port", "-p", type=int, default=DEFAULT_PORT)
    parser.add_argument("--host", default=DEFAULT_HOST)
    parser.add_argument("--dir", "-d", default=".")
    a = parser.parse_args()
    run(host=a.host, port=a.port, directory=a.dir)