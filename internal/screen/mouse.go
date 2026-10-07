// SGR mouse frame helpers: enabling the terminal mouse protocol and
// mapping terminal rows back to item indexes through the viewport.
package screen

// ItemFirstRow is the 1-indexed terminal row of the first item in a
// list frame (title=1, top border=2, items start at 3). The JS stack
// renders two extra escape-only lines first, so its first item row is
// 5 (see lib/screen/mouse.js FIRST_ITEM_ROW).
const ItemFirstRow = 3

// ItemIndexAtRow maps a 1-indexed terminal row to an item index,
// honoring the viewport around selected. It returns false when the
// row is chrome (title/borders/footer) or outside the visible items.
func ItemIndexAtRow(items []Item, selected, height, termRow int) (int, bool) {
	if len(items) == 0 {
		return 0, false
	}
	rel := termRow - ItemFirstRow
	if rel < 0 {
		return 0, false
	}
	visible, offset := Viewport(items, selected, height)
	used := 0
	for i, it := range visible {
		need := linesPerItem(it)
		if rel >= used && rel < used+need {
			return offset + i, true
		}
		used += need
	}
	return 0, false
}

// EnterFrame is the byte prefix written once when the fullscreen loop
// starts; LeaveFrame restores the previous screen and cursor.
func EnterFrame() string {
	return AltEnter + Clear + Home + HideCursor + MouseEnable
}

// LeaveFrame is written once on every exit path, including errors.
func LeaveFrame() string {
	return MouseDisable + ShowCursor + AltLeave
}
