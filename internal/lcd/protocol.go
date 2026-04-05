// Package lcd provides communication with the QNAP A125 front-panel LCD
// display and buttons over a serial connection.
package lcd

// Serial protocol constants for the QNAP A125 LCD controller.
//
// Commands are sent with preamble 0x4D followed by a command byte and
// optional parameters. Responses arrive with preamble 0x53 (or 0x83)
// followed by a response code and optional data.
const (
	// Command preamble byte — all commands start with this.
	cmdPreamble byte = 0x4D

	// Command bytes (sent after preamble).
	cmdGetBoardID   byte = 0x00
	cmdGetButtons   byte = 0x06
	cmdGetProtocol  byte = 0x07
	cmdDisplayText  byte = 0x0C
	cmdDisplayClear byte = 0x0D
	cmdBacklight    byte = 0x5E
	cmdReset        byte = 0xFF

	// Response preamble bytes — responses start with one of these.
	respPreamble    byte = 0x53
	respPreambleAlt byte = 0x83

	// Response command bytes (received after preamble).
	respBoardID      byte = 0x01
	respButtonStatus byte = 0x05
	respProtocolVer  byte = 0x08
	respResetOK      byte = 0xAA
	respACK          byte = 0xFA
	respNACK         byte = 0xFB

	// Backlight parameter values.
	backlightOn  byte = 0x01
	backlightOff byte = 0x00

	// Display dimensions for the A125 LCD.
	MaxColumns = 16
	MaxLines   = 2
)

// EventType identifies the kind of event received from the LCD controller.
type EventType int

const (
	EventBoardID         EventType = iota // Board ID response
	EventButtonPress                      // Button press notification
	EventProtocolVersion                  // Protocol version response
	EventResetOK                          // Reset acknowledgment
	EventAck                              // Command acknowledgment
	EventNack                             // Negative acknowledgment
)

// String returns a human-readable name for the event type.
func (e EventType) String() string {
	switch e {
	case EventBoardID:
		return "BoardID"
	case EventButtonPress:
		return "ButtonPress"
	case EventProtocolVersion:
		return "ProtocolVersion"
	case EventResetOK:
		return "ResetOK"
	case EventAck:
		return "Ack"
	case EventNack:
		return "Nack"
	default:
		return "Unknown"
	}
}

// Event represents a message received from the LCD controller.
type Event struct {
	Type EventType
	Data uint16
}

// Button codes as reported in EventButtonPress events.
// Codes are a bitmask: Up=0x01, Down=0x02, Both=0x03.
// Each physical press sends two events: the button code, then ButtonNone
// when the button is released.
const (
	ButtonNone uint16 = 0x0000
	ButtonUp   uint16 = 0x0001
	ButtonDown uint16 = 0x0002
	ButtonBoth uint16 = 0x0003 // Up + Down pressed simultaneously
)

// ButtonName returns a human-readable name for a button code.
func ButtonName(code uint16) string {
	switch code {
	case ButtonNone:
		return "Released"
	case ButtonUp:
		return "Up"
	case ButtonDown:
		return "Down"
	case ButtonBoth:
		return "Both"
	default:
		return "Unknown"
	}
}

// Command builders — each returns the byte slice to send over serial.

func buildGetBoardID() []byte  { return []byte{cmdPreamble, cmdGetBoardID} }
func buildGetButtons() []byte  { return []byte{cmdPreamble, cmdGetButtons} }
func buildGetProtocol() []byte { return []byte{cmdPreamble, cmdGetProtocol} }
func buildClear() []byte       { return []byte{cmdPreamble, cmdDisplayClear} }
func buildReset() []byte       { return []byte{cmdPreamble, cmdReset} }

func buildBacklight(on bool) []byte {
	if on {
		return []byte{cmdPreamble, cmdBacklight, backlightOn}
	}
	return []byte{cmdPreamble, cmdBacklight, backlightOff}
}

// buildDisplayText builds the command to write text to a display line.
// Line 1 is the top line, line 2 is the bottom line.
// Text is truncated to MaxColumns (16) characters.
//
// Wire format: 0x4D, 0x0C, <line_byte>, <length>, <text_bytes...>
// Line mapping: line 1 → 0x00, line 2 → 0x01.
func buildDisplayText(line int, text string) []byte {
	if len(text) > MaxColumns {
		text = text[:MaxColumns]
	}

	var lineByte byte
	switch line {
	case 1:
		lineByte = 0x00
	default:
		lineByte = 0x01
	}

	cmd := make([]byte, 0, 4+len(text))
	cmd = append(cmd, cmdPreamble, cmdDisplayText, lineByte, byte(len(text)))
	cmd = append(cmd, []byte(text)...)
	return cmd
}
