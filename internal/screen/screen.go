// Package screen owns the fullscreen TUI primitives: palette, frame
// rendering, viewport scrolling, and ANSI key parsing. All rendering is
// pure string building (stdlib only); console setup lives behind the
// windows/linux/other build split so the package compiles everywhere.
// The TUI above this package is identical on every console: same views,
// keys, mouse and palette.
package screen

import (
	"strings"
)

// Item is one selectable row. Detail renders as a dim second line and may
// be empty (menu rows use Label only).
type Item struct {
	Label  string
	Detail string
	State  string
}

// Frame is one screen: bordered title, a scrollable item list with one
// selected row, and a key footer. Height caps visible item rows (details
// included); 0 means no cap.
type Frame struct {
	Title    string
	Items    []Item
	Selected int
	Footer   string
	Height   int
}

// linesPerItem reports how many terminal rows one item occupies.
func linesPerItem(it Item) int {
	if it.Detail == "" {
		return 1
	}
	return 2
}

// fitRow fits one pre-selection row into the window, restarting the
// window at (or just after) the row on overflow.
func fitRow(start, used, i, need, height int, isSel bool) (int, int) {
	if used+need <= height {
		return start, used + need
	}
	if isSel {
		return i, need
	}
	return i + 1, 0
}

// Viewport slices items to the rows around selected so the selection stays
// visible. It returns the visible items and how many leading items were
// skipped. Height <= 0 disables scrolling.
func Viewport(items []Item, selected, height int) ([]Item, int) {
	if height <= 0 || len(items) == 0 {
		return items, 0
	}
	selected = ClampSelection(len(items), selected)
	// Grow the window from the top until the selected row fits.
	start := 0
	used := 0
	for i, it := range items {
		need := linesPerItem(it)
		if i > selected {
			if used+need > height {
				return items[start:i], start
			}
			used += need
			continue
		}
		start, used = fitRow(start, used, i, need, height, i == selected)
	}
	return items[start:], start
}

func border(width int) string {
	if width < 2 {
		width = 2
	}
	return FgRed + "+" + strings.Repeat("-", width-2) + "+" + Reset
}

// cellWidth counts visible runes. ANSI codes must wrap already-padded
// text, never sit inside the measured span otherwise the background ends
// short; byte counting also splits multibyte runes on truncate.
func cellWidth(s string) int {
	return len([]rune(s))
}

func padCell(text string, width int) string {
	runes := []rune(text)
	if len(runes) > width {
		return string(runes[:width])
	}
	return text + strings.Repeat(" ", width-len(runes))
}

// renderLabel paints one item row with its selection marker.
func renderLabel(it Item, selected bool, width int) string {
	marker := "  "
	if selected {
		marker = "> "
	}
	label := it.Label
	if it.State != "" {
		label += " [" + it.State + "]"
	}
	cell := marker + padCell(label, width-len(marker))
	if !selected {
		return cell + "\n"
	}
	return BgDarkRed + FgWhite + Bold + cell + Reset + BgBlack + FgWhite + "\n"
}

// renderDetail paints the dim second line of one item, or nothing.
// Dim wraps the padded text so every row ends on the same column.
func renderDetail(it Item, selected bool, width int) string {
	if it.Detail == "" {
		return ""
	}
	cell := "  " + Dim + padCell(it.Detail, width-2)
	if !selected {
		return cell + Reset + BgBlack + FgWhite + "\n"
	}
	return BgDarkRed + FgWhite + cell + Reset + BgBlack + FgWhite + "\n"
}

// Render builds one full screen: home cursor, title, bordered item list
// with the selected row highlighted, and the footer. Selection marker `>`
// stays visible even on terminals that ignore colors.
func Render(f Frame, width int) string {
	if width < 10 {
		width = 10
	}
	visible, offset := Viewport(f.Items, f.Selected, f.Height)
	selected := ClampSelection(len(f.Items), f.Selected)
	var b strings.Builder
	b.WriteString(Home + EraseBelow)
	b.WriteString(BgBlack + FgWhite)
	b.WriteString(Bold + f.Title + Reset + BgBlack + FgWhite + "\n")
	b.WriteString(border(width) + "\n")
	for i, it := range visible {
		sel := offset+i == selected
		b.WriteString(renderLabel(it, sel, width))
		b.WriteString(renderDetail(it, sel, width))
	}
	b.WriteString(border(width) + "\n")
	b.WriteString(FgGray + f.Footer + Reset + "\n")
	return b.String()
}

// ClampSelection keeps selected inside the list bounds.
func ClampSelection(n, selected int) int {
	if n <= 0 {
		return 0
	}
	if selected < 0 {
		return 0
	}
	if selected >= n {
		return n - 1
	}
	return selected
}

// MoveSteps returns the new selection after steps (negative is up),
// clamped and wrapping around the list.
func MoveSteps(n, selected, steps int) int {
	if n <= 0 {
		return 0
	}
	next := (selected + steps) % n
	if next < 0 {
		next += n
	}
	return next
}

// FrameHeight caps the item area to the terminal rows minus chrome
// (title + 2 borders + footer + margins). It never drops below 1.
func FrameHeight(termRows int) int {
	h := termRows - 6
	if h < 1 {
		return 1
	}
	return h
}
