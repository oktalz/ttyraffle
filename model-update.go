package main

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Per-bar frames must not redraw the entire list.
	if _, frame := msg.(progress.FrameMsg); !frame {
		defer m.updateViewportContent()
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || (key == "q" && m.state != StateAddUser) {
			return m, tea.Quit
		}
		switch m.state {
		case StateChooseFile:
			switch key {
			case "up":
				m.cursor = max(0, m.cursor-1)
			case "down":
				m.cursor = min(len(m.choices)-1, m.cursor+1)
			case "enter":
				if m.cursor >= 0 && m.cursor < len(m.choices) {
					m.err = m.readUsers(m.choices[m.cursor])
					if m.err == nil {
						m.vp.GotoTop()
					}
				}
			}
			m.keepSelectionVisible()
			return m, nil
		case StateAddUser:
			switch key {
			case "enter":
				if name := cleanName(m.textInput.Value()); name != "" {
					m.users = append(m.users, User{Name: name, progress: progress.New(progress.WithDefaultGradient())})
					m.recalculateMaxNameLen()
				}
				m.textInput.Reset()
				m.state = StatePaused
			case "esc":
				m.textInput.Reset()
				m.state = StatePaused
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
			return m, nil
		case StateDeleteUser:
			switch key {
			case "up":
				m.selectedUser = max(0, m.selectedUser-1)
			case "down":
				m.selectedUser = min(len(m.users)-1, m.selectedUser+1)
			case "enter":
				if m.selectedUser >= 0 && m.selectedUser < len(m.users) {
					m.users = append(m.users[:m.selectedUser], m.users[m.selectedUser+1:]...)
					if m.winner == m.selectedUser {
						m.winner = -1
					} else if m.winner > m.selectedUser {
						m.winner--
					}
					m.recalculateMaxNameLen()
				}
				m.state = StatePaused
			case "esc":
				m.state = StatePaused
			}
			m.keepSelectionVisible()
			return m, nil
		}
		if m.state == StatePaused {
			switch key {
			case "r":
				rand.Shuffle(len(m.users), func(i, j int) {
					m.users[i], m.users[j] = m.users[j], m.users[i]
					if m.winner == i {
						m.winner = j
					} else if m.winner == j {
						m.winner = i
					}
				})
			case "i", "insert":
				m.state = StateAddUser
				m.vp.GotoTop()
				return m, m.textInput.Focus()
			case "d", "delete":
				if len(m.users) > 0 {
					m.state = StateDeleteUser
					m.selectedUser = 0
					m.vp.GotoTop()
				}
			case "x":
				m.reset()
			case "s", "c", " ", "enter":
				m.start(key == "s")
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.updateSize(msg.Width, msg.Height)
		m.keepSelectionVisible()
	case tickMsg:
		updateUsers(m, time.Time(msg))
		cmds := []tea.Cmd{tickCmd()}
		for i := range m.users {
			percent := float64(m.users[i].Percentage) / 100
			if m.users[i].progress.Percent() != percent {
				cmds = append(cmds, m.users[i].progress.SetPercent(percent))
			}
		}
		return m, tea.Batch(cmds...)
	case progress.FrameMsg:
		var cmds []tea.Cmd
		for i := range m.users {
			updated, cmd := m.users[i].progress.Update(msg)
			m.users[i].progress = updated.(progress.Model)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	default:
		if m.state == StateAddUser {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) keepSelectionVisible() {
	var row int
	switch m.state {
	case StateChooseFile:
		row = m.cursor + 1
	case StateDeleteUser:
		row = m.selectedUser + 1
	default:
		return
	}
	m.updateViewportContent()
	if row < m.vp.YOffset {
		m.vp.SetYOffset(row)
	}
	if row >= m.vp.YOffset+m.vp.Height {
		m.vp.SetYOffset(row - m.vp.Height + 1)
	}
}
