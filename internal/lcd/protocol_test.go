package lcd

import (
	"bytes"
	"testing"
)

func TestBuildDisplayText(t *testing.T) {
	tests := []struct {
		name     string
		line     int
		text     string
		expected []byte
	}{
		{
			name:     "line 1 short text",
			line:     1,
			text:     "Hello",
			expected: []byte{0x4D, 0x0C, 0x00, 0x05, 'H', 'e', 'l', 'l', 'o'},
		},
		{
			name:     "line 2 short text",
			line:     2,
			text:     "World",
			expected: []byte{0x4D, 0x0C, 0x01, 0x05, 'W', 'o', 'r', 'l', 'd'},
		},
		{
			name:     "truncate to 16 chars",
			line:     1,
			text:     "12345678901234567890",
			expected: append([]byte{0x4D, 0x0C, 0x00, 16}, []byte("1234567890123456")...),
		},
		{
			name:     "empty text",
			line:     1,
			text:     "",
			expected: []byte{0x4D, 0x0C, 0x00, 0x00},
		},
		{
			name:     "exactly 16 chars",
			line:     2,
			text:     "1234567890123456",
			expected: append([]byte{0x4D, 0x0C, 0x01, 16}, []byte("1234567890123456")...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDisplayText(tt.line, tt.text)
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("buildDisplayText(%d, %q):\n  got  %v\n  want %v", tt.line, tt.text, got, tt.expected)
			}
		})
	}
}

func TestBuildBacklight(t *testing.T) {
	on := buildBacklight(true)
	expected := []byte{0x4D, 0x5E, 0x01}
	if !bytes.Equal(on, expected) {
		t.Errorf("backlight on: got %v, want %v", on, expected)
	}

	off := buildBacklight(false)
	expected = []byte{0x4D, 0x5E, 0x00}
	if !bytes.Equal(off, expected) {
		t.Errorf("backlight off: got %v, want %v", off, expected)
	}
}

func TestBuildSimpleCommands(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() []byte
		expected []byte
	}{
		{"get board id", buildGetBoardID, []byte{0x4D, 0x00}},
		{"get buttons", buildGetButtons, []byte{0x4D, 0x06}},
		{"get protocol", buildGetProtocol, []byte{0x4D, 0x07}},
		{"clear", buildClear, []byte{0x4D, 0x0D}},
		{"reset", buildReset, []byte{0x4D, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if !bytes.Equal(got, tt.expected) {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		et   EventType
		want string
	}{
		{EventBoardID, "BoardID"},
		{EventButtonPress, "ButtonPress"},
		{EventProtocolVersion, "ProtocolVersion"},
		{EventResetOK, "ResetOK"},
		{EventAck, "Ack"},
		{EventNack, "Nack"},
		{EventType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.et.String(); got != tt.want {
			t.Errorf("EventType(%d).String() = %q, want %q", tt.et, got, tt.want)
		}
	}
}

func TestButtonName(t *testing.T) {
	tests := []struct {
		code uint16
		want string
	}{
		{ButtonNone, "Released"},
		{ButtonUp, "Up"},
		{ButtonDown, "Down"},
		{ButtonBoth, "Both"},
		{0x0004, "Unknown"},
	}

	for _, tt := range tests {
		if got := ButtonName(tt.code); got != tt.want {
			t.Errorf("ButtonName(0x%04X) = %q, want %q", tt.code, got, tt.want)
		}
	}
}
