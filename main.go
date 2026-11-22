package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/oktalz/version"
)

func main() {
	_ = version.Set()
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		fmt.Println(logo)
		fmt.Println("ttyraffle", version.Version)
		fmt.Println("built-from", version.Repo)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalENV := filepath.Join(homeDir, ".ttyraffle", ".env")
	if err := godotenv.Overload(globalENV); err != nil && !os.IsNotExist(err) {
		return err
	}
	if wd := os.Getenv("WD"); wd != "" {
		if err := os.Chdir(wd); err != nil {
			return err
		}
	}
	m := newModel()
	if err := m.readFiles(os.Args[1:]); err != nil {
		return err
	}
	m.updateViewportContent()
	_, err = tea.NewProgram(m).Run()
	return err
}
