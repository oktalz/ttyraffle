package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func key(m *model, value string) tea.Cmd {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
	switch value {
	case "enter":
		msg.Type = tea.KeyEnter
	case "ctrl+c":
		msg.Type = tea.KeyCtrlC
	case "down":
		msg.Type = tea.KeyDown
	}
	_, cmd := m.Update(msg)
	return cmd
}

func testModel(names ...string) *model {
	m := newModel()
	for _, name := range names {
		m.users = append(m.users, User{Name: name, progress: progress.New()})
	}
	m.recalculateMaxNameLen()
	return m
}

func TestEmptyDeletion(t *testing.T) {
	m := testModel()
	key(m, "d")
	if m.state != StatePaused {
		t.Fatal("entered empty deletion")
	}
	m.state = StateDeleteUser
	key(m, "enter")
	m = testModel("only")
	key(m, "d")
	key(m, "enter")
	key(m, "d")
	if len(m.users) != 0 || m.state != StatePaused {
		t.Fatal("last-item deletion failed")
	}
}

func TestQuitEveryState(t *testing.T) {
	for _, state := range []state{StatePaused, StateRunning, StateResetting, StateWaitProgress, StateChooseFile, StateAddUser, StateDeleteUser} {
		t.Run(string(state), func(t *testing.T) {
			m := testModel("one")
			m.state = state
			cmd := key(m, "ctrl+c")
			if cmd == nil {
				t.Fatal("no quit command")
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Fatal("not a quit command")
			}
		})
	}
	m := testModel()
	m.state = StateAddUser
	key(m, "q")
	if m.textInput.Value() != "q" {
		t.Fatal("cannot type q")
	}
}

func TestReadUsers(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "items.txt")
	m := testModel("original")
	for _, content := range []string{" \n\t\n", strings.Repeat("a", 100000)} {
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if err := m.readUsers(file); err == nil {
			t.Fatal("expected loading error")
		}
		if len(m.users) != 1 || m.users[0].Name != "original" {
			t.Fatal("failed load changed items")
		}
	}
	if err := os.WriteFile(file, []byte(" Alice \n\nAlice\nBob\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.readUsers(file); err != nil {
		t.Fatal(err)
	}
	if len(m.users) != 3 {
		t.Fatal("blank lines or duplicates mishandled")
	}
	m.state = StateChooseFile
	m.choices = []string{filepath.Join(dir, "missing.txt")}
	key(m, "enter")
	if m.err == nil || m.state != StateChooseFile {
		t.Fatal("file error not retained")
	}
	if !strings.Contains(m.View(), "Error:") {
		t.Fatal("file error not visible")
	}
}

func TestDrawLifecycle(t *testing.T) {
	m := testModel("one", "two")
	m.winner = 0
	key(m, "s")
	if m.users[0].Removed || m.winner != -1 || m.state != StateRunning {
		t.Fatal("skip excluded winner")
	}
	m.state, m.winner = StatePaused, 0
	key(m, "c")
	if !m.users[0].Removed || m.state != StateRunning {
		t.Fatal("continue did not exclude winner")
	}
	m.state, m.winner = StatePaused, 1
	key(m, "c")
	if m.state != StatePaused || !strings.Contains(m.View(), "all items drawn") {
		t.Fatal("missing completion")
	}
	key(m, "x")
	for _, u := range m.users {
		if u.Removed || u.Percentage != 0 {
			t.Fatal("reset did not restore items")
		}
	}
	if m.winner != -1 {
		t.Fatal("reset kept winner")
	}
}

func TestDeletionPreservesWinner(t *testing.T) {
	m := testModel("one", "two", "winner")
	m.winner = 2
	key(m, "d")
	key(m, "enter")
	if m.winner != 1 || m.users[m.winner].Name != "winner" {
		t.Fatal("winner shifted incorrectly")
	}
}

func TestScrollingAndUnicode(t *testing.T) {
	m := testModel()
	for i := range 100 {
		m.users = append(m.users, User{Name: fmt.Sprintf("名字🙂%d", i), progress: progress.New()})
	}
	m.updateSize(30, 10)
	key(m, "down")
	if m.vp.YOffset == 0 {
		t.Fatal("viewport did not scroll")
	}
	key(m, "d")
	for range 50 {
		key(m, "down")
	}
	row := m.selectedUser + 1
	if row < m.vp.YOffset || row >= m.vp.YOffset+m.vp.Height {
		t.Fatal("selection not visible")
	}
	for _, width := range []int{0, 1, 2, 5, 20, 80} {
		m.updateSize(width, 5)
		if m.vp.Width < 1 || m.vp.Height < 1 {
			t.Fatal("invalid dimensions")
		}
		rendered := m.renderUser(m.users[0])
		if !utf8.ValidString(rendered) || lipgloss.Width(rendered) > m.vp.Width {
			t.Fatal("invalid unicode layout")
		}
		_ = m.View()
	}
}

func TestNoExtraTickOnStart(t *testing.T) {
	m := testModel("one")
	if cmd := key(m, "s"); cmd != nil {
		t.Fatal("start created another tick chain")
	}
	if cmd := key(m, "s"); cmd != nil {
		t.Fatal("repeated start created another tick chain")
	}
}

func TestHeaderFitsCommonTerminals(t *testing.T) {
	logoLine := strings.Split(strings.TrimRight(logo, "\n"), "\n")[1]
	m := testModel("Alice")
	m.updateSize(80, 24)
	m.updateViewportContent()
	if !strings.Contains(m.vp.View(), logoLine) {
		t.Fatal("banner hidden on an 80-column terminal")
	}
	if !strings.Contains(m.vp.View(), "Alice") {
		t.Fatal("entry hidden under the banner")
	}
	m.updateSize(80, 8)
	m.updateViewportContent()
	if strings.Contains(m.vp.View(), logoLine) {
		t.Fatal("banner hides entries on a short terminal")
	}
	if !strings.Contains(m.vp.View(), "Alice") {
		t.Fatal("entry hidden on a short terminal")
	}
}

func TestWinnerBannerDoesNotShiftRows(t *testing.T) {
	m := testModel("Alice", "Bob")
	m.updateSize(120, 24)
	m.state, m.winner = StateWaitProgress, 0
	m.updateViewportContent()
	before := strings.Split(m.vp.View(), "\n")
	m.state = StatePaused
	m.updateViewportContent()
	after := strings.Split(m.vp.View(), "\n")
	for i, line := range before {
		if strings.Contains(line, "0%") && after[i] != line {
			t.Fatalf("participant row %d moved when the winner appeared", i)
		}
	}
	if !strings.Contains(after[1], "Alice") {
		t.Fatal("winner name is not aligned with the middle of the logo")
	}
}

func TestWinnerRevealedAfterAnimation(t *testing.T) {
	m := testModel("bob", "alice", "carol")
	m.users[1].Removed, m.users[2].Removed = true, true
	m.updateSize(80, 6)
	m.start(false)
	now := time.Now()
	_, cmd := m.Update(tickMsg(now))
	if m.state != StateWaitProgress || !m.users[0].progress.IsAnimating() {
		t.Fatal("race did not wait for the final animation")
	}
	if m.winnerBanner() != "" {
		t.Fatal("winner revealed before the final animation")
	}
	cmds := cmd().(tea.BatchMsg)
	frameCmd := cmds[1]
	for frames := 0; m.users[0].progress.IsAnimating(); frames++ {
		if frames >= 120 || frameCmd == nil {
			t.Fatal("animation did not finish")
		}
		before := m.View()
		_, frameCmd = m.Update(frameCmd())
		if m.View() != before {
			t.Fatal("individual frame rebuilt the screen instead of waiting for the tick")
		}
		if !m.users[0].progress.IsAnimating() {
			break
		}
		now = now.Add(15 * time.Millisecond)
		m.Update(tickMsg(now))
		if m.state != StateWaitProgress || m.winnerBanner() != "" {
			t.Fatal("winner revealed while the bar was still animating")
		}
	}
	m.vp.SetYOffset(1)
	m.Update(tickMsg(now.Add(15 * time.Millisecond)))
	if m.state != StatePaused || !strings.Contains(m.winnerBanner(), "bob") {
		t.Fatal("winner banner missing after animation")
	}
	if lipgloss.Height(m.winnerBanner()) < 3 || m.vp.YOffset != 0 {
		t.Fatal("prominent winner banner not revealed at the top")
	}
	if !strings.Contains(m.renderUser(m.users[m.winner]), "100%") {
		t.Fatal("winner revealed before the rendered bar reached 100%")
	}
	key(m, "down")
	if m.vp.YOffset == 0 {
		t.Fatal("winner reveal prevents scrolling")
	}
}

func TestNextRunWaitsForZero(t *testing.T) {
	for _, index := range []int{0, 1} {
		t.Run(fmt.Sprintf("bar%d", index), func(t *testing.T) {
			m := testModel("winner", "remaining")
			m.users[index].Percentage = 100
			frameCmd := m.users[index].progress.SetPercent(1)
			for frames := 0; m.users[index].progress.IsAnimating(); frames++ {
				if frames >= 120 || frameCmd == nil {
					t.Fatal("initial animation did not finish")
				}
				_, frameCmd = m.Update(frameCmd())
			}
			m.winner = 0
			if cmd := key(m, " "); cmd != nil {
				t.Fatal("space created an extra tick chain")
			}
			if m.state != StateResetting || !m.users[0].Removed {
				t.Fatal("next run did not enter reset phase")
			}
			key(m, " ")
			now := time.Now()
			_, cmd := m.Update(tickMsg(now))
			if m.state != StateResetting || !m.users[index].progress.IsAnimating() {
				t.Fatal("race started before the zero target was applied")
			}
			frameCmd = cmd().(tea.BatchMsg)[1]
			for frames := 0; m.state == StateResetting; frames++ {
				if frames >= 120 || frameCmd == nil {
					t.Fatal("reset animation did not finish")
				}
				_, frameCmd = m.Update(frameCmd())
				now = now.Add(15 * time.Millisecond)
				m.Update(tickMsg(now))
				for _, u := range m.users {
					if u.Percentage != 0 || u.progress.Percent() != 0 {
						t.Fatal("race advanced before all bars returned to zero")
					}
				}
				if m.users[index].progress.IsAnimating() && m.state != StateResetting {
					t.Fatal("race started while a bar was still resetting")
				}
			}
			if m.state != StateRunning || !strings.Contains(m.users[index].progress.View(), "  0%") {
				t.Fatal("race did not resume from a rendered zero")
			}
			m.Update(tickMsg(now.Add(time.Second)))
			if m.lastUserUpdate.IsZero() {
				t.Fatal("race did not resume after resetting")
			}
		})
	}
}

func TestViewReusesRenderedScreen(t *testing.T) {
	m := testModel("bob")
	m.updateViewportContent()
	if m.View() == "" {
		t.Fatal("initial screen missing")
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = m.View() }); allocations != 0 {
		t.Fatalf("cached screen allocated %g times per frame", allocations)
	}
	key(m, "i")
	if !strings.Contains(m.View(), "Enter new item") {
		t.Fatal("cached screen did not refresh after input")
	}
}

func TestNoRemainingCounter(t *testing.T) {
	m := testModel("alice", "bob")
	for _, state := range []state{StatePaused, StateRunning, StateWaitProgress} {
		m.state = state
		m.updateViewportContent()
		if strings.Contains(m.View(), "items remaining") {
			t.Fatalf("remaining counter rendered in %s", state)
		}
	}
}

func TestPacingScalesWithList(t *testing.T) {
	m := testModel()
	for i := range 40 {
		m.users = append(m.users, User{Name: fmt.Sprintf("u%d", i), progress: progress.New()})
	}
	m.start(false)
	now := time.Now()
	old := now.Add(-200 * time.Millisecond)
	m.lastUserUpdate = old
	updateUsers(m, now)
	if !m.lastUserUpdate.Equal(old) {
		t.Fatal("large list advanced too often")
	}
	m.users = m.users[:3]
	m.lastUserUpdate = old
	updateUsers(m, now)
	if !m.lastUserUpdate.Equal(now) {
		t.Fatal("small list waited too long")
	}
}

func TestSingleParticipantStartsSmoothly(t *testing.T) {
	for _, count := range []int{1, 3} {
		t.Run(fmt.Sprintf("entries%d", count), func(t *testing.T) {
			m := testModel("last")
			for i := 1; i < count; i++ {
				m.users = append(m.users, User{Name: "drawn", Removed: true, progress: progress.New()})
			}
			key(m, " ")
			_, cmd := m.Update(tickMsg(time.Now()))
			if m.winner != 0 || m.state != StateWaitProgress || m.users[0].Percentage != 100 {
				t.Fatal("single participant did not immediately target the finish")
			}
			if !m.users[0].progress.IsAnimating() || m.winnerBanner() != "" {
				t.Fatal("single participant skipped the animation")
			}
			before := m.users[0].progress.View()
			frameCmd := cmd().(tea.BatchMsg)[1]
			frames := 0
			for ; m.users[0].progress.IsAnimating(); frames++ {
				if frames >= 120 || frameCmd == nil {
					t.Fatal("sweep did not finish")
				}
				_, frameCmd = m.Update(frameCmd())
				if frames == 0 {
					after := m.users[0].progress.View()
					if after == before || strings.Contains(after, "100%") {
						t.Fatal("bar did not move through an intermediate animation frame")
					}
				}
			}
			if frames < 20 || frames > 60 {
				t.Fatalf("single-bar sweep took %d frames, want a short but visible sweep", frames)
			}
		})
	}
}

func TestRaceCompletes(t *testing.T) {
	m := testModel("one")
	m.start(false)
	now := time.Now()
	for i := 0; i < 1000 && m.state == StateRunning; i++ {
		updateUsers(m, now.Add(time.Duration(i)*time.Second))
	}
	if m.winner != 0 || m.state != StateWaitProgress {
		t.Fatal("race did not finish")
	}
}
