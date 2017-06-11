package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

// TODO: be graceful

var validSet = []struct {
	outData []byte
	wl      *watchList
}{
	{
		outData: []byte(`{
  "restriction": [
    "version",
    "status"
  ],
  "repository": {
    "testdata": {
      "gitdir": "/path/to/git/.git",
      "workdir": "/path/to/work"
    }
  }
}`),
		wl: &watchList{
			Restriction: []string{"version", "status"},
			Map: map[string]repoInfo{
				"testdata": repoInfo{
					Gitdir:  "/path/to/git/.git",
					Workdir: "/path/to/work",
				},
			},
		},
	},
}

func TestWatchList(t *testing.T) {
	t.Run("test isAllow", func(t *testing.T) {
		tests := []struct {
			restricton []string
			firstArg   string
			wantBool   bool
		}{
			{
				restricton: nil,
				firstArg:   "",
				wantBool:   true,
			},
			{
				restricton: []string{},
				firstArg:   "",
				wantBool:   true,
			},
			{
				restricton: []string{"test", "version"},
				firstArg:   "version",
				wantBool:   true,
			},
			{
				restricton: []string{"version"},
				firstArg:   "test",
				wantBool:   false,
			},
		}
		for i, test := range tests {
			wl := &watchList{Restriction: test.restricton}
			if out := wl.isAllow(test.firstArg); out != test.wantBool {
				t.Errorf("t.Errorf [%d]:\n\texp:%+v\n\tout:%+v", i, out, test.wantBool)
			}
		}
	})
}

func TestReadWatchList(t *testing.T) {
	f, err := ioutil.TempFile("", "gits_test_readwatchlist")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	t.Run("check content", func(t *testing.T) {
		tests := validSet
		for i, test := range tests {
			if err := f.Truncate(0); err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(test.outData); err != nil {
				t.Fatal(err)
			}
			out, err := readWatchList(f.Name())
			if err != nil {
				t.Errorf("t.Errorf:%+v", err)
				continue
			}
			if !reflect.DeepEqual(test.wl, out) {
				t.Errorf("t.Errorf [%d]\nexp:%+v\nout:%+v", i, test.wl, out)
				continue
			} else {
				t.Logf("t.Logf out: %+v", out)
			}
		}
	})

	t.Run("invalid filepath", func(t *testing.T) {
		// invalid filepath
		if _, err := readWatchList(""); err == nil {
			t.Fatal("expected error but nil")
		} else {
			t.Logf("t.Logf err: %+v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if err := f.Truncate(0); err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("invalid json"); err != nil {
			t.Fatal(err)
		}
		if _, err := readWatchList(f.Name()); err == nil {
			t.Fatal("expected error but nil")
		} else {
			t.Logf("t.Logf err: %+v", err)
		}
	})
}

func TestWriteWatchList(t *testing.T) {
	f, err := ioutil.TempFile("", "gits_test_writewatchlist")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	t.Run("invalid filepath", func(t *testing.T) {
		wl := &watchList{}

		// case ""
		if err := wl.writeFile(""); err == nil {
			t.Fatal("expected error but nil")
		} else {
			t.Logf("t.Logf err: %+v", err)
		}

		// case "directory"
		dir, err := ioutil.TempDir("", "gits_test_tempdir")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dir)
		if err := wl.writeFile(dir); err == nil {
			t.Fatal("expected error but nil")
		} else {
			t.Logf("t.Logf err: %+v", err)
		}

		// TODO: case "not regular"
	})

	t.Run("check writed content", func(t *testing.T) {
		tests := validSet
		for i, test := range tests {
			if err := f.Truncate(0); err != nil {
				t.Fatal(err)
			}
			if err := test.wl.writeFile(f.Name()); err != nil {
				t.Errorf("t.Errorf [%d]: %+v", i, err)
				continue
			}
			out, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(test.outData, out) {
				t.Errorf("t.Errorf [%d]: exp:%+v\nout:%+v", i, string(test.outData), string(out))
			} else {
				t.Logf("t.Logf [%d]: out:%+v", i, string(out))
			}
		}
	})
}

func TestTemplate(t *testing.T) {
	var s string
	buf := bytes.NewBufferString(s)
	template(buf)
	t.Log(buf)
}
