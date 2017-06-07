package main

import (
	"bytes"
	"testing"
)

func TestTemplate(t *testing.T) {
	var s string
	buf := bytes.NewBufferString(s)
	template(buf)
	t.Log(buf)
}
