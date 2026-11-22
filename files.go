package main

import (
	"fmt"
	"path/filepath"
)

func (m *model) readFiles(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: ttyraffle [file.txt]")
	}
	files := args
	if len(files) == 0 {
		var err error
		files, err = filepath.Glob("*.txt")
		if err != nil {
			return err
		}
	}
	switch len(files) {
	case 0:
		return fmt.Errorf("no .txt files found in the current directory")
	case 1:
		return m.readUsers(files[0])
	default:
		m.state, m.choices = StateChooseFile, files
		return nil
	}
}
