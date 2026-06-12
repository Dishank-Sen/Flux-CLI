# Flux CLI

A command-line interface that sends user commands to a Flux daemon (IPC RPC) to manage a local Flux repository, record file changes, generate SSH keys, and push snapshots/deltas to a remote server.

# Overview

- What problem this project solves  
  Provides a local CLI to initialize and manage a Flux repository (.flux), record file change history, generate/manage SSH keys, and push repository snapshots/deltas to a server. The CLI acts as a controller that forwards user actions to a separate flux daemon over an IPC/RPC channel.

- High-level architecture  
  - CLI (this repository) — user-facing binary (`flux`) built in Go that:
    - manages .flux repo files (init, create config, create .flowignore, file tree)
    - talks to the flux daemon using RPC (lesismal/arpc) over:
      - Unix domain socket: /tmp/flux.sock (Linux/macOS)
      - TCP localhost:43899 (Windows / fallback)
    - performs client-side tasks (create zip payloads, SSH key generation, config file management)
    - invokes daemon RPC endpoints: "/init", "/start", "/login" (via arpc)
  - Flux daemon (not included here) — expected to listen on the IPC channel and implement the RPC endpoints for the actual recording/operations.
  - Remote server — HTTP endpoint used by `flux push` (default: http://localhost:3000/api/v1/push).

- Main technologies used
  - Go (module: github.com/Dishank-Sen/Flux-CLI)
  - CLI framework: github.com/spf13/cobra
  - RPC: github.com/lesismal/arpc (IPC client)
  - SSH keys: golang.org/x/crypto/ssh, ed25519
  - Zip + multipart HTTP uploads
  - Standard library: filesystem, os, http, encoding/json

# Features

- Repository initialization:
  - `flux init` — initializes `.flux` directory layout and default config (creates `.flux/config.json`, directories `.flux/files`, `.flux/history`, `.flux/root-timeline`).
  - Reinitialization detection and safe re-init behavior.

- Recording control:
  - `flux start` — instructs daemon to start recording changes (RPC call to "/start").

- Authentication:
  - `flux login` — RPC call to "/login" that forwards repository username.

- Snapshot/push:
  - `flux push` — creates file-tree and zips `.flux/history`, `.flux/files`, `.flux/root-timeline` and uploads as multipart/form-data to a server endpoint (default hard-coded to http://localhost:3000/api/v1/push).

- Configuration:
  - `flux set` — set repository fields: username, remote URL, code threshold, debounce time (writes `.flux/config.json`).
  - `flux config -l` - print current config.

- SSH key management:
  - `flux genk` — generate ed25519 SSH key pair and store in user key dir (~/.local/share/flux-daemon/ssh-keys by default). Can also sync existing keys with `--sync <path>`.

- Utilities:
  - `flux init-ignore` — create a default `.flowignore`.
  - `flux showk` — print SSH public and private keys as configured in `.flux/config.json`.

# Architecture

- Major components
  - CLI command registry (cli.Register + Registered map) — each command registers a factory that returns a *cobra.Command
  - IPC client layer:
    - `cli.DialIPC()` & client.dialIPC() establish connection via unix socket (/tmp/flux.sock) for linux/darwin or tcp 127.0.0.1:43899 for windows.
    - Uses arpc.NewClient(...) to call endpoints like "/init", "/start", "/login" with JSON payload strings and expect JSON string responses.
  - Config & local storage:
    - `.flux/config.json` — types.Config: WorkingDir, Repository{UserName, RemoteUrl}, Recorder{DebounceTime, CodeThreshold}, SSHKeys{PublicKeyPath, PrivateKeyPath}
    - `.flux/files/fileTree.json` — generated file tree JSON (types.FileTree -> []*types.Node)
    - `.flux/history` and `.flux/root-timeline` — contain event files (Write events etc.)
  - Push flow:
    - `flux push` validates config & SSH keys, authenticates via arpc login call, builds `.flux` file tree and then:
      - zips `.flux/history` (optionally filters by .flowignore),
      - zips `.flux/files` and `.flux/root-timeline`,
      - creates multipart form with metadata and the zip parts and POSTs to the endpoint.
  - SSH key generation:
    - Uses ed25519 keys and marshals to OpenSSH format/private PEM via ssh.MarshalPrivateKey / ssh.MarshalAuthorizedKey (see genk.go).

- Data flow (push example)
  1. User runs `flux push`.
  2. CLI reads `.flux/config.json`, checks SSH keys and username/remoteUrl.
  3. CLI calls daemon RPC `/login` to authenticate user (arpc).
  4. CLI creates file tree JSON (.flux/files/fileTree.json).
  5. CLI zips relevant .flux directories (history, files, root-timeline) filtering by `.flowignore`.
  6. CLI POSTs multipart/form-data to configured endpoint (default http://localhost:3000/api/v1/push).
  7. Server responds; CLI logs status/body.

# Installation

- Prerequisites
  - Go toolchain (module indicates Go 1.25 in go.mod).
  - curl (optional, for docs/install.sh).
  - For building multi-arch binaries: set GOOS/GOARCH or use provided build.sh.

- Build (local)
  - Simple build:
    ```sh
    go build -o ./bin/flux ./cmd/flux
    ```
    Note: the repository's main is at `cmd/flux/main.go`; Makefile and build.sh reference `cmd/flux` — path inconsistency (Needs clarification).
  - Multi-arch via script:
    ```sh
    chmod +x build.sh && ./build.sh
    ```
  - Makefile targets:
    - `make build`  (Makefile references `./cmd/flux` which does not exist in this tree; see Limitations → Needs clarification)

- Install (system)
  - After building: `sudo cp ./bin/flux /usr/local/bin/flux`
  - Or use `docs/install.sh` to download prebuilt release artifacts (script targets Linux only).

# Configuration

- Config file (.flux/config.json)
  - Located inside repository root under `.flux/config.json`.
  - JSON structure (types.Config). Example:

```json
{
  "WorkingDir": "/abs/path/to/repo",
  "Repository": {
    "UserName": "",
    "RemoteUrl": ""
  },
  "Recorder": {
    "DebounceTime": 3,
    "CodeThreshold": 10
  },
  "SSHKeys": {
    "PublicKeyPath": "",
    "PrivateKeyPath": ""
  }
}
```

  - Defaults created by `flux init` / CreateConfig.

- Ignore rules
  - `.flowignore` at repo root controls which repo paths are excluded from file-tree and history zips.
  - The CLI always ignores `.flux`, `.git`, and `node_modules` (strictIgnore map).

- SSH key storage
  - Default directory for generated keys: `~/.local/share/flux-daemon/ssh-keys` (functions in constants/constants.go).
  - `flux genk` will place `id_ed25519` (private) and `id_ed25519.pub` (public) there and update `.flux/config.json`.

- IPC paths and network
  - Unix socket: `/tmp/flux.sock` (linux, darwin)
  - TCP fallback/Windows: `127.0.0.1:43899`
  - Push endpoint (hard-coded in code): `http://localhost:3000/api/v1/push`

# Usage

- Initialize repository
  - `flux init`

- Configure repository
  - `flux set --username alice --remoteUrl example.com/alice/myrepo.flux`
  - `flux config -l`

- Generate or sync SSH keys
  - `flux genk`
  - `flux genk --sync /path/to/ssh-keys-dir`

- Start recording (delegates to daemon)
  - `flux start`

- Authenticate (delegates to daemon)
  - `flux login`

- Push to server
  - `flux push`
  - (The push client posts to http://localhost:3000/api/v1/push by default.)

- Other utilities
  - `flux init-ignore` — create default .flowignore
  - `flux showk` — display configured SSH keys

Example workflow:

```sh
flux init
flux set --username alice --remoteUrl example.com/alice/myrepo.flux
flux genk
flux start
# work on files; daemon records changes
flux push
```

# Project Structure

- `cmd/flux/main.go` — CLI entrypoint (constructs root cobra.Command)
- `cli/` — cobra command implementations and registration
  - `cli.go` — command registry and DialIPC helper
  - `root.go` — root cobra command and persistent pre-run checks
  - `init.go`, `start.go`, `push.go`, `login.go`, `genk.go`, `set.go`, `config.go`, `showk.go`, `init-ignore.go` — command implementations
  - `initDir/`, `initFiles/` — helpers to register directories/files created on init
  - `utils/` — helpers used by CLI init/push/file-tree creation
- `client/` — small arpc client wrapper (NewClient)
- `types/` — JSON-serializable structs used by CLI (Create, Remove, Write, Repository, Config, Node/FileTree, Metadata)
- `constants/` — helper functions for default SSH key paths and constants
- `utils/` — shared filesystem and config utilities
- `utils/logger/` — lightweight colored logging wrapper
- `bin/` — prebuilt binaries (checked into repo)
- `build.sh`, `Makefile` — build scripts
- `docs/install.sh` — installer script (Linux-only)
- `go.mod`, `go.sum` — module files
- `debug/` — small tests and helpers used during development

# Technical Details

- IPC & RPC
  - Uses `lesismal/arpc` to create an RPC client that dials via a function that returns `net.Conn` (unix socket or TCP).
  - Commands call daemon RPC endpoints using `arpc.Call` with JSON payload strings and expect JSON string responses.

- File tree & history
  - File tree is built by scanning `WorkingDir` and writing `.flux/files/fileTree.json` using `types.Node` representation (Name, Path, IsDir, Size, Children).
  - History files (`.flux/history`) appear to store serialized events (`types.Write` etc.). `push` filters these files by `.flowignore` before upload.

- SSH keys
  - ed25519 generation via `crypto/ed25519`; public keys marshaled with `golang.org/x/crypto/ssh`.
  - Private key stored as PEM OpenSSH private key; public as authorized_keys-style line.

- Push protocol
  - The client builds a multipart/form-data request with:
    - `metadata` form field containing `types.Metadata` JSON
    - form files: `fileTree` (fileTree.zip), `history` (history.zip), `root-timeline` (root-timeline.zip)
  - Upload performed with `net/http` Client.Do

- Concurrency & cancellation
  - Many operations accept `context.Context` and use cancellation checks when creating directories/files (`utils.CreateDir/CreateConfig`).
  - `push` uses an `io.Pipe` and a goroutine to stream zip contents into the HTTP request body.

# Development

- Build
  - `go build -o ./bin/flux ./cmd/flux`
  - or use `build.sh` for multiple platforms (requires correct cmd path; see Limitations)

- Run tests
  - `go test ./...`
  - Minimal tests exist in `debug/debug_test.go` (TestDebug, TestPromptEmail, TestFileTree).

- Debugging
  - CLI logs via `utils/logger` with colored INFO/WARN/ERROR messages.
  - Use `go test` and `go run` to step through commands. Many commands require a running daemon reachable at `/tmp/flux.sock` or `127.0.0.1:43899`.

# Limitations

- Missing daemon implementation: this repository implements the CLI and client-side logic only. The flux daemon that implements RPC endpoints (`/init`, `/start`, `/login`, etc.) is not present here — CLI RPC calls will fail unless a compatible daemon is running and listening on the expected IPC endpoints. (Needs clarification: intended daemon location or repository.)
- Hard-coded push endpoint: `flux push` posts to `http://localhost:3000/api/v1/push` in code (cli/push.go). There is no configuration hook in the CLI for this endpoint.
- Binaries in `bin/`: prebuilt binaries are checked in; verify compatibility and trust before use.
- Authentication protocol and server API contract are not defined here beyond the CLI's expectations (strings JSON for login, multipart shape for push). Server-side validation/response formats are not included (Needs clarification).

# Future Improvements

- Make push endpoint configurable via `.flux/config.json` or environment variable.
- Add server-client protocol docs and tests (specify RPC method names, payload formats, error cases).
- Add unit & integration tests around push flow and RPC interactions; mock arpc client for CLI tests.
- Add support for Windows-specific sockets or better fallback documentation.
- Improve error reporting from RPC calls and add retries/timeouts for push operation.
- Add more robust ignore rules parsing (glob patterns) and support for per-directory ignore files.

# Contact / Notes

- This README is based strictly on the source found in this repository (Go sources, scripts, and build files). Where behavior or components could not be determined from code (daemon implementation, server contract, or intended build target name), the README explicitly marks "Needs clarification".
- To run the CLI successfully, a compatible Flux daemon must be running and reachable via `/tmp/flux.sock` (unix) or `127.0.0.1:43899` (tcp). The server endpoint for push is currently hard-coded in the CLI.
