# QnapLCD

A lightweight Go application for controlling the front-panel LCD display and
buttons on QNAP NAS hardware running TrueNAS SCALE (or other Linux distributions).

Displays system information (hostname, uptime, network addresses, ZFS pool status)
on the QNAP A125 2-line LCD with Up/Down button navigation.

Based on the Python [QnapLCD-Menu](https://github.com/shouser/QnapLCD-Menu)
project by Stephen Houser.

## Compatibility

| Component     | Tested With                                    |
|---------------|------------------------------------------------|
| Hardware      | QNAP TVS-471 (should work on other A125 models) |
| LCD           | A125 — 2 lines, 16 columns, 2 buttons         |
| Serial Port   | `/dev/ttyS1` at 1200 baud                     |
| OS            | TrueNAS SCALE 25.10.x (Debian 12)             |
| Architecture  | x86_64 (amd64)                                |

## Building

Requires Docker. The build produces two static Linux binaries:

```bash
./build.sh
```

This outputs:
- `qnaplcd` — main application
- `qnaplcd-test` — hardware test tool

To build without Docker (requires Go 1.25+):

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o qnaplcd ./cmd/qnaplcd
CGO_ENABLED=0 go build -ldflags="-s -w" -o qnaplcd-test ./cmd/qnaplcd-test
```

## Installation on TrueNAS

### Important: Filesystem Requirements

TrueNAS SCALE enforces `noexec` on the boot pool (home directories, `/root`,
etc.). Binaries placed there will be killed on execution with:

```
sudo: process NNNNN unexpected status 0x57f
zsh: killed     ./binary
```

**Binaries must be placed on a data pool dataset with `exec=on`.**

### Steps

1. Copy binaries to TrueNAS:
   ```bash
   scp qnaplcd qnaplcd-test truenas_admin@truenas:/tmp/
   ```

2. On TrueNAS (as root), move to a data pool path:
   ```bash
   mkdir -p /mnt/POOLNAME/scripts
   cp /tmp/qnaplcd /tmp/qnaplcd-test /mnt/POOLNAME/scripts/
   chmod 755 /mnt/POOLNAME/scripts/qnaplcd /mnt/POOLNAME/scripts/qnaplcd-test
   ```
   Replace `POOLNAME` with your ZFS pool name (e.g., `tank`, `data`, `media`).

3. Verify:
   ```bash
   /mnt/POOLNAME/scripts/qnaplcd -version
   ```

4. If the dataset doesn't allow execution, enable it:
   ```bash
   zfs get exec POOLNAME/scripts       # check current setting
   zfs set exec=on POOLNAME/scripts    # enable if needed
   ```

## Usage

### Display a message and exit

Used for TrueNAS Init/Shutdown scripts:

```bash
# Single line
qnaplcd -show "System Ready"

# Two lines (separated by |)
qnaplcd -show "truenas|Initializing..."
```

### Run the menu daemon

Runs in the background, displays system info, responds to button presses:

```bash
tmux new-session -d -s qnaplcd /mnt/POOLNAME/scripts/qnaplcd
```

Using `tmux` ensures the daemon survives the caller's lifecycle (important for
TrueNAS Init scripts — a bare `&` may be killed when the init context is
cleaned up after boot). You can later attach to the session for debugging:

```bash
tmux attach -t qnaplcd
```

### All flags

```
  -port string    Serial port device (default "/dev/ttyS1")
  -baud int       Baud rate (default 1200)
  -timeout int    Backlight timeout in seconds (default 30)
  -refresh int    Menu refresh interval in seconds (default 30)
  -show string    Display message and exit (format: "line1" or "line1|line2")
  -v              Enable debug logging
  -version        Print version and exit
```

## TrueNAS Init/Shutdown Scripts

Configure under **System Settings > Advanced > Init/Shutdown Scripts**:

| Script Type | Command |
|-------------|---------|
| Pre-Init    | `/mnt/POOLNAME/scripts/qnaplcd -show "$(hostname -s)\|Initializing..."` |
| Post-Init   | `tmux new-session -d -s qnaplcd /mnt/POOLNAME/scripts/qnaplcd` |
| Shutdown    | `/mnt/POOLNAME/scripts/qnaplcd -show "$(hostname -s)\|Shutting Down..."`; `tmux kill-session -t qnaplcd 2>/dev/null` |

**Why `tmux` instead of `&`?** TrueNAS post-init commands are meant for
one-shot tasks. A long-running process started with `&` is still tied to the
init task's environment and may be killed when TrueNAS cleans up that context.
`tmux` creates a fully detached session that survives boot completion and shell
logouts. `tmux` is pre-installed on TrueNAS SCALE.

## Menu Items

When running in daemon mode, the LCD shows these screens (navigate with Up/Down buttons):

| Screen          | Line 1              | Line 2                  |
|-----------------|----------------------|-------------------------|
| TrueNAS Version | Version prefix       | Build identifier        |
| Hostname        | Hostname             | OS (architecture)       |
| Uptime          | Uptime duration      | Load average            |
| Network (per interface) | Interface name | IP address             |
| ZFS Pool (per pool)     | Pool (health)  | Used of Total          |

- Backlight turns off after 30 seconds of inactivity (configurable with `-timeout`)
- Menu content refreshes every 30 seconds (configurable with `-refresh`)
- If `zpool` or TrueNAS `cli` are not available, those items are simply skipped

## Hardware Test Tool

A standalone binary for verifying LCD and button hardware on the NAS:

```bash
# Listen for button presses (Ctrl+C to stop)
qnaplcd-test -buttons

# Write test text to the display
qnaplcd-test -lcd

# Full hardware test with response verification
qnaplcd-test -all
```

Use `-port` and `-baud` flags if your serial port differs from the defaults.

## Running Tests

```bash
go test ./...
```

## Project Structure

```
cmd/
  qnaplcd/          Main application
  qnaplcd-test/     Hardware test tool
internal/
  lcd/              LCD protocol, device driver, event reader
  menu/             Menu system (planned)
  sysinfo/          System information gathering (planned)
```

## License

See [LICENSE](LICENSE) file.
