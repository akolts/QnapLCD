package lcd

import (
	"fmt"
	"io"
)

// Port abstracts serial port operations for testability.
// The go.bug.st/serial.Port type satisfies this interface.
type Port interface {
	io.ReadWriteCloser
}

// Device controls the QNAP A125 LCD display over a serial port.
type Device struct {
	port    Port
	columns int
	lines   int
}

// NewDevice creates a new LCD device using the given serial port.
func NewDevice(port Port) *Device {
	return &Device{
		port:    port,
		columns: MaxColumns,
		lines:   MaxLines,
	}
}

// WriteLine writes text to the specified display line.
// Line 1 is the top line, line 2 is the bottom line.
// Text longer than 16 characters is truncated.
func (d *Device) WriteLine(line int, text string) error {
	if line < 1 || line > d.lines {
		return fmt.Errorf("lcd: line %d out of range (1-%d)", line, d.lines)
	}
	_, err := d.port.Write(buildDisplayText(line, text))
	return err
}

// WriteLines writes text to both display lines at once.
func (d *Device) WriteLines(line1, line2 string) error {
	if err := d.WriteLine(1, line1); err != nil {
		return err
	}
	return d.WriteLine(2, line2)
}

// Clear clears the display.
func (d *Device) Clear() error {
	_, err := d.port.Write(buildClear())
	return err
}

// Backlight turns the display backlight on or off.
func (d *Device) Backlight(on bool) error {
	_, err := d.port.Write(buildBacklight(on))
	return err
}

// Reset sends a reset command to the LCD controller.
func (d *Device) Reset() error {
	_, err := d.port.Write(buildReset())
	return err
}

// GetBoardID requests the board ID from the LCD controller.
// The response arrives as an EventBoardID on the reader channel.
func (d *Device) GetBoardID() error {
	_, err := d.port.Write(buildGetBoardID())
	return err
}

// GetProtocol requests the protocol version from the LCD controller.
// The response arrives as an EventProtocolVersion on the reader channel.
func (d *Device) GetProtocol() error {
	_, err := d.port.Write(buildGetProtocol())
	return err
}

// GetButtons requests the current button status from the LCD controller.
// The response arrives as an EventButtonPress on the reader channel.
func (d *Device) GetButtons() error {
	_, err := d.port.Write(buildGetButtons())
	return err
}

// Close closes the underlying serial port.
func (d *Device) Close() error {
	return d.port.Close()
}
