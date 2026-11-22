package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m *model) updateViewportContent() {
	var s strings.Builder
	switch m.state {
	case StateChooseFile:
		s.WriteString("Select file\n")
		for i, choice := range m.choices {
			marker := "  "
			if i == m.cursor {
				marker = "> "
			}
			s.WriteString(ansi.Truncate(marker+cleanName(choice), m.vp.Width, "…") + "\n")
		}
	case StateAddUser:
		s.WriteString("Enter new item\n\n" + m.textInput.View())
	case StateDeleteUser:
		s.WriteString("Select item to delete\n")
		for i, u := range m.users {
			marker := "  "
			if i == m.selectedUser {
				marker = "> "
			}
			s.WriteString(ansi.Truncate(marker+u.Name, m.vp.Width, "…") + "\n")
		}
	default:
		logoArt := strings.TrimRight(logo, "\n")
		logoW, logoH := lipgloss.Width(logoArt), lipgloss.Height(logoArt)
		banner := m.winnerBanner()
		bannerW, bannerH := lipgloss.Width(banner), lipgloss.Height(banner)
		const minRows = 3 // keep entries visible under the header
		switch {
		case banner != "" && m.vp.Width >= logoW+bannerW+3 && m.vp.Height >= logoH+minRows:
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, logoArt, "   ", banner) + "\n\n")
		case banner != "" && m.vp.Height >= bannerH+minRows:
			s.WriteString(banner + "\n\n")
		case m.vp.Width >= logoW && m.vp.Height >= logoH+minRows:
			s.WriteString(logoArt + "\n\n")
		}
		if len(m.users) == 0 {
			s.WriteString("No items. Press i to add one.\n")
		}
		for _, u := range m.users {
			s.WriteString(m.renderUser(u) + "\n")
		}
	}
	m.vp.SetContent(s.String())
	m.view = m.renderView()
}

func (m *model) winnerBanner() string {
	if m.state != StatePaused || m.winner < 0 || m.winner >= len(m.users) {
		return ""
	}
	name := ansi.Truncate(m.users[m.winner].Name, max(1, m.vp.Width-8), "…")
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#045e9a")).
		Padding(1, 3).
		Render(name)
}

func (m *model) renderUser(u User) string {
	name := ansi.Truncate(u.Name, m.maxNameLen, "…")
	name += strings.Repeat(" ", max(0, m.maxNameLen-lipgloss.Width(name)))
	if u.Removed {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(ansi.Truncate(name+" (drawn)", m.vp.Width, "…"))
	}
	return ansi.Truncate(name+" "+u.progress.View(), m.vp.Width, "…")
}
