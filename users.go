package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/bubbles/progress"
)

func cleanName(name string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name))
}

func (m *model) readUsers(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	var users []User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := cleanName(scanner.Text())
		if name != "" {
			users = append(users, User{Name: name, progress: progress.New(progress.WithDefaultGradient())})
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("%s contains no items", filename)
	}
	rand.Shuffle(len(users), func(i, j int) { users[i], users[j] = users[j], users[i] })
	m.users = users
	m.reset()
	m.recalculateMaxNameLen()
	return nil
}

func (m *model) reset() {
	for i := range m.users {
		m.users[i].Removed = false
		m.users[i].Percentage = 0
		m.users[i].progress = progress.New(progress.WithDefaultGradient())
	}
	m.winner = -1
	m.state = StatePaused
	m.err = nil
	m.recalculateMaxNameLen()
}

func (m *model) start(skip bool) {
	if m.state != StatePaused {
		return
	}
	if m.winner >= 0 && !skip {
		m.users[m.winner].Removed = true
	}
	m.winner = -1
	active := false
	nextState := StateRunning
	for i := range m.users {
		m.users[i].Percentage = 0
		active = active || !m.users[i].Removed
		if m.users[i].progress.Percent() != 0 || m.users[i].progress.IsAnimating() {
			nextState = StateResetting
		}
	}
	if active {
		m.state = nextState
		m.lastUserUpdate = time.Time{}
	}
}

func updateUsers(m *model, now time.Time) {
	switch m.state {
	case StateResetting:
		for i := range m.users {
			if m.users[i].progress.Percent() != 0 || m.users[i].progress.IsAnimating() {
				return
			}
		}
		m.state = StateRunning
	case StateWaitProgress:
		for i := range m.users {
			if m.users[i].progress.IsAnimating() {
				return
			}
		}
		m.state = StatePaused
		m.vp.GotoTop() // reveal the winner banner
	case StateRunning:
		activeUsers := 0
		for i := range m.users {
			if !m.users[i].Removed {
				activeUsers++
			}
		}
		// scale pacing so long lists do not take forever
		waitTime := 125
		if activeUsers < 5 {
			waitTime += 30 * (5 - activeUsers)
		}
		if activeUsers > 15 {
			waitTime += 25 * (activeUsers - 15)
		}
		if activeUsers == 1 {
			waitTime = 0
		}
		if now.Sub(m.lastUserUpdate) < time.Duration(waitTime)*time.Millisecond {
			return
		}
		m.lastUserUpdate = now

		maxStep := 10
		if activeUsers > 15 {
			maxStep += (activeUsers - 15) * 5
		}
		var winners []int
		for i := range m.users {
			if m.users[i].Removed {
				continue
			}
			if activeUsers == 1 {
				m.users[i].progress.SetSpringOptions(14, 1)
				m.users[i].Percentage = 100
			} else {
				m.users[i].Percentage = min(100, m.users[i].Percentage+rand.Intn(maxStep))
			}
			if m.users[i].Percentage >= 99 {
				winners = append(winners, i)
			}
		}
		if len(winners) > 0 {
			m.winner = winners[rand.Intn(len(winners))]
			m.users[m.winner].Percentage = 100
			for i := range m.users {
				if i != m.winner && m.users[i].Percentage > 99 {
					m.users[i].Percentage = 99
				}
			}
			m.state = StateWaitProgress
		}
	}
}
