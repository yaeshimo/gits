package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

// use json
type repoInfo struct {
	Gitdir   string `json:"gitdir"`
	Workdir  string `json:"workdir"`
	Readonly bool   `json:"readonly"`
}
type watchList map[string]repoInfo

func readWatchList(fpath string) (watchList, error) {
	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	w := new(watchList)
	if err := json.Unmarshal(b, w); err != nil {
		return nil, err
	}
	return *w, nil
}

// TODO: implementation backup
func writeWatchList(w watchList, file string) error {
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
	watch := make(watchList)
	watch["repository name"] = repoInfo{
		Gitdir:   "/path/to/repo/.git",
		Workdir:  "/path/to/repo",
		Readonly: false,
	}
	watch["template"] = repoInfo{
		Gitdir:   "",
		Workdir:  "",
		Readonly: false,
	}
	b, err := json.MarshalIndent(watch, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	fmt.Fprintln(w, string(b))
	return nil
}
