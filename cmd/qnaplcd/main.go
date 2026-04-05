// Command qnaplcd controls the QNAP front-panel LCD display and buttons.
//
// In daemon mode (default), it runs an interactive menu showing system
// information. In show mode (-show flag), it displays a message and exits.
//
// Usage:
//
//	qnaplcd                              # run menu daemon
//	qnaplcd -show "hostname|Ready..."    # display message and exit
//	qnaplcd -port /dev/ttyS0             # use alternate serial port
//	qnaplcd -v                           # enable debug logging
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"context"

	"QnapLCD/internal/lcd"
	"QnapLCD/internal/menu"
	"QnapLCD/internal/sysinfo"

	"go.bug.st/serial"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	port := flag.String("port", "/dev/ttyS1", "serial port device")
	baud := flag.Int("baud", 1200, "baud rate")
	show := flag.String("show", "", `display a message and exit (format: "line1" or "line1|line2")`)
	timeout := flag.Int("timeout", 30, "backlight timeout in seconds")
	refresh := flag.Int("refresh", 30, "menu refresh interval in seconds")
	verbose := flag.Bool("v", false, "enable debug logging")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("qnaplcd %s\n", version)
		return
	}

	// Configure logging.
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	// Open serial port.
	sp, err := serial.Open(*port, &serial.Mode{
		BaudRate: *baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		slog.Error("failed to open serial port", "port", *port, "error", err)
		os.Exit(1)
	}
	defer sp.Close()

	dev := lcd.NewDevice(sp)

	if *show != "" {
		runShow(dev, *show)
		return
	}

	runMenu(dev, *timeout, *refresh)
}

// runShow displays a message on the LCD and exits.
func runShow(dev *lcd.Device, msg string) {
	parts := strings.SplitN(msg, "|", 2)
	line1 := parts[0]
	line2 := ""
	if len(parts) == 2 {
		line2 = parts[1]
	}

	if err := dev.Backlight(true); err != nil {
		slog.Error("backlight failed", "error", err)
	}
	if err := dev.Clear(); err != nil {
		slog.Error("clear failed", "error", err)
	}
	if err := dev.WriteLines(line1, line2); err != nil {
		slog.Error("write failed", "error", err)
	}
}

// runMenu runs the interactive menu system.
func runMenu(dev *lcd.Device, backlightTimeout, refreshInterval int) {
	slog.Info("starting menu daemon",
		"timeout", backlightTimeout,
		"refresh", refreshInterval)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize display.
	_ = dev.Backlight(true)
	_ = dev.Reset()
	_ = dev.Clear()

	// Cache static system info (doesn't change at runtime).
	static := cacheStaticInfo()

	// Build initial menu and display.
	m := menu.New(buildMenuItems(static))
	displayCurrent(dev, m)

	// Start event reader.
	ch := dev.StartReader(ctx)

	// Timers.
	refreshTicker := time.NewTicker(time.Duration(refreshInterval) * time.Second)
	defer refreshTicker.Stop()

	backlightTimer := time.NewTimer(time.Duration(backlightTimeout) * time.Second)
	defer backlightTimer.Stop()
	backlightOn := true

	slog.Info("menu daemon running", "items", m.Len())

loop:
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				slog.Info("event channel closed")
				break loop
			}

			// Only handle actual button presses (ignore releases and non-button events).
			if event.Type != lcd.EventButtonPress || event.Data == lcd.ButtonNone {
				slog.Debug("event", "type", event.Type, "data", event.Data)
				continue
			}

			slog.Debug("button press", "button", lcd.ButtonName(event.Data))

			// Turn on backlight on any press.
			if !backlightOn {
				_ = dev.Backlight(true)
				backlightOn = true
			}
			backlightTimer.Reset(time.Duration(backlightTimeout) * time.Second)

			switch event.Data {
			case lcd.ButtonUp:
				m.Prev()
			case lcd.ButtonDown:
				m.Next()
			}

			displayCurrent(dev, m)

		case <-refreshTicker.C:
			slog.Debug("refreshing menu")
			m.SetItems(buildMenuItems(static))
			displayCurrent(dev, m)

		case <-backlightTimer.C:
			slog.Debug("backlight timeout")
			_ = dev.Backlight(false)
			backlightOn = false

		case <-ctx.Done():
			break loop
		}
	}

	// Shutdown: backlight off, close port (unblocks reader if still running).
	slog.Info("shutting down")
	_ = dev.Backlight(false)
	_ = dev.Close()

	// Drain reader channel to ensure goroutine exits cleanly.
	for range ch {
	}
}

// staticInfo holds system information that doesn't change at runtime.
// Computed once at startup to avoid repeated exec/syscalls.
type staticInfo struct {
	hostname string
	osInfo   string
	truenas  struct {
		line1, line2 string
		available    bool
	}
}

func cacheStaticInfo() staticInfo {
	var s staticInfo
	s.hostname = sysinfo.Hostname()
	s.osInfo = sysinfo.OSInfo()
	if l1, l2, err := sysinfo.TrueNASVersion(); err == nil {
		s.truenas.line1 = l1
		s.truenas.line2 = l2
		s.truenas.available = true
	}
	slog.Debug("cached static info",
		"hostname", s.hostname,
		"os", s.osInfo,
		"truenas", s.truenas.available)
	return s
}

// buildMenuItems creates the full list of menu items from current system state.
// Static items use cached values; dynamic items are rediscovered each call.
func buildMenuItems(s staticInfo) []menu.Item {
	var items []menu.Item

	// TrueNAS version (cached, optional).
	if s.truenas.available {
		items = append(items, func() (string, string) {
			return s.truenas.line1, s.truenas.line2
		})
	}

	// Hostname + OS (cached).
	items = append(items, func() (string, string) {
		return s.hostname, s.osInfo
	})

	// Uptime + load average (fetched live on each display).
	items = append(items, func() (string, string) {
		up, err := sysinfo.Uptime()
		if err != nil {
			up = "N/A"
		}
		load, err := sysinfo.LoadAvg()
		if err != nil {
			load = "N/A"
		}
		return "Up: " + up, load
	})

	// Network interfaces (rediscovered each refresh cycle).
	for _, iface := range sysinfo.NetworkInterfaces() {
		iface := iface // capture for closure
		items = append(items, func() (string, string) {
			return iface.Name, iface.Addr
		})
	}

	// ZFS pools (rediscovered each refresh cycle).
	if pools, err := sysinfo.ZFSPools(); err == nil {
		for _, pool := range pools {
			pool := pool // capture for closure
			items = append(items, func() (string, string) {
				return pool.Line1(), pool.Line2()
			})
		}
	}

	return items
}

// displayCurrent writes the current menu item to the LCD.
func displayCurrent(dev *lcd.Device, m *menu.Menu) {
	l1, l2 := m.Current()
	if err := dev.Clear(); err != nil {
		slog.Error("clear failed", "error", err)
	}
	if err := dev.WriteLines(l1, l2); err != nil {
		slog.Error("write failed", "error", err)
	}
}
