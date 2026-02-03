# Changelog — Extended Inputs & Build Fixes

Summary of uncommitted changes on `develop` (7 files, +257 / -13 lines).

---

## 1. Build script (`build/build.sh`)

- **nasm optional for macOS**: `nasm` is no longer a hard dependency. It is only required when building FFmpeg locally. If missing, a note is printed; the Homebrew install hint no longer includes `nasm` by default, with a separate line for “if building FFmpeg locally.”
- **FFmpeg build guard**: `build_ffmpeg()` now checks for `nasm` and exits with a clear error if it’s missing.
- **Go workspace isolation**: Both the main `build()` and `buildWin()` paths set `export GOWORK=off` so the Go binary is built in module mode only, ignoring any parent `go.work` and building this repo in isolation.

---

## 2. Extended inputs (up to 12 players)

### Command-line flags (`src/main.go`)

- **`--enable-extended-inputs`**: Enables up to 12 human input slots (P1–P12). P3–P12 can be toggled at runtime.
- **`--input-ipc`**: Read inputs from an external process via a Unix socket (requires `--enable-extended-inputs`).
- **`--input-socket-path <path>`**: Socket path for IPC (default: `/tmp/ikemen-input.sock`).

Help text was updated to document these options.

### Config (`src/config.go`)

- When extended inputs are enabled, `Config.Players` is normalized to 12 instead of clamping to `MaxSimul*2`.

### Input handling (`src/input.go`)

- When extended inputs and IPC mode are on, input for controllers 0–11 can be read from shared state (filled by the IPC path) instead of SDL: button bitmask (14 bits) and 6 axes per slot, guarded by `sys.externalInputMu`.
- Comment updated to mention “IPC” in addition to local human (SDL), replay, network, and AI.

### System state and IPC (`src/system.go`)

- **New `System` fields**: `extendedInputsEnabled`, `inputActive[12]`, `externalInputMode`, `inputSocketPath`, `externalInputChan`, `lastButtonStates[12]`, `lastAxes[12][6]`, `externalInputMu`.
- **Init**: If extended inputs are enabled, `commandLists` is extended to 12 and `inputActive` is set (P1–P2 active, P3–P12 inactive). If IPC mode is also on, `externalInputChan` is created and `startInputListener(inputSocketPath)` is started in a goroutine.
- **`stepCommandLists`**: Drains `externalInputChan` (non-blocking), parses messages via `parseInputMessage`, and only steps command lists for slots where `inputActive[i]` is true.
- **`TogglePlayerInput(index, enable)`**: Toggles input for slot `index` (0–11). Only indices 2–11 (P3–P12) can be toggled; P1–P2 stay active. When disabling in IPC mode, that slot’s button/axis state is cleared.
- **`applyMotifInputCap(maxSlot)`**: Disables slots `>= maxSlot` (1-based) when extended inputs are on, so motifs that only define up to P8 don’t poll P9–P12.
- **IPC server**: `startInputListener` listens on the given Unix socket, removes stale socket file, and hands each client to `handleInputConn` in a goroutine.
- **IPC protocol**: Binary. Magic `"IK"` (2 bytes), `numPlayers` (1 byte), then per player: `idx` (1), `buttons` (2 bytes, 14 bits used), `axes` (6 × float32). `parseInputMessage` updates `lastButtonStates` and `lastAxes` under `externalInputMu`.

### Motif (`src/motif.go`)

- **`parseSelectInfoMaxPlayer(iniFile)`**: Scans `[Select Info]` for keys like `p1.face`, `p2.cursor`, etc., and returns the highest player number (1–12), or 0 if none.
- After loading motif music, if extended inputs are enabled, `parseSelectInfoMaxPlayer` is called and `sys.applyMotifInputCap(maxSlot)` is applied so the number of polled slots matches the motif.

### Script / Lua (`src/script.go`)

- **`setPlayers(total)`**: When extended inputs are enabled, the maximum allowed players is 12; otherwise it remains `MaxSimul*2`. `total` is clamped to that max.
- **`togglePlayerInput(index, enable)`**: New Lua-callable function. `index` is 1-based (P1–P12). Calls `sys.TogglePlayerInput(index-1, enable)`; only indices 1–12 are accepted.

---

## 3. Summary table

| Area           | Change |
|----------------|--------|
| Build (macOS)  | nasm optional; required only for local FFmpeg build; `GOWORK=off` for isolated Go build |
| CLI            | `--enable-extended-inputs`, `--input-ipc`, `--input-socket-path` |
| Config         | Players cap set to 12 when extended inputs enabled |
| Input          | IPC path feeds 12 slots (buttons + 6 axes) from Unix socket state |
| System         | 12 command lists, `inputActive`, IPC listener, binary protocol, toggle/cap helpers |
| Motif          | Derive max player from Select Info and cap input slots |
| Script         | `setPlayers` respects 12 max; new `togglePlayerInput` |

These changes add an optional “extended inputs” mode (up to 12 players with optional IPC) and make the macOS build script more flexible and workspace-safe.
