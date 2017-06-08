package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestReadWatchList(t *testing.T) {
	f, err := ioutil.TempFile("", "gits_test_readwatchlist")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	t.Run("check content", func(t *testing.T) {
		tests := []struct {
			writeData string
			expected  watchList
		}{
			{
				writeData: `{
  "testdata": {
    "gitdir": "/path/to/git/.git",
    "workdir": "/path/to/work",
    "readonly": false
  }
}`,
				expected: map[string]repoInfo{
					"testdata": repoInfo{
						Gitdir:   "/path/to/git/.git",
						Workdir:  "/path/to/work",
						Readonly: false,
					},
				},
			},
		}
		for i, test := range tests {
			if err := f.Truncate(0); err != nil {
				t.Fatal(err)
			}
			if _, err := f.WriteString(test.writeData); err != nil {
				t.Fatal(err)
			}
			out, err := readWatchList(f.Name())
			if err != nil {
				t.Errorf("t.Errorf:%+v", err)
				continue
			}
			if !reflect.DeepEqual(test.expected, out) {
				t.Errorf("t.Errorf [%d]\nexp:%+v\nout:%+v", i, test.expected, out)
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
		if err := writeWatchList(watchList{}, ""); err == nil {
			t.Fatal("expected error but nil")
		} else {
			t.Logf("t.Logf err: %+v", err)
		}
	})

	// TODO: consider case nil
	//t.Run("invalid json marshal", func(t *testing.T) {
	//	if err := writeWatchList(nil, f.Name()); err == nil {
	//		t.Fatal("expected error but nil")
	//	} else {
	//		t.Logf("t.Logf err: %+v", err)
	//	}
	//})

	t.Run("check writed content", func(t *testing.T) {
		tests := []struct {
			writeData watchList
			expected  []byte
		}{
			{
				writeData: map[string]repoInfo{
					"testdata": repoInfo{
						Gitdir:   "/path/to/git/.git",
						Workdir:  "/path/to/work",
						Readonly: false,
					},
				},
				expected: []byte(`{
  "testdata": {
    "gitdir": "/path/to/git/.git",
    "workdir": "/path/to/work",
    "readonly": false
  }
}`),
			},
		}
		for i, test := range tests {
			if err := f.Truncate(0); err != nil {
				t.Fatal(err)
			}
			if err := writeWatchList(test.writeData, f.Name()); err != nil {
				t.Errorf("t.Errorf [%d]: %+v", i, err)
				continue
			}
			out, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(test.expected, out) {
				t.Errorf("t.Errorf [%d]: exp:%+v\nout:%+v", i, string(test.expected), string(out))
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
