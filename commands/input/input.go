// Inspired by the git-appraise project

// Package input contains helpers to use a text editor as an input for
// various field of a bug
package input

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-billy/v5/util"

	"github.com/MichaelMure/git-bug/repository"
)

// LaunchEditorWithTemplate will launch an editor as LaunchEditor do, but with a
// provided template.
func LaunchEditorWithTemplate(repo repository.RepoCommonStorage, fileName string, template string) (string, error) {
	err := util.WriteFile(repo.LocalStorage(), fileName, []byte(template), 0644)
	if err != nil {
		return "", err
	}

	return LaunchEditor(repo, fileName)
}

// LaunchEditor launches the default editor configured for the given repo. This
// method blocks until the editor command has returned.
//
// The specified filename should be a temporary file and provided as a relative path
// from the repo (e.g. "FILENAME" will be converted to "[<reporoot>/].git/git-bug/FILENAME"). This file
// will be deleted after the editor is closed and its contents have been read.
//
// This method returns the text that was read from the temporary file, or
// an error if any step in the process failed.
func LaunchEditor(repo repository.RepoCommonStorage, fileName string) (string, error) {
	var err error

	defer func() {
		stError := repo.LocalStorage().Remove(fileName)
		if stError != nil {
			err = errors.Join(err, stError)
		}
	}()

	editor, err := repo.GetCoreEditor()
	if err != nil {
		err = fmt.Errorf("unable to detect default git editor: %w", err)

		return "", err
	}

	repo.LocalStorage().Root()

	// bypass the interface but that's ok: we need that because we are communicating
	// the absolute path to an external program
	path := filepath.Join(repo.LocalStorage().Root(), fileName)
	cmd, err := startInlineCommand(editor, path)
	if err != nil {
		// Running the editor directly did not work. This might mean that
		// the editor string is not a path to an executable, but rather
		// a shell command (e.g. "emacsclient --tty"). As such, we'll try
		// to run the command through bash, and if that fails, try with sh
		args := []string{"-c", fmt.Sprintf("%s %q", editor, path)}
		cmd, err = startInlineCommand("bash", args...)
		if err != nil {
			cmd, err = startInlineCommand("sh", args...)
		}
	}

	if err != nil {
		err = fmt.Errorf("unable to start editor: %w", err)

		return "", err
	}

	if err := cmd.Wait(); err != nil {
		err = fmt.Errorf("editing finished with error: %w", err)

		return "", err
	}

	output, err := os.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("error reading edited file: %w", err)

		return "", err
	}

	return string(output), err
}

// FromFile loads and returns the contents of a given file. If - is passed
// through, much like git, it will read from stdin. This can be piped data,
// unless there is a tty in which case the user will be prompted to enter a
// message.
func FromFile(fileName string) (string, error) {
	if fileName == "-" {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return "", fmt.Errorf("error reading from stdin: %w", err)
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// There is no tty. This will allow us to read piped data instead.
			output, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("error reading from stdin: %w", err)
			}

			return string(output), err
		}

		fmt.Printf("(reading comment from standard input)")
		var output bytes.Buffer
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			output.Write(s.Bytes())
			output.WriteRune('\n')
		}
		return output.String(), nil
	}

	output, err := os.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(output), err
}

func startInlineCommand(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	return cmd, err
}
