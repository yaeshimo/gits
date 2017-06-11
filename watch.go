package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

// TODO: test
// stringer
func (wl *watchList) String() string {
	var str string
	if wl.Restriction == nil || len(wl.Restriction) == 0 {
		str = fmt.Sprintf("Allow First Arguments:\n\tAll allow\n")
	} else {
		str = fmt.Sprintf("Allow First Arguments:\n\t%+v\n", wl.Restriction)
	}
	str += fmt.Sprintln("Watch List:")
	for key := range wl.Map {
		str += fmt.Sprintf("\t[%s]\n", key)
	}
	return str
}

// TODO: test
// add
func (wl *watchList) watch(fullpath string, key string) error {
	if _, ok := wl.Map[key]; ok {
		return fmt.Errorf("[%s] is exists", key)
	}
	wl.Map[key] = repoInfo{
		Gitdir:  filepath.Join(fullpath, ".git"),
		Workdir: fullpath,
	}
	return nil
}

// TODO: test
// delete
func (wl *watchList) unwatch(key string) error {
	if _, ok := wl.Map[key]; !ok {
		return fmt.Errorf("[%s] is not exists", key)
	}
	delete(wl.Map, key)
	return nil
}

// TODO: implementation backup
func (wl *watchList) writeFile(file string) error {
	b, err := json.MarshalIndent(wl, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	if f, err := os.Stat(file); err == nil {
		switch {
		case f.IsDir():
			return fmt.Errorf("%s is directory", f.Name())
		case f.Mode().IsRegular():
			// accept override
		default:
			return fmt.Errorf("%s is not regure file", f.Name())
		}
	}
	if err := ioutil.WriteFile(file, b, 0666); err != nil {
		return err
	}
	return nil
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

// for watch, unwatch
// first: Abs path, second: base key
func keyAbs(path string) (fullpath string, key string, err error) {
	fullpath, err = filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	key = filepath.Base(fullpath)
	return fullpath, key, nil
}

// TODO: fix filepath for windows
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
	watch.Map["repo"] = repoInfo{
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
