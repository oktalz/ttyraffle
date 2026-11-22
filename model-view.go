package main

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m *model) View() string {
	return m.view
}

func (m *model) renderView() string {
	menu := "↑↓ scroll · q quit"
	switch m.state {
	case StatePaused:
		menu = "s start/skip · c continue · x reset · i insert · d delete · r shuffle · ↑↓ scroll · q quit"
		if m.w < 90 {
			menu = "s/c draw · x reset · i/d edit · r shuffle · ↑↓ · q quit"
		}
		drawn := 0
		for _, u := range m.users {
			if u.Removed {
				drawn++
			}
		}
		if len(m.users) > 0 && drawn == len(m.users) {
			menu = "all items drawn · x reset · i insert · q quit"
		}
	case StateAddUser:
		menu = "enter add · esc cancel · ctrl+c quit"
	case StateChooseFile:
		menu = "↑↓ select · enter open · q quit"
	case StateDeleteUser:
		menu = "↑↓ select · enter delete · esc cancel · q quit"
	}
	if m.err != nil {
		menu = "Error: " + cleanName(m.err.Error())
	}
	content := m.vp.View() + "\n" + ansi.Truncate(menu, m.vp.Width, "…")
	if m.w < 3 || m.h < 4 {
		return ansi.Truncate("Resize terminal", m.w, "")
	}
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(content)
}
