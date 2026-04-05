package menu

import "testing"

func staticItem(l1, l2 string) Item {
	return func() (string, string) { return l1, l2 }
}

func TestNewMenu(t *testing.T) {
	m := New([]Item{
		staticItem("A1", "A2"),
		staticItem("B1", "B2"),
	})

	if m.Len() != 2 {
		t.Errorf("Len() = %d, want 2", m.Len())
	}
	if m.Index() != 0 {
		t.Errorf("Index() = %d, want 0", m.Index())
	}

	l1, l2 := m.Current()
	if l1 != "A1" || l2 != "A2" {
		t.Errorf("Current() = (%q, %q), want (A1, A2)", l1, l2)
	}
}

func TestNextWraps(t *testing.T) {
	m := New([]Item{
		staticItem("A1", "A2"),
		staticItem("B1", "B2"),
		staticItem("C1", "C2"),
	})

	l1, _ := m.Next() // → B
	if l1 != "B1" {
		t.Errorf("after Next: got %q, want B1", l1)
	}

	l1, _ = m.Next() // → C
	if l1 != "C1" {
		t.Errorf("after Next: got %q, want C1", l1)
	}

	l1, _ = m.Next() // → A (wrap)
	if l1 != "A1" {
		t.Errorf("after Next (wrap): got %q, want A1", l1)
	}
}

func TestPrevWraps(t *testing.T) {
	m := New([]Item{
		staticItem("A1", "A2"),
		staticItem("B1", "B2"),
		staticItem("C1", "C2"),
	})

	l1, _ := m.Prev() // → C (wrap backwards)
	if l1 != "C1" {
		t.Errorf("after Prev: got %q, want C1", l1)
	}

	l1, _ = m.Prev() // → B
	if l1 != "B1" {
		t.Errorf("after Prev: got %q, want B1", l1)
	}

	l1, _ = m.Prev() // → A
	if l1 != "A1" {
		t.Errorf("after Prev: got %q, want A1", l1)
	}
}

func TestEmptyMenu(t *testing.T) {
	m := New(nil)

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}

	l1, l2 := m.Current()
	if l1 != "" || l2 != "" {
		t.Errorf("Current() on empty = (%q, %q), want empty", l1, l2)
	}

	l1, l2 = m.Next()
	if l1 != "" || l2 != "" {
		t.Errorf("Next() on empty = (%q, %q), want empty", l1, l2)
	}

	l1, l2 = m.Prev()
	if l1 != "" || l2 != "" {
		t.Errorf("Prev() on empty = (%q, %q), want empty", l1, l2)
	}
}

func TestSingleItem(t *testing.T) {
	m := New([]Item{staticItem("X1", "X2")})

	l1, l2 := m.Current()
	if l1 != "X1" || l2 != "X2" {
		t.Errorf("Current() = (%q, %q), want (X1, X2)", l1, l2)
	}

	// Next and Prev on single item should stay on that item.
	l1, _ = m.Next()
	if l1 != "X1" {
		t.Errorf("Next() on single = %q, want X1", l1)
	}

	l1, _ = m.Prev()
	if l1 != "X1" {
		t.Errorf("Prev() on single = %q, want X1", l1)
	}
}

func TestSetItems(t *testing.T) {
	m := New([]Item{
		staticItem("A1", "A2"),
		staticItem("B1", "B2"),
		staticItem("C1", "C2"),
	})

	// Navigate to item 1 (B).
	m.Next()
	if m.Index() != 1 {
		t.Fatalf("Index() = %d, want 1", m.Index())
	}

	// Replace with 4 items — index 1 is still valid.
	m.SetItems([]Item{
		staticItem("W1", "W2"),
		staticItem("X1", "X2"),
		staticItem("Y1", "Y2"),
		staticItem("Z1", "Z2"),
	})

	if m.Index() != 1 {
		t.Errorf("after SetItems (valid index): Index() = %d, want 1", m.Index())
	}
	l1, _ := m.Current()
	if l1 != "X1" {
		t.Errorf("after SetItems: Current() = %q, want X1", l1)
	}
}

func TestSetItemsOverflow(t *testing.T) {
	m := New([]Item{
		staticItem("A1", "A2"),
		staticItem("B1", "B2"),
		staticItem("C1", "C2"),
	})

	// Navigate to item 2 (C).
	m.Next()
	m.Next()
	if m.Index() != 2 {
		t.Fatalf("Index() = %d, want 2", m.Index())
	}

	// Replace with 2 items — index 2 is out of bounds, should reset to 0.
	m.SetItems([]Item{
		staticItem("X1", "X2"),
		staticItem("Y1", "Y2"),
	})

	if m.Index() != 0 {
		t.Errorf("after SetItems (overflow): Index() = %d, want 0", m.Index())
	}
}

func TestSetItemsEmpty(t *testing.T) {
	m := New([]Item{staticItem("A1", "A2")})
	m.SetItems(nil)

	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
	if m.Index() != 0 {
		t.Errorf("Index() = %d, want 0", m.Index())
	}

	l1, l2 := m.Current()
	if l1 != "" || l2 != "" {
		t.Errorf("Current() on empty = (%q, %q), want empty", l1, l2)
	}
}

func TestDynamicItem(t *testing.T) {
	counter := 0
	m := New([]Item{
		func() (string, string) {
			counter++
			return "call", "count"
		},
	})

	// Each call to Current() should invoke the item function.
	m.Current()
	m.Current()
	m.Current()

	if counter != 3 {
		t.Errorf("item called %d times, want 3", counter)
	}
}
