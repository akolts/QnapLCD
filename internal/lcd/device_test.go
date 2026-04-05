package lcd

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

// MockPort simulates a serial port for testing.
type MockPort struct {
	ReadBuf  *bytes.Buffer // data "received" from the LCD controller
	WriteBuf *bytes.Buffer // data "sent" to the LCD controller
	closed   bool
}

func newMockPort() *MockPort {
	return &MockPort{
		ReadBuf:  &bytes.Buffer{},
		WriteBuf: &bytes.Buffer{},
	}
}

func (m *MockPort) Read(buf []byte) (int, error) {
	return m.ReadBuf.Read(buf)
}

func (m *MockPort) Write(buf []byte) (int, error) {
	return m.WriteBuf.Write(buf)
}

func (m *MockPort) Close() error {
	m.closed = true
	return nil
}

// --- Device operation tests ---

func TestWriteLine(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.WriteLine(1, "Hello"); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x0C, 0x00, 0x05, 'H', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestWriteLineLine2(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.WriteLine(2, "World"); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x0C, 0x01, 0x05, 'W', 'o', 'r', 'l', 'd'}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestWriteLines(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.WriteLines("Line 1", "Line 2"); err != nil {
		t.Fatal(err)
	}

	got := mock.WriteBuf.Bytes()
	line1cmd := []byte{0x4D, 0x0C, 0x00, 0x06, 'L', 'i', 'n', 'e', ' ', '1'}
	line2cmd := []byte{0x4D, 0x0C, 0x01, 0x06, 'L', 'i', 'n', 'e', ' ', '2'}
	expected := append(line1cmd, line2cmd...)

	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestWriteLineOutOfRange(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.WriteLine(0, "test"); err == nil {
		t.Error("expected error for line 0")
	}
	if err := dev.WriteLine(3, "test"); err == nil {
		t.Error("expected error for line 3")
	}
	if err := dev.WriteLine(-1, "test"); err == nil {
		t.Error("expected error for line -1")
	}
}

func TestClear(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.Clear(); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x0D}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestBacklight(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.Backlight(true); err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x4D, 0x5E, 0x01}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("backlight on: got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}

	mock.WriteBuf.Reset()

	if err := dev.Backlight(false); err != nil {
		t.Fatal(err)
	}
	expected = []byte{0x4D, 0x5E, 0x00}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("backlight off: got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestReset(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.Reset(); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0xFF}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestGetBoardID(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.GetBoardID(); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x00}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestGetProtocol(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.GetProtocol(); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x07}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestGetButtons(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.GetButtons(); err != nil {
		t.Fatal(err)
	}

	expected := []byte{0x4D, 0x06}
	if !bytes.Equal(mock.WriteBuf.Bytes(), expected) {
		t.Errorf("got %v, want %v", mock.WriteBuf.Bytes(), expected)
	}
}

func TestClose(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}
	if !mock.closed {
		t.Error("port not closed")
	}
}

// --- Event reader tests ---

func TestReadEventButtonUp(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0x05, 0x00, 0x01})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventButtonPress {
		t.Errorf("type: got %v, want %v", event.Type, EventButtonPress)
	}
	if event.Data != ButtonUp {
		t.Errorf("data: got 0x%04X, want 0x%04X", event.Data, ButtonUp)
	}
}

func TestReadEventButtonDown(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0x05, 0x00, 0x02})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventButtonPress {
		t.Errorf("type: got %v, want %v", event.Type, EventButtonPress)
	}
	if event.Data != ButtonDown {
		t.Errorf("data: got 0x%04X, want 0x%04X", event.Data, ButtonDown)
	}
}

func TestReadEventBoardID(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0x01, 0x12, 0x34})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventBoardID {
		t.Errorf("type: got %v, want %v", event.Type, EventBoardID)
	}
	if event.Data != 0x1234 {
		t.Errorf("data: got 0x%04X, want 0x1234", event.Data)
	}
}

func TestReadEventProtocolVersion(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0x08, 0x00, 0x01})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventProtocolVersion {
		t.Errorf("type: got %v, want %v", event.Type, EventProtocolVersion)
	}
	if event.Data != 0x0001 {
		t.Errorf("data: got 0x%04X, want 0x0001", event.Data)
	}
}

func TestReadEventResetOK(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0xAA})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventResetOK {
		t.Errorf("type: got %v, want %v", event.Type, EventResetOK)
	}
}

func TestReadEventACK(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0xFA})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventAck {
		t.Errorf("type: got %v, want %v", event.Type, EventAck)
	}
}

func TestReadEventNACK(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0xFB, 0x0C})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventNack {
		t.Errorf("type: got %v, want %v", event.Type, EventNack)
	}
	if event.Data != 0x0C {
		t.Errorf("data: got 0x%04X, want 0x000C", event.Data)
	}
}

func TestReadEventAltPreamble(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	// Alternate preamble 0x83 should also be accepted.
	mock.ReadBuf.Write([]byte{0x83, 0xFA})

	event, err := dev.readEvent()
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventAck {
		t.Errorf("type: got %v, want %v", event.Type, EventAck)
	}
}

func TestReadEventBadPreamble(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0xFF, 0xFA})

	_, err := dev.readEvent()
	if err == nil {
		t.Error("expected error for bad preamble")
	}
}

func TestReadEventUnknownCommand(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	mock.ReadBuf.Write([]byte{0x53, 0x99})

	_, err := dev.readEvent()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestReadEventEOF(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	// Empty buffer → EOF on first read.
	_, err := dev.readEvent()
	if err == nil {
		t.Error("expected error on EOF")
	}
}

func TestReadEventPartialResponse(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	// Preamble + command but missing data bytes for board ID.
	mock.ReadBuf.Write([]byte{0x53, 0x01, 0x12})

	_, err := dev.readEvent()
	if err == nil {
		t.Error("expected error for partial response")
	}
}

func TestStartReaderMultipleEvents(t *testing.T) {
	mock := newMockPort()
	dev := NewDevice(mock)

	// Feed three events: Button Up, ACK, Button Down, then EOF.
	mock.ReadBuf.Write([]byte{
		0x53, 0x05, 0x00, 0x01, // Button Up
		0x53, 0xFA, // ACK
		0x53, 0x05, 0x00, 0x02, // Button Down
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := dev.StartReader(ctx)

	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[0].Type != EventButtonPress || events[0].Data != ButtonUp {
		t.Errorf("event 0: got %v, want ButtonPress Up", events[0])
	}
	if events[1].Type != EventAck {
		t.Errorf("event 1: got %v, want Ack", events[1])
	}
	if events[2].Type != EventButtonPress || events[2].Data != ButtonDown {
		t.Errorf("event 2: got %v, want ButtonPress Down", events[2])
	}
}

// blockingPort blocks on Read until Close is called. Used to test
// context cancellation of the reader goroutine.
type blockingPort struct {
	done chan struct{}
}

func newBlockingPort() *blockingPort {
	return &blockingPort{done: make(chan struct{})}
}

func (b *blockingPort) Read(buf []byte) (int, error) {
	<-b.done
	return 0, io.EOF
}

func (b *blockingPort) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func (b *blockingPort) Close() error {
	select {
	case <-b.done:
		// Already closed.
	default:
		close(b.done)
	}
	return nil
}

func TestStartReaderContextCancel(t *testing.T) {
	bp := newBlockingPort()
	dev := NewDevice(bp)

	ctx, cancel := context.WithCancel(context.Background())
	ch := dev.StartReader(ctx)

	// Cancel context and close port to unblock the reader.
	cancel()
	bp.Close()

	// The channel should close promptly.
	select {
	case _, ok := <-ch:
		if ok {
			for range ch {
			}
		}
		// Channel closed — good.
	case <-time.After(2 * time.Second):
		t.Error("reader did not exit after context cancel")
	}
}
