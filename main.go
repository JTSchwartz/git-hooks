package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func main() {
	var textEditor string
	var shouldUseEmacs bool
	var shouldUseNano bool
	var shouldOutput bool
	var shouldInput bool
	var replace bool

	app := &cli.App{
		Name:  "git-hooks",
		Usage: "Manipulate git hooks",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "emacs",
				Aliases:     []string{"e"},
				Usage:       "Use Emacs text editor",
				Value:       false,
				Destination: &shouldUseEmacs,
			},
			&cli.BoolFlag{
				Name:        "nano",
				Aliases:     []string{"n"},
				Usage:       "Use Nano text editor",
				Value:       false,
				Destination: &shouldUseNano,
			},
			&cli.StringFlag{
				Name:        "text-editor",
				Aliases:     []string{"t"},
				Usage:       "Use Nano text editor",
				Value:       "vim",
				Destination: &textEditor,
			},
			&cli.BoolFlag{
				Name:        "input",
				Aliases:     []string{"i"},
				Usage:       "Append content from pipe or arguments to specified hook file",
				Value:       false,
				Destination: &shouldInput,
			},
			&cli.BoolFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Usage:       "Display content of specified hook file",
				Value:       false,
				Destination: &shouldOutput,
			},
			&cli.BoolFlag{
				Name:        "replace",
				Aliases:     []string{"r"},
				Usage:       "Replace contents of file rather than append (Can only be used alongside --input)",
				Value:       false,
				Destination: &replace,
			},
		},
		Action: func(c *cli.Context) (err error) {
			hooksPath, err := getHooksPath()
			hook := c.Args().Get(0)

			if len(hook) == 0 {
				recognized := []string{"applypatch-msg", "fsmonitor-watchman", "pre-applypatch", "pre-merge-commit", "pre-push", "pre-receive", "update", "commit-msg", "post-update", "pre-commit", "prepare-commit-msg", "pre-rebase", "push-to-checkout"}
				found, e := getExistingHooks(hooksPath)
				if e != nil {
					return e
				}
				fmt.Printf("Recognized hooks:\n%v\n", Union(recognized, found))
				return
			}

			selectedHookPath := hooksPath + "/" + hook
			if err != nil {
				return
			} else if _, err = os.Stat(selectedHookPath); errors.Is(err, os.ErrNotExist) {
				selectedHooPath := selectedHookPath + ".sample"
				if _, err = os.Stat(selectedHooPath); errors.Is(err, os.ErrNotExist) {
					fmt.Printf("Hook not recoginized, generating new hook: %s\n", hook)
				} else {
					err = os.Rename(selectedHooPath, selectedHookPath)
					if err != nil {
						return
					}
				}
			}

			if shouldInput {
				var instructions []string
				if IsInputFromPipe() {
					scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
					for scanner.Scan() {
						instructions = append(instructions, scanner.Text())
					}
				} else {
					instructions = c.Args().Tail()
				}

				f, e := os.OpenFile(selectedHookPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if e != nil {
					return e
				}
				defer f.Close()

				if replace {
					err = f.Truncate(0)
					_, err = f.Seek(0, 0)
				}

				for _, line := range instructions {
					_, err = f.WriteString(line + "\n")
					if err != nil {
						return err
					}
				}
			} else {
				editor := getEditor(shouldUseEmacs, shouldUseNano, textEditor)
				err = openEditor(editor, selectedHookPath)
			}
			if err != nil {
				return
			}

			if shouldOutput {
				err = OutputFile(selectedHookPath)
			}
			return
		},
	}

	if e := app.Run(os.Args); e != nil {
		log.Fatal(e)
	}
}

func openEditor(editor string, selectedHookPath string) (err error) {
	cmd := exec.Command(editor, selectedHookPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getEditor(emacs bool, nano bool, defaultEditor string) string {
	if emacs {
		return "emacs"
	} else if nano {
		return "nano"
	}
	return defaultEditor
}

func getHooksPath() (path string, err error) {
	path, err = ExecGit([]string{"config", "--get", "--null", "core.hooksPath"})
	if err != nil {
		path, err = getGitRoot()
		path += "/.git/hooks"
	}
	return
}

func getGitRoot() (string, error) {
	return ExecGit([]string{"rev-parse", "--show-toplevel"})
}

func getExistingHooks(path string) (existing []string, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			existing = append(existing, file.Name())
		}
	}
	return
}
