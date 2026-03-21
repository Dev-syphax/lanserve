"""
Entry point for `python -m lanserve` and the `lanserve` console script.
"""

import argparse
import os
import sys


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        prog="lanserve",
        description="LANserve — local network file server with browser UI",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  lanserve                          Serve current directory on port 8080
  lanserve --port 9000              Custom port
  lanserve --dir ~/Downloads        Serve a specific folder
        """,
    )
    parser.add_argument("--port", "-p", type=int, default=None,
                        help="Port to listen on (default: 8080 for HTTP)")
    parser.add_argument("--dir", "-d", default=".",
                        help="Directory to serve (default: current directory)")
    parser.add_argument("--host", default="0.0.0.0",
                        help="Address to bind to (default: 0.0.0.0)")
    parser.add_argument("--version", "-v", action="version",
                        version=f"%(prog)s {_get_version()}")
    return parser.parse_args()


def _get_version() -> str:
    try:
        from lanserve import __version__
        return __version__
    except ImportError:
        return "unknown"


def main():
    args = parse_args()

    # Validate directory
    directory = os.path.abspath(args.dir)
    if not os.path.isdir(directory):
        print(f"Error: '{directory}' is not a directory.", file=sys.stderr)
        sys.exit(1)
    else:
        # ── HTTP mode ─────────────────────────────────────────────────────────

        from lanserve.server import run as run_http
        port = args.port or 8080
        run_http(host=args.host, port=port, directory=directory)


if __name__ == "__main__":
    main()
