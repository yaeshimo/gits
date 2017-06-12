// +build !windows

package main

import (
	"os"
	"os/user"
	"path/filepath"
)

var defConfPath = ""
var defWorkDir = ""

func init() {
	u, err := user.Current()
	if err != nil {
		return
	}
	if u.HomeDir == "" {
		// unreachable?
		return
	}
	defWorkDir = filepath.Join(u.HomeDir, ".config", "gits")
	if err := os.Chdir(defWorkDir); err != nil {
		defWorkDir = ""
	}
	defConfPath = filepath.Join(defWorkDir, "watchlist.json")
}
