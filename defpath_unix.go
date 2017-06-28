// +build !windows

package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

var (
	DefConfName = "watchlist.json"
	DefConfDirList = []string{}
	DefConfDir = ""
	DefEditor = ""
)

func init() {
	if vim, err := exec.LookPath("vim"); err == nil {
		DefEditor = vim
	}
	DefEditor = os.Getenv("EDITOR")

	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		DefConfDirList = []string{
			filepath.Join(u.HomeDir, ".gits"),
			filepath.Join(u.HomeDir, ".config", "gits"),
		}
		for _, dir := range DefConfDirList {
			if f, err := os.Stat(dir); err == nil && f.IsDir() {
				DefConfDir = dir
				break
			}
		}
	}
}
