package main

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type User struct {
	Name       string
	Percentage int
	Removed    bool
	progress   progress.Model
}

type model struct {
	w, h           int
	vp             viewport.Model
	maxNameLen     int
	users          []User
	winner         int
	state          state
	textInput      textinput.Model
	lastUserUpdate time.Time
	selectedUser   int
	choices        []string
	cursor         int
	err            error
	view           string
}

func newModel() *model {
	ti := textinput.New()
	ti.Placeholder = "New item"
	ti.Focus()
	ti.CharLimit = 200
	m := &model{winner: -1, state: StatePaused, textInput: ti, vp: viewport.New(1, 1)}
	m.updateSize(80, 24)
	return m
}

func tickCmd() tea.Cmd {
	return tea.Tick(15*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m *model) Init() tea.Cmd { return tickCmd() }

func (m *model) recalculateMaxNameLen() {
	m.maxNameLen = 0
	for _, u := range m.users {
		m.maxNameLen = max(m.maxNameLen, lipgloss.Width(u.Name))
	}
	m.maxNameLen = min(m.maxNameLen, 35, max(1, (m.w-2)/2))
	for i := range m.users {
		m.users[i].progress.Width = max(1, m.w-2-m.maxNameLen-2)
	}
}

func (m *model) updateSize(w, h int) {
	m.w, m.h = max(1, w), max(1, h)
	m.vp.Width, m.vp.Height = max(1, m.w-2), max(1, m.h-3)
	m.textInput.Width = max(1, m.w-6)
	m.recalculateMaxNameLen()
	m.updateViewportContent()
}
