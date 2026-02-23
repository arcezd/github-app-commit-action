package github_helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type FileInfo struct {
	Permissions string
	Number      string
	Owner       string
	Group       string
	Size        string
	Month       string
	Day         string
	Time        string
	Name        string
}

func ListFiles(dir *string) ([]FileInfo, error) {
	output, err := executeCommand("ls", "-al")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var files []FileInfo

	for _, line := range lines[1:] {
		re := regexp.MustCompile(`(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.+)`)
		match := re.FindStringSubmatch(line)

		if len(match) == 0 {
			continue
		}

		files = append(files, FileInfo{
			Permissions: match[1],
			Number:      match[2],
			Owner:       match[3],
			Group:       match[4],
			Size:        match[5],
			Month:       match[6],
			Day:         match[7],
			Time:        match[8],
			Name:        match[9],
		})
	}

	return files, nil
}

func StageModifiedAndNewFiles() error {
	_, err := executeCommand("git", "add", "-A")
	if err != nil {
		return err
	}

	return nil
}

func StageModifiedFiles() error {
	_, err := executeCommand("git", "add", "-u")
	if err != nil {
		return err
	}

	return nil
}

func GetModifiedAndNewFiles() ([]string, error) {
	var modfiles []string

	err := StageModifiedAndNewFiles()
	if err != nil {
		return modfiles, err
	}

	return GetModifiedFilesFromGitDiff(true)
}

func GetModifiedFiles() ([]string, error) {
	var modfiles []string

	err := StageModifiedFiles()
	if err != nil {
		return modfiles, err
	}

	return GetModifiedFilesFromGitDiff(true)
}

func GetModifiedFilesFromGitDiff(fromStaged bool) ([]string, error) {
	var modfiles []string

	cmdArgs := []string{"diff", "--name-only"}
	if fromStaged {
		cmdArgs = append(cmdArgs, "--cached")
	}

	output, err := executeCommand("git", cmdArgs...)
	if err != nil {
		fmt.Printf("Error during 'GetModifiedFilesFromGitDiff': %s\n", err)
		return modfiles, err
	}

	files := strings.Split(string(output), "\n")
	for _, file := range files {
		if file != "" {
			modfiles = append(modfiles, file)
		}
	}

	return modfiles, nil
}

func AppendToGHActionsSummary(summary string) {
	if IsGitHubActions() {
		summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
		writeToGHActionsVar(summaryPath, summary)
	}
}

func SendToGHActionsOutput(variable string, output string) {
	if IsGitHubActions() {
		writeToGHActionsVar(os.Getenv("GITHUB_OUTPUT"), fmt.Sprintf("%s=%s", variable, output))
	}
}

var execCommandFunc = executeCommandDefault

func executeCommand(command string, args ...string) ([]byte, error) {
	return execCommandFunc(command, args)
}

func executeCommandDefault(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("-----------------------------\n")
		fmt.Printf("Error during 'StageModifiedAndNewFiles'\n")
		fmt.Printf("Path: %s\n", cmd.Dir)
		fmt.Printf("Stderr: %s\n", stderr.String())
		fmt.Printf("Error: %s\n", err)
		fmt.Printf("-----------------------------\n")

		return nil, err
	}

	return output, nil
}

func writeToGHActionsVar(name string, value string) {
	if name != "" {
		f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0644) //nolint:gosec // name comes from GitHub Actions env vars
		if err != nil {
			log.Fatalf("failed opening file: %s", err)
		}
		defer f.Close()

		_, err = fmt.Fprintln(f, value)
		if err != nil {
			log.Fatalf("failed writing to file: %s", err)
		}
	}
}

func IsGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}
