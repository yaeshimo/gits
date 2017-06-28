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

func (wl *watchList) unwatch(key string) error {
	if _, ok := wl.Map[key]; !ok {
		return fmt.Errorf("[%s] is not exists", key)
	}
	delete(wl.Map, key)
	return nil
}

// TODO: implemetation wirte backup
func (wl *watchList) writeFile(file string) error {
	b, err := json.MarshalIndent(wl, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	if info, err := os.Stat(file); err == nil {
		switch {
		case info.IsDir():
			return fmt.Errorf("%s is directory", info.Name())
		case info.Mode().IsRegular():
			// accept override
		default:
			return fmt.Errorf("%s is not regure file", info.Name())
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
	wl := &watchList{Map: make(map[string]repoInfo)}
	if err := json.Unmarshal(b, wl); err != nil {
		return nil, err
	}
	return wl, nil
}

/// TODO: reconsider
type watchConf struct {
	wl   *watchList
	path string
}

/// TODO: reconsider
func newWatchConf(path string) *watchConf {
	return &watchConf{
		wl:   &watchList{Map: make(map[string]repoInfo)},
		path: path,
	}
}
func (wc *watchConf) watch(repoPath string) (string, error) {
	fullpath, key, err := absWithBase(repoPath)
	if err != nil {
		return "", err
	}
	if err := wc.wl.watch(fullpath, key); err != nil {
		return "", err
	}
	if err := wc.wl.writeFile(wc.path); err != nil {
		return "", err
	}
	return key, nil
}
func (wc *watchConf) unwatch(repoPath string) (string, error) {
	_, key, err := absWithBase(repoPath)
	if err != nil {
		return "", err
	}
	if err := wc.wl.unwatch(key); err != nil {
		return "", err
	}
	if err := wc.wl.writeFile(wc.path); err != nil {
		return "", err
	}
	return key, nil
}
func (wc *watchConf) getConfList() ([]string, error) {
	infos, err := ioutil.ReadDir(filepath.Dir(wc.path))
	if err != nil {
		return nil, err
	}
	var s []string
	for _, info := range infos {
		if info.Mode().IsRegular() {
			s = append(s, info.Name())
		}
	}
	return s, nil
}
func createConf(mkpath string) error {
	if _, err := os.Stat(mkpath); err == nil {
		return fmt.Errorf(mkpath + " is exist")
	} else if os.IsNotExist(err) {
		f, err := os.Create(mkpath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := template(f); err != nil {
			return err
		}
		return nil
	} else {
		return err
	}
}

// for watch, unwatch
// first: Abs path, second: base key
func absWithBase(path string) (fullpath string, base string, err error) {
	fullpath, err = filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	return fullpath, filepath.Base(fullpath), nil
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
	watch.Map["repo2"] = repoInfo{
		Gitdir:  "/another/repo2/.git",
		Workdir: "/another/repo2",
	}
	b, err := json.MarshalIndent(watch, "", "  ")
	if err != nil {
		return err // unreachable?
	}
	fmt.Fprintln(w, string(b))
	return nil
}
