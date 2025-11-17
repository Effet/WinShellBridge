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
- `--port 9000` bind to custom port (still 127.0.0.1).
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
  "timeout_sec": 120
}
```
Response: `text/plain`, streaming combined stdout/stderr. On non-zero exit it appends a line `[exit error: ...]`.

## Security
- The server only binds `127.0.0.1`; for remote use, SSH port-forward (`ssh -L 8088:127.0.0.1:8088 user@host`).
- Add authentication/whitelisting before exposing beyond localhost; current build is minimal.

## GitHub Actions release
A workflow builds the Windows binary on tags (or manual dispatch) and uploads:
- Artifact: `WinShellBridge_windows_amd64.zip`
- Release upload when running on a tag; set `prerelease` when dispatching manually.

Nightly builds: pushes to `main` (or the scheduled cron) publish a prerelease under the `nightly` tag with the same zipped binary.
