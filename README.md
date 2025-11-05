# GoTorch — Torchlight Infinite Tracker

GoTorch is a desktop app (Wails v2 + React) and CLI that monitors Torchlight Infinite's `UE_game.log`, detects map runs, and tallies item drops during each run.

## Project layout

- `internal/` — core library
  - `parser` — parses UE log lines into events
  - `tailer` — follows the log file in real time
  - `tracker` — aggregates sessions and tallies drops
  - `app` — Wails backend that exposes the tracker to the UI
- `cmd/cli` — small CLI to validate parsing/tallying on a log file
- `cmd/wails` — Wails entrypoint (also mirrored by root `main.go`)
- `frontend/` — React + Vite UI for the desktop app
- `UE_game.log` — sample game log for local verification

## Prerequisites

- Go (1.24.x recommended; 1.23+ should work)
- Node.js LTS (18+)
- Wails v2 CLI

Install Wails CLI:

```
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

On Linux CI/hosts you may also need GUI build deps (GTK/WebKit):

```
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev build-essential libglib2.0-dev
```

## Quick start (CLI)

Use the included sample log:

```
go run ./cmd/cli --log UE_game.log --once --debug=true
```

Tail your real log (Windows example):

```
go run ./cmd/cli --log "%USERPROFILE%\AppData\LocalLow\XD\Torchlight Infinite\UE_game.log" --debug=true
```

Useful flags:
- `--from-start=true` — process the whole file from the beginning
- `--once` — process once and exit with a summary

## Run the desktop app (Wails)

1) Install deps

```
# Frontend
cd frontend && npm install
cd ..

# Go deps
go mod tidy
```

2) Dev run

```
wails dev
```

If you previously saw "no Go files in" when running `wails dev` it’s because Wails expects a `main.go` in the project root. This repository includes `main.go` at the root to satisfy that.

3) Build package

```
wails build
```

Artifacts will be placed in `build/bin`.

## Troubleshooting

- Wails build fails on Linux: ensure GTK/WebKit dependencies are installed (see above).
- "no Go files in ...": ensure you run `wails dev` at the repo root which now contains `main.go`.
- Frontend dev server port conflicts: update `wails.json -> build.devServerURL` or run `vite` on a free port.

## CI/CD (GitHub Actions)

The repo includes a split pipeline under `.github/workflows/`:

- `style.yml` — code style and static checks (`gofmt`, `go vet`, optional `golangci-lint`)
- `tests.yml` — unit tests with coverage (enforces ≥80% line coverage)
- `security.yml` — Go security scan (`gosec`) and frontend `npm audit` (frontend audit is non-blocking by default)
- `build.yml` — build & package via Wails on Linux/macOS/Windows, plus CLI build; artifacts are uploaded

Coverage is measured via `go test -coverprofile` across all packages. If you change the coverage bar, tweak the threshold in `tests.yml`.

## License

MIT (see your repository choice).
