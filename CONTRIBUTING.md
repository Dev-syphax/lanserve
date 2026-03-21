# Contributing to LANserve

First off — thank you for taking the time to contribute. LANserve is a small, focused tool and every improvement matters, whether it's a typo fix or a new feature.

---

## Philosophy

LANserve has one hard rule: **zero required dependencies**. The HTTP server must run on a clean Python 3.11+ install with nothing extra. Any contribution that adds a required dependency to `server.py` will not be merged. Keep it simple, keep it portable.

---

## Getting started

```bash
# Fork the repo on GitHub, then clone your fork
git clone https://github.com/Dev-syphax/lanserve.git
cd lanserve

# Install in editable mode (no build step needed)
pip install -e .

# Run it on the project folder itself to test
lanserve --dir .
```

---

## Project structure

```
lanserve/
├── src/
│   ├── __init__.py       ← version and package metadata
│   ├── __main__.py       ← CLI entry point (`python -m lanserve`)
│   ├── server.py         ← HTTP server (stdlib only)
│   └── template.html     ← browser UI (HTML/CSS/JS)
├── pyproject.toml
└── .github/workflows/
    ├── ci.yml            ← runs on every push and PR
    └── publish.yml       ← auto-publishes to PyPI on version tags
```

**Key convention:** `server.py` and `template.html` are strictly separated. No HTML strings in Python files, no Python logic in the template. Dynamic values are injected via `<!-- SLOT:name -->` comment markers.

---

## Ways to contribute

**Good first issues:**
- Bug fixes with a clear reproduction case
- Improving error messages
- Documentation and README improvements
- Adding or improving CLI help text

**Larger contributions:**
- New features (open an issue to discuss before coding)
- Test coverage
- Windows / macOS compatibility improvements

---

## Code style

- Follow PEP 8
- Use type hints where they aid clarity
- Keep functions small and focused
- Comment the *why*, not the *what*

---

## Submitting a pull request

1. Create a branch from `main`:
   ```bash
   git checkout -b fix/describe-your-fix
   ```
2. Make your changes
3. Test manually — run the server and verify the affected feature in a browser, on both desktop and mobile if relevant
4. Commit with a clear message:
   ```
   fix: suppress log_message crash on HTTPStatus enum
   feat: add --no-delete flag to disable file deletion from UI
   ```
5. Push and open a pull request against `main` with a description of what and why

---

## Reporting bugs

Please open an issue and include:

- Your OS and Python version (`python --version`)
- How you installed LANserve (`pip install` or cloned)
- The full traceback from the terminal
- Steps to reproduce

---

## Questions

Open a [GitHub Discussion](../../discussions) or file an issue — happy to help.
