// +build windows

package main

import (
	"os/user"
	"path/filepath"
)

var defConfPath = func() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(u.HomeDir, "AppData", "Local", "gits", "watchlist.json")
}()
