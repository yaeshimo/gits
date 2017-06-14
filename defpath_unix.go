// +build !windows

package main

import (
	"os"
	"os/user"
	"path/filepath"
)

var defConfName = "watchlist.json"
var defConfDirList = []string{}
var defConfDir = ""

func init() {
	u, err := user.Current()
	if err != nil {
		return
	}
	if u.HomeDir == "" {
		// unreachable?
		return
	}

	defConfDirList = []string{
		filepath.Join(u.HomeDir, "gits"),
		filepath.Join(u.HomeDir, ".config", "gits"),
	}
	for _, dir := range defConfDirList {
		if f, err := os.Stat(dir); err == nil && f.IsDir() {
			defConfDir = dir
			break
		}
	}
}
