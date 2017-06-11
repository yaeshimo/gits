package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	_, err := exec.LookPath("git")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "test is stopped")
		os.Exit(2)
	}
	defConfPath = ""
	os.Exit(m.Run())
}

// TODO: be graceful
//     : fix for windows
func TestRun(t *testing.T) {
	gitdir, err := ioutil.TempDir("", "gits_gitdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(gitdir)
	conf, err := ioutil.TempFile("", "gits_test_conf_json")
	if err != nil {
		t.Fatal(err)
	}
	defer conf.Close()
	defer os.Remove(conf.Name())
	var defaultContent = []byte(`{
  "restriction": [
    "version",
    "status"
  ],
  "repository":{
    "test_repository": {
      "gitdir": "` + filepath.Join(gitdir, ".git") + `",
      "workdir": "` + gitdir + `"
    }
  }
}`)
	type testData struct {
		args    []string
		wanterr bool
	}
	testRun := func(t *testing.T, tests []testData) {
		var s, errs string
		buf := bytes.NewBufferString(s)
		errbuf := bytes.NewBufferString(errs)
		for i, test := range tests {
			exitCode := run(buf, errbuf, nil, test.args)
			switch exitCode {
			case validExit:
				if test.wanterr {
					t.Errorf("t.Errorf [%d] expected error but nil", i)
				}
			case exitWithErr:
				if test.wanterr {
					t.Logf("t.Logf [%d] passed error: %+v", i, errbuf)
				} else {
					t.Errorf("t.Errorf [%d] err: %+v", i, errbuf)
				}
			}
			t.Logf("t.Logf [%d] outbuf: %+v", i, buf)
			buf.Reset()
			errbuf.Reset()
		}
	}

	t.Run("not modify conf", func(t *testing.T) {
		if err := conf.Truncate(0); err != nil {
			t.Fatal(err)
		}
		if _, err := conf.WriteAt(defaultContent, 0); err != nil {
			t.Fatal(err)
		}

		// for error check, -conf=vanishedFilePath
		vanishedFilePath, err := ioutil.TempDir("", "gits_vanished_file_path")
		if err != nil {
			t.Fatal(err)
		}
		os.Remove(vanishedFilePath)

		tests := []testData{
			// valid args
			{
				args:    []string{"gits", "-version"},
				wanterr: false,
			},
			{
				args:    []string{"gits", "-template"},
				wanterr: false,
			},
			{
				args:    []string{"gits", "-list"},
				wanterr: false,
			},
			{
				args:    []string{"gits", "version"},
				wanterr: false,
			},
			// invalid args
			{
				args:    []string{"gits"},
				wanterr: true,
			},
			{
				args:    []string{"gits", `-git=""`, "version"},
				wanterr: true,
			},
			{
				args:    []string{"gits", "status", "--invalid--git--flags"},
				wanterr: true,
			},
			{
				args:    []string{"gits", "not implementation"},
				wanterr: true,
			},
			// conf valid
			{
				args:    []string{"gits", "-conf", conf.Name(), "version"},
				wanterr: false,
			},
			{
				args:    []string{"gits", "-conf", conf.Name(), "-list"},
				wanterr: false,
			},
			// conf invalid
			{
				args:    []string{"gits", "-conf", vanishedFilePath},
				wanterr: true,
			},
			{
				args:    []string{"gits", "-conf", conf.Name(), "fetch"},
				wanterr: true,
			},
			{
				args:    []string{"gits", "-conf", conf.Name(), "status", "--invalid-git-flags"},
				wanterr: true,
			},
		}
		testRun(t, tests)
	})

	t.Run("modify conf", func(t *testing.T) {
		if err := conf.Truncate(0); err != nil {
			t.Fatal(err)
		}
		if _, err := conf.WriteAt(defaultContent, 0); err != nil {
			t.Fatal(err)
		}
		prefix := []string{"gits", "-conf", conf.Name()}
		dir, err := ioutil.TempDir("", "gits_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dir)
		tests := []testData{
			{
				args:    append(prefix, "-watch", dir),
				wanterr: false,
			},
			{
				args:    append(prefix, "-watch", dir),
				wanterr: true, // already watched
			},
			{
				args:    append(prefix, "-unwatch", dir),
				wanterr: false,
			},
			{
				args:    append(prefix, "-unwatch", dir),
				wanterr: true, // already is not watched
			},
		}
		testRun(t, tests)
	})
}
