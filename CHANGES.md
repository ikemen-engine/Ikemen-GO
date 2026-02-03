# Changelog — Extended Inputs, Build Fixes & Battle Event Channel

Summary of uncommitted changes on `develop` (includes extended inputs, build script fixes, and battle event channel for CLI consumption).

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

---

## 4. Battle event channel (CLI-consumable events)

A global buffered channel emits JSON battle events so external tools (e.g. CLI, scripts) can consume them by reading stdout lines prefixed with `[BATTLE_EVENT]`.

### Channel and consumer

- **`src/common.go`**
  - **`BattleEventChan`**: New global `chan string` with buffer size 100. Event hooks marshal JSON and send the string here; non-blocking sends are used so the game never blocks if the consumer is slow.

- **`src/main.go`**
  - In **`realMain()`**, after `processCommandLine()` and cmdFlags/stats setup: a goroutine runs `for msg := range BattleEventChan { fmt.Printf("[BATTLE_EVENT] %s\n", msg) }`. All events are printed as single lines for easy parsing (e.g. `grep '\[BATTLE_EVENT\]'`).

### Event types and hooks

- **`src/lifebar.go`**
  - **`FinishType.String()`**: Returns `"FT_NotYet"`, `"FT_KO"`, `"FT_DKO"`, `"FT_TO"`, `"FT_TODraw"` (and a fallback for unknown values).
  - **`WinType.String()`**: Returns readable names for all WinType constants and their P/C variants (e.g. `"WT_Normal"`, `"WT_PPerfect"`, `"WT_CTime"`).

- **`src/system.go`**
  - **New `System` field**: `fightEndEventEmitted bool`, reset when a round starts (with `finishType` / `winTeam`).
  - **Post-round block**: When `s.roundEnded() || s.roundEndDecision()` and `!s.fightEndEventEmitted && s.finishType != FT_NotYet`, emits one **`fight_end`** event with `type`, `win_team`, `win_type` (winner’s WinType, or team 0 on draw), `finish_type`. Then sets `fightEndEventEmitted = true` so it emits once per round.

- **`src/char.go`**
  - **Hit with HP deduction**: In the block that applies `c.ghv.damage` (around `lifeAdd`), after applying damage: if `damage > 0`, emits **`hit_hp_update`** with `attacker_id` (from `c.ghv.playerid`), `target_id` (`c.id`), `damage`, `hp_after`.
  - **Power (MP) updates**: In **`setPower()`**, after updating `c.power`, if `Abs(c.power - oldPower) > 100`, emits **`mp_update`** with `char_id`, `change`, `power_after`.
  - **Trigger fire**: At the end of **`hitResultCheck()`**, when `hitResult != 0`, emits **`trigger_fire`** with `trigger_id` `"hit"` or `"guard"` and `context` `"attacker:%d target:%d"` (attacker and getter IDs). One event per hit/guard.

- **`src/stage.go`**
  - **New `Stage` field**: `lastEventEmit time.Time` for throttling.
  - **`action()`**: When `canStep && time.Since(s.lastEventEmit) > time.Second`, emits **`stage_update`** with `stage_time`, `shake` (from `sys.envShake.time`), `music_pos` (0), `bounds` (left/right/top/bot as ints). Then sets `s.lastEventEmit = time.Now()` so stage events are limited to about once per second.

### Summary table (battle events)

| File        | Change |
|------------|--------|
| `common.go` | `BattleEventChan` (buffered channel) |
| `main.go`   | Goroutine that prints `[BATTLE_EVENT] <json>` to stdout |
| `lifebar.go`| `FinishType.String()`, `WinType.String()` for event payloads |
| `system.go` | `fightEndEventEmitted`; one `fight_end` per round when fight ends |
| `char.go`   | `hit_hp_update` on damage; `mp_update` on power change > 100; `trigger_fire` on hit/guard |
| `stage.go`  | `lastEventEmit`; `stage_update` throttled to ~1/s when `canStep` |

All event sends use a non-blocking `select { case BattleEventChan <- ...: default: }` so a full channel does not block the game loop.
