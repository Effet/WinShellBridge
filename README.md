# WinShellBridge

A tiny Windows tray app that lives in the logged-in user session. It exposes a local HTTP UI + REST API so you can start GUI/CLI programs from a browser or via SSH port-forwarding, without PsExec. It registers itself in `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` to start after login.

## Features
- Tray icon with “Open UI” / “Quit” and status.
- Local HTTP UI + REST (`POST /api/run`) with streaming stdout/stderr.
- Autostart on login via HKCU Run (can be disabled with a flag).

## Quick start (Windows)
```sh
go mod tidy          # first time to fetch deps
go run . --open-ui   # starts on http://127.0.0.1:8088
```

Flags:
- `--config /path/to/config.json` set config file (defaults to `%APPDATA%\WinShellBridge\config.json` on Windows, `~/.config/WinShellBridge/config.json` elsewhere).
- `--host 0.0.0.0` bind to all interfaces (default 127.0.0.1; be sure to secure it).
- `--port 9000` bind to custom port.
- `--no-autostart` skip writing the registry entry.
- `--open-ui` open the browser on start.

Build:
```sh
go build -o WinShellBridge.exe
./WinShellBridge.exe --port 8088
```
The first run writes the HKCU Run entry so it starts automatically after login.

## Cross-compiling from macOS to Windows
The tray library needs CGO on Windows. Install a Mingw-w64 toolchain (e.g. `brew install mingw-w64`) and build with the helper Makefile:
```sh
make build-windows               # produces dist/WinShellBridge.exe
# or tweak:
GOARCH=386 make build-windows
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc make build-windows
```
You can also run `GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o dist/WinShellBridge.exe .` directly.

## REST API
`POST /api/run`
```json
{
  "cmd": "notepad.exe",
  "args": ["README.md"],
  "workdir": "C:\\Users\\me",
  "timeout_sec": 120,
  "background": false
}
```
Response: `text/plain`, streaming combined stdout/stderr. On non-zero exit it appends a line `[exit error: ...]`.

Background mode:
- Set `"background": true` to start the process and return immediately with JSON (`202 Accepted`) containing the PID. Output is discarded; the server won’t stream logs.
- If you set `timeout_sec` with `"background": true`, the process will be terminated when the deadline is hit; omit or set `0` for no timeout.
- Child processes spawned by your command are not tracked.

## Config file
Default path: `%APPDATA%\WinShellBridge\config.json` on Windows, or `~/.config/WinShellBridge/config.json` elsewhere. Override with `--config`.

Example:
```json
{
  "host": "0.0.0.0",
  "port": 9000,
  "autostart": false,
  "open_ui": true
}
```
Flag values take precedence when provided.

## Security
- Default bind is `127.0.0.1`. If you set `--host 0.0.0.0`, ensure you are on a trusted network or place it behind a reverse proxy/VPN and add auth/whitelisting.
- For remote use without binding wide, prefer SSH port-forward (`ssh -L 8088:127.0.0.1:8088 user@host`).

## GitHub Actions release
A workflow builds the Windows binary on tags (or manual dispatch) and uploads:
- Artifact: `WinShellBridge_windows_amd64.zip`
- Release upload when running on a tag; set `prerelease` when dispatching manually.

Nightly builds: pushes to `main` (or the scheduled cron) publish a prerelease under the `nightly` tag with the same zipped binary.
