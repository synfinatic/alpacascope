package main

import (
	"strings"

	"fyne.io/fyne/v2/widget"
)

type StatusBox struct {
	TextGrid *widget.TextGrid
	Lines    int
	numLines int
	lines    []string
}

func NewStatusBox(lineCount int) *StatusBox {
	status := widget.NewTextGrid()
	var zeroValue []string

	for i := 0; i < lineCount; i++ {
		zeroValue = append(zeroValue, "")
	}
	status.SetText(strings.Join(zeroValue, "\n"))

	sbox := StatusBox{
		TextGrid: status,
		Lines:    lineCount,
		numLines: 0,
		lines:    []string{},
	}
	return &sbox
}

func (sb *StatusBox) AddLine(line string) {
	sb.lines = append(sb.lines, line)
	for len(sb.lines) > sb.Lines {
		sb.lines = sb.lines[1:]
	}
	displayLines := sb.lines
	for len(displayLines) < sb.Lines {
		displayLines = append(displayLines, "")
	}

	lines := strings.Join(displayLines, "\n")
	sb.TextGrid.SetText(lines)
}

func (sb *StatusBox) Widget() *widget.TextGrid {
	return sb.TextGrid
}

func (sb *StatusBox) Clear() {
	var zeroValue []string
	sb.lines = []string{}
	sb.numLines = 0
	for i := 0; i < sb.Lines; i++ {
		zeroValue = append(zeroValue, "")
	}

	lines := strings.Join(zeroValue, "\n")
	sb.TextGrid.SetText(lines)
}
