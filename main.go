package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	gh "github.com/arcezd/github-app-commit-action/helper"
)

var (
	BuildVersion = "v0.0.1"
)

const (
	// github pem env var
	githubAppPrivateKeyEnvVar = "GITHUB_APP_PRIVATE_KEY"
	defaultCommitMessage      = "chore: autopublish ${date}"
)

func main() {
	var branch, headBranch, repository, privateKeyPemFilename, commitMsg, coauthors string
	var version, help, addNewFiles bool

	// parse flags
	flag.BoolVar(&help, "help", false, "CLI help")
	flag.BoolVar(&version, "version", false, "Version of the CLI")
	flag.StringVar(&headBranch, "h", "", "GitHub head branch to commit from. Default is the same as branch")
	flag.StringVar(&branch, "b", "main", "GitHub target branch to commit to")
	flag.StringVar(&repository, "r", "", "GitHub repository in the format owner/repo")
	flag.StringVar(&privateKeyPemFilename, "p", "", fmt.Sprintf("Path to the private key pem file. %s env variable has priority over this", githubAppPrivateKeyEnvVar))
	flag.StringVar(&commitMsg, "m", defaultCommitMessage, "Commit message")
	flag.StringVar(&coauthors, "c", "", "Coauthors in the format 'Name1 <email1>, Name2 <email2>'")
	flag.BoolVar(&addNewFiles, "a", true, "Add new files to the commit")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}
	if version {
		fmt.Println(BuildVersion)
		return
	}

	// parse repository to get owner and repo
	repo := gh.GitHubRepo{}
	if repository != "" {
		// validate repository format with regex and extract owner and repo
		validRepoPattern := `^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)$`
		re := regexp.MustCompile(validRepoPattern)
		matches := re.FindStringSubmatch(repository)
		if matches == nil {
			// return error, because the input is not in the expected format
			panic(fmt.Errorf("invalid repository format '%s', expected format is 'owner/repo'", repository))
		}
		// if valid, assign owner and repo
		repo.Owner = matches[1]
		repo.Repo = matches[2]
		fmt.Printf("Owner: %s, Repo: %s\n", repo.Owner, repo.Repo)
	} else {
		// return error, because the repository is required
		flag.PrintDefaults()
		panic(fmt.Errorf("repository flag is required. Use -r flag to specify the repository in the format owner/repo"))
	}

	if headBranch == "" {
		headBranch = branch
	}

	// sign the JWT token with the private key
	privateKeyPemString := os.Getenv(githubAppPrivateKeyEnvVar)
	if privateKeyPemString != "" {
		// validate that the format for the private key is correct
		block, _ := pem.Decode([]byte(privateKeyPemString))
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			log.Fatal("failed to decode PEM block containing private key")
		}
		// sign the JWT token with the private key from the env var
		gh.SignJWTAppToken([]byte(privateKeyPemString))
	} else if privateKeyPemFilename != "" {
		// sign the JWT token with the private key from the filename
		gh.SignJWTAppTokenWithFilename(privateKeyPemFilename)
	} else {
		panic("You need to provide a private key in the environment variable GITHUB_APP_PRIVATE_KEY or a filename with the -p flag")
	}

	// parse coauthors
	coauthorsParam := []gh.GitUser{}
	if coauthors != "" {
		// slice the coauthors string by comma
		coauthorsSlice := strings.Split(coauthors, ",")
		for _, coauthor := range coauthorsSlice {
			validCoauthorPattern := `^(.+) <(.+)>$`
			re := regexp.MustCompile(validCoauthorPattern)
			matches := re.FindStringSubmatch(coauthor)
			if matches == nil {
				log.Fatalf("invalid coauthor format '%s', expected format is 'Name <email@example.com>'", coauthor)
			}
			// if valid, assign name and email
			name := matches[1]
			email := matches[2]
			coauthorsParam = append(coauthorsParam, gh.GitUser{
				Name:  name,
				Email: email,
			})
		}
	}

	// set commit message if not defined
	dt := time.Now()
	if commitMsg == defaultCommitMessage {
		commitMsg = fmt.Sprintf("chore: autopublish %s", dt.Format(time.RFC3339))
	}

	token := gh.GenerateInstallationAppToken(repo)
	gh.SetGithubAppToken(&token)
	//onBehalfOf := gh.GitHubOrg{
	//	Name:  "BitsCR",
	//	Slug:  "bits-cr",
	//	Email: "dev@bits.cr",
	//}
	gh.CommitAndPush(
		repo,
		gh.GitCommit{
			Branch:     branch,
			HeadBranch: &headBranch,
			Message:    commitMsg,
			Coauthors:  &coauthorsParam,
			//OnBehalfOf: &onBehalfOf,
			Options: gh.CommitOptions{
				AddNewFiles: addNewFiles,
			},
		},
	)
}
