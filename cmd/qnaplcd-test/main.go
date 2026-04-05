// Command qnaplcd-test is a standalone hardware test tool for the QNAP LCD
// panel. It validates serial communication with the A125 display controller
// by reading button presses, writing to the LCD, or running a full test
// sequence.
//
// Usage:
//
//	qnaplcd-test -buttons           # listen for button presses
//	qnaplcd-test -lcd               # write test text to display
//	qnaplcd-test -all               # run full hardware test
//	qnaplcd-test -port /dev/ttyS0   # use alternate serial port
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QnapLCD/internal/lcd"

	"go.bug.st/serial"
)

func main() {
	port := flag.String("port", "/dev/ttyS1", "serial port device")
	baud := flag.Int("baud", 1200, "baud rate")
	buttons := flag.Bool("buttons", false, "listen for button presses (runs until Ctrl+C)")
	lcdTest := flag.Bool("lcd", false, "test LCD write (writes test text to both lines)")
	all := flag.Bool("all", false, "run full hardware test sequence")
	flag.Parse()

	if !*buttons && !*lcdTest && !*all {
		fmt.Fprintln(os.Stderr, "specify one of: -buttons, -lcd, -all")
		flag.Usage()
		os.Exit(1)
	}

	sp, err := serial.Open(*port, &serial.Mode{
		BaudRate: *baud,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		log.Fatalf("failed to open serial port %s: %v", *port, err)
	}
	defer sp.Close()

	dev := lcd.NewDevice(sp)

	switch {
	case *buttons:
		runButtonTest(dev, *port, *baud)
	case *lcdTest:
		runLCDTest(dev, *port, *baud)
	case *all:
		runFullTest(dev, *port, *baud)
	}
}

// runButtonTest listens for button presses and prints them to stdout.
func runButtonTest(dev *lcd.Device, port string, baud int) {
	fmt.Printf("Listening for button presses on %s at %d baud...\n", port, baud)
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println()

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ch := dev.StartReader(ctx)

	// Close the port when context is cancelled to unblock the reader goroutine
	// (which is stuck in a blocking serial read).
	go func() {
		<-ctx.Done()
		dev.Close()
	}()

	for event := range ch {
		ts := time.Now().Format("2006-01-02 15:04:05")
		switch event.Type {
		case lcd.EventButtonPress:
			// Skip "released" events (0x0000) — each press sends
			// the button code followed by 0x0000 on release.
			if event.Data == lcd.ButtonNone {
				continue
			}
			fmt.Printf("[%s] Button: %-5s (0x%04X)\n", ts, lcd.ButtonName(event.Data), event.Data)
		default:
			fmt.Printf("[%s] Event:  %-18s (0x%04X)\n", ts, event.Type, event.Data)
		}
	}

	fmt.Println("\nReceived interrupt, closing port.")
}

// runLCDTest writes test text to the display for visual verification.
func runLCDTest(dev *lcd.Device, port string, baud int) {
	fmt.Printf("Testing LCD on %s at %d baud...\n", port, baud)

	steps := []struct {
		name string
		fn   func() error
		info string
	}{
		{"Reset", dev.Reset, ""},
		{"Clear display", dev.Clear, ""},
		{"Backlight on", func() error { return dev.Backlight(true) }, ""},
		{"Write line 1", func() error { return dev.WriteLine(1, "LCD Test Line 1") }, `"LCD Test Line 1"`},
		{"Write line 2", func() error { return dev.WriteLine(2, "QNAP HW Test") }, `"QNAP HW Test"`},
	}

	for i, step := range steps {
		fmt.Printf("[%d/%d] %-20s", i+1, len(steps), step.name+"...")
		if err := step.fn(); err != nil {
			fmt.Printf("FAIL (%v)\n", err)
			log.Fatalf("test failed at step %d", i+1)
		}
		if step.info != "" {
			fmt.Printf("OK (%s)\n", step.info)
		} else {
			fmt.Println("OK")
		}
		// Small delay between commands to let the controller process.
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("\nDone. Check LCD visually.")
}

// runFullTest runs a comprehensive hardware test with response verification.
func runFullTest(dev *lcd.Device, port string, baud int) {
	fmt.Printf("Running full hardware test on %s at %d baud...\n\n", port, baud)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ch := dev.StartReader(ctx)

	// Close the port when context expires to unblock the reader.
	go func() {
		<-ctx.Done()
		dev.Close()
	}()

	type testStep struct {
		name     string
		fn       func() error
		expectEv lcd.EventType
	}

	steps := []testStep{
		{"Reset", dev.Reset, lcd.EventResetOK},
		{"Get Board ID", dev.GetBoardID, lcd.EventBoardID},
		{"Get Protocol", dev.GetProtocol, lcd.EventProtocolVersion},
		{"Clear display", dev.Clear, lcd.EventAck},
		{"Backlight on", func() error { return dev.Backlight(true) }, lcd.EventAck},
		{"Write line 1", func() error { return dev.WriteLine(1, "Full Test L1") }, lcd.EventAck},
		{"Write line 2", func() error { return dev.WriteLine(2, "Full Test L2") }, lcd.EventAck},
	}

	passed := 0
	for i, step := range steps {
		fmt.Printf("[%d/%d] %-20s", i+1, len(steps), step.name+"...")

		if err := step.fn(); err != nil {
			fmt.Printf("FAIL (send: %v)\n", err)
			log.Fatalf("test failed at step %d", i+1)
		}

		// Wait for the expected response event.
		event, err := waitForEvent(ch, step.expectEv, 5*time.Second)
		if err != nil {
			fmt.Printf("FAIL (%v)\n", err)
			log.Fatalf("test failed at step %d", i+1)
		}

		switch event.Type {
		case lcd.EventBoardID:
			fmt.Printf("OK (ID: 0x%04X)\n", event.Data)
		case lcd.EventProtocolVersion:
			fmt.Printf("OK (Version: 0x%04X)\n", event.Data)
		default:
			fmt.Printf("OK (%s)\n", event.Type)
		}

		passed++
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("\nAll %d tests passed.\n", passed)
}

// waitForEvent reads events from the channel until the expected event type
// arrives or the timeout expires. Unexpected events are logged and skipped.
func waitForEvent(ch <-chan lcd.Event, expected lcd.EventType, timeout time.Duration) (lcd.Event, error) {
	deadline := time.After(timeout)
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return lcd.Event{}, fmt.Errorf("event channel closed")
			}
			if event.Type == expected {
				return event, nil
			}
			// Skip unexpected events (e.g., extra ACKs).
			fmt.Printf("  (skipped: %s 0x%04X) ", event.Type, event.Data)
		case <-deadline:
			return lcd.Event{}, fmt.Errorf("timeout waiting for %s", expected)
		}
	}
}
