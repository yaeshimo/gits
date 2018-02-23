// +build linux

package main

import (
	"os"
	"path/filepath"
)

func init() {
	CandidateConfPaths = []string{}
	home := os.Getenv("HOME")
	if home != "" {
		CandidateConfPaths = append(CandidateConfPaths, filepath.Join(home, ".gits", "gits.json"))
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		CandidateConfPaths = append(CandidateConfPaths, filepath.Join(xdg, "gits", "gits.json"))
	} else if home != "" {
		CandidateConfPaths = append(CandidateConfPaths, filepath.Join(home, ".config", "gits", "gits.json"))
	}

	// TODO: consider
	if editor := os.Getenv("EDITOR"); editor != "" {
		EditorWithArgs = []string{editor}
		// append option
		switch editor {
		case "vim":
			EditorWithArgs = append(EditorWithArgs, "--")
		default:
			EditorWithArgs = []string{editor}
		}
	}
}
