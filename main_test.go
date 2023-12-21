package main

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func Test_ExecuteCommand(t *testing.T) {
	cmd := newFetchCommitsCmd("", "", 1,
		"desc", false, "https://api.github.com", "console", "",
		"")
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--github-token", "", "--username", "", "--date-range", "1"})
	cmd.Execute()
	_, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

}
