package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

func ExecGit(gitArgs []string) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("git", gitArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = ioutil.Discard

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() == 1 {
				return "", errors.New("wait status returned non-zero")
			}
		}
		return "", err
	}

	return strings.TrimRight(stdout.String(), "\000\n"), nil
}

func IsInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func OutputFile(path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "unable to read the provided file")
	}
	scanner := bufio.NewScanner(bufio.NewReader(file))
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	return
}

func Union(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		item = strings.ReplaceAll(item, ".sample", "")
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}
