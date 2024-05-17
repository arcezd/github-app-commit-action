package github_helper

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func StageModifiedAndNewFiles() error {
	cmd := exec.Command("git", "add", "-A")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

func StageModifiedFiles() error {
	cmd := exec.Command("git", "add", "-u")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
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
		fmt.Println("Error:", err)
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
