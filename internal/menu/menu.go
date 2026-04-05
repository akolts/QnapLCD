// Package menu provides a circular menu for the QNAP LCD display.
// Items are functions that return display content when called, allowing
// dynamic data (like uptime) to refresh on each display.
package menu

// Item is a function that returns the two LCD display lines for a menu entry.
// It is called each time the item needs to be shown, so it can return
// up-to-date information.
type Item func() (line1, line2 string)

// Menu manages a list of display items with circular navigation.
type Menu struct {
	items   []Item
	current int
}

// New creates a menu with the given items. The first item is selected.
func New(items []Item) *Menu {
	return &Menu{items: items}
}

// Len returns the number of items in the menu.
func (m *Menu) Len() int {
	return len(m.items)
}

// Index returns the current item index (0-based).
func (m *Menu) Index() int {
	return m.current
}

// Current returns the content of the currently selected item.
// Returns empty strings if the menu is empty.
func (m *Menu) Current() (string, string) {
	if len(m.items) == 0 {
		return "", ""
	}
	return m.items[m.current]()
}

// Next moves to the next item (wraps around) and returns its content.
func (m *Menu) Next() (string, string) {
	if len(m.items) == 0 {
		return "", ""
	}
	m.current = (m.current + 1) % len(m.items)
	return m.Current()
}

// Prev moves to the previous item (wraps around) and returns its content.
func (m *Menu) Prev() (string, string) {
	if len(m.items) == 0 {
		return "", ""
	}
	m.current = (m.current - 1 + len(m.items)) % len(m.items)
	return m.Current()
}

// SetItems replaces the menu items. If the current index is out of bounds
// for the new item list, it wraps to the first item.
func (m *Menu) SetItems(items []Item) {
	m.items = items
	if len(items) == 0 {
		m.current = 0
	} else if m.current >= len(items) {
		m.current = 0
	}
}
