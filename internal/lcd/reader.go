package lcd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

// StartReader launches a goroutine that reads events from the LCD controller
// and sends them on the returned channel. The goroutine exits when the context
// is cancelled or when a read error occurs (e.g., the port is closed).
// The channel is closed when the goroutine exits.
func (d *Device) StartReader(ctx context.Context) <-chan Event {
	ch := make(chan Event, 16)
	go func() {
		defer close(ch)
		for {
			event, err := d.readEvent()
			if err != nil {
				// If context was cancelled, exit silently.
				select {
				case <-ctx.Done():
					return
				default:
				}
				slog.Error("lcd: failed to read event", "error", err)
				return
			}
			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// readEvent reads a single event from the serial port. It blocks until
// a complete response is received.
func (d *Device) readEvent() (Event, error) {
	buf := make([]byte, 1)

	// Read preamble byte.
	if _, err := io.ReadFull(d.port, buf); err != nil {
		return Event{}, fmt.Errorf("read preamble: %w", err)
	}

	if buf[0] != respPreamble && buf[0] != respPreambleAlt {
		return Event{}, fmt.Errorf("unexpected preamble: 0x%02X", buf[0])
	}

	// Read response command byte.
	if _, err := io.ReadFull(d.port, buf); err != nil {
		return Event{}, fmt.Errorf("read command: %w", err)
	}

	switch buf[0] {
	case respBoardID:
		data, err := readUint16(d.port)
		if err != nil {
			return Event{}, err
		}
		return Event{Type: EventBoardID, Data: data}, nil

	case respButtonStatus:
		data, err := readUint16(d.port)
		if err != nil {
			return Event{}, err
		}
		return Event{Type: EventButtonPress, Data: data}, nil

	case respProtocolVer:
		data, err := readUint16(d.port)
		if err != nil {
			return Event{}, err
		}
		return Event{Type: EventProtocolVersion, Data: data}, nil

	case respResetOK:
		return Event{Type: EventResetOK}, nil

	case respACK:
		return Event{Type: EventAck}, nil

	case respNACK:
		if _, err := io.ReadFull(d.port, buf); err != nil {
			return Event{}, fmt.Errorf("read nack data: %w", err)
		}
		return Event{Type: EventNack, Data: uint16(buf[0])}, nil

	default:
		return Event{}, fmt.Errorf("unknown response command: 0x%02X", buf[0])
	}
}

// readUint16 reads two bytes from the reader and returns them as a big-endian uint16.
func readUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, fmt.Errorf("read uint16: %w", err)
	}
	return uint16(buf[0])<<8 | uint16(buf[1]), nil
}
