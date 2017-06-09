package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// use json
type repoInfo struct {
	Gitdir  string `json:"gitdir"`
	Workdir string `json:"workdir"`
}
type watchList struct {
	// command restriction
	// if Restriction is nil or length 0 then
	// allow all command
	Restriction []string            `json:"restriction"`
	Map         map[string]repoInfo `json:"repository"`
}

func (wl *watchList) isAllow(firstArg string) bool {
	if wl.Restriction == nil || len(wl.Restriction) == 0 {
		return true // Allow all commands
	}
	for _, s := range wl.Restriction {
		if s == firstArg {
			return true
		}
	}
	return false
}

func readWatchList(fpath string) (*watchList, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	wl := &watchList{}
	if err := json.Unmarshal(b, wl); err != nil {
		return nil, err
	}
	return wl, nil
}

// TODO: implementation backup
func writeWatchList(w *watchList, file string) error {
	b, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	if err := ioutil.WriteFile(file, b, 0666); err != nil {
		return err
	}
	return nil
}

// TODO: fix for windows?
func template(w io.Writer) error {
	watch := &watchList{
		Restriction: []string{
			"status",
			"version",
			"fetch",
			"grep",
			"ls-remote",
			"ls-files",
			"ls-tree",
		},
		Map: make(map[string]repoInfo),
	}
	watch.Map["your_repository_name"] = repoInfo{
		Gitdir:  "/path/to/repo/.git",
		Workdir: "/path/to/repo",
	}
	watch.Map["template"] = repoInfo{
		Gitdir:  "",
		Workdir: "",
	}
	b, err := json.MarshalIndent(watch, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	fmt.Fprintln(w, string(b))
	return nil
}
