package github_helper

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func ListFiles(dir *string) error {
	cmd := exec.Command("ls", "-al")
	if dir != nil {
		cmd.Dir = *dir
	}
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error during 'ListFiles' - Path %s - %s, :%s\n", cmd.Dir, cmd.Stderr, err)
		return err
	}
	fmt.Printf("ListFiles - Path %s - :%s\n", cmd.Dir, string(output))
	return nil
}

func StageModifiedAndNewFiles() error {
	cmd := exec.Command("git", "add", "-A")

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
		panic(err)
	}

	fmt.Printf("Output: %s\n", output) // TODO: remove
	return nil
}

func StageModifiedFiles() error {
	cmd := exec.Command("git", "add", "-u")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error during 'StageModifiedFiles' - Path %s - :%s\n", cmd.Dir, err)
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
	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error during 'GetModifiedFilesFromGitDiff' - Path %s - :%s\n", cmd.Dir, err)
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

func writeToGHActionsVar(name string, value string) {
	if name != "" {
		f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0644)
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
