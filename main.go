package main

import (
	"encoding/base64"
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

const (
	// github pem env var
	githubAppPrivateKeyEnvVar = "GH_APP_PRIVATE_KEY"
	defaultCommitMessage      = "chore: autopublish ${date}"
)

func run(args []string, getenv func(string) string) error {
	var appId, branch, headBranch, repository, privateKeyPemFilename, commitMsg, coauthors, tags string
	var version, help, force, addNewFiles bool

	// parse flags
	fs := flag.NewFlagSet("github-app-commit-action", flag.ContinueOnError)
	fs.BoolVar(&help, "help", false, "CLI help")
	fs.BoolVar(&version, "version", false, "Version of the CLI")
	fs.StringVar(&appId, "i", "", "GitHub app id")
	fs.StringVar(&headBranch, "h", "", "GitHub head branch to commit from. Default is the same as branch")
	fs.StringVar(&branch, "b", "main", "GitHub target branch to commit to")
	fs.StringVar(&repository, "r", "", "GitHub repository in the format owner/repo")
	fs.StringVar(&privateKeyPemFilename, "p", "", fmt.Sprintf("Path to the private key pem file. %s env variable has priority over this", githubAppPrivateKeyEnvVar))
	fs.StringVar(&commitMsg, "m", defaultCommitMessage, "Commit message")
	fs.StringVar(&coauthors, "c", "", "Coauthors in the format 'Name1 <email1>, Name2 <email2>'")
	fs.StringVar(&tags, "t", "", "Tags separated by commass, 'tag1, tag2, tag3'")
	fs.BoolVar(&addNewFiles, "a", true, "Add new files to the commit")
	fs.BoolVar(&force, "f", false, "Force push to the branch")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if help {
		fs.PrintDefaults()
		return nil
	}

	if version {
		fmt.Println(BuildVersion)
		return nil
	}

	// parse repository to get owner and repo
	repo := gh.GitHubRepo{}

	if repository != "" {
		// validate repository format with regex and extract owner and repo
		validRepoPattern := `^([a-zA-Z0-9_-]+)/([a-zA-Z0-9_-]+)$`
		re := regexp.MustCompile(validRepoPattern)
		matches := re.FindStringSubmatch(repository)

		if matches == nil {
			return fmt.Errorf("invalid repository format '%s', expected format is 'owner/repo'", repository)
		}
		// if valid, assign owner and repo
		repo.Owner = matches[1]
		repo.Repo = matches[2]
		fmt.Printf("Owner: %s, Repo: %s\n", repo.Owner, repo.Repo)
	} else {
		return fmt.Errorf("repository flag is required. Use -r flag to specify the repository in the format owner/repo")
	}

	if headBranch == "" {
		headBranch = branch
	}

	// sign the JWT token with the private key
	if appId == "" {
		return fmt.Errorf("GitHub app id is required. Use -i flag to specify the GitHub app id")
	}

	privateKeyPemString := getenv(githubAppPrivateKeyEnvVar)

	if privateKeyPemString != "" {
		privateKeyPemBytes := []byte(privateKeyPemString)

		// try base64 decoding in case the key was passed encoded
		decoded, err := base64.StdEncoding.DecodeString(privateKeyPemString)
		if err == nil {
			// check if the decoded content is valid PEM
			if block, _ := pem.Decode(decoded); block != nil {
				privateKeyPemBytes = decoded
			}
		}

		// validate that the format for the private key is correct
		block, _ := pem.Decode(privateKeyPemBytes)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return fmt.Errorf("failed to decode PEM block containing private key")
		}
		// sign the JWT token with the private key from the env var
		if err := gh.SignJWTAppToken(appId, privateKeyPemBytes); err != nil {
			return fmt.Errorf("failed to sign JWT token: %w", err)
		}
	} else if privateKeyPemFilename != "" {
		// sign the JWT token with the private key from the filename
		if err := gh.SignJWTAppTokenWithFilename(appId, privateKeyPemFilename); err != nil {
			return fmt.Errorf("failed to sign JWT token from file: %w", err)
		}
	} else {
		return fmt.Errorf("you need to provide a private key in the environment variable %s or a filename with the -p flag", githubAppPrivateKeyEnvVar)
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
				return fmt.Errorf("invalid coauthor format '%s', expected format is 'Name <email@example.com>'", coauthor)
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

	token, err := gh.GenerateInstallationAppToken(repo)
	if err != nil {
		return fmt.Errorf("failed to generate installation token: %w", err)
	}

	if err := gh.SetGithubAppToken(&token); err != nil {
		return fmt.Errorf("failed to set GitHub App token: %w", err)
	}

	commitSha, err := gh.CommitAndPush(
		repo,
		gh.GitCommit{
			Branch:     branch,
			HeadBranch: &headBranch,
			Message:    commitMsg,
			Coauthors:  &coauthorsParam,
			Options: gh.CommitOptions{
				AddNewFiles: addNewFiles,
				Force:       force,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to commit and push: %w", err)
	}

	if tags != "" {
		// split tags by comma
		tagsList := strings.Split(tags, ",")
		// create tags
		for _, tag := range tagsList {
			// remove leading and trailing spaces
			tag = strings.TrimSpace(tag)

			if err := gh.CreateTagAndPush(gh.GitTag{
				TagName:   tag,
				Message:   commitMsg,
				CommitSha: commitSha,
			}); err != nil {
				return fmt.Errorf("failed to create tag '%s': %w", tag, err)
			}
		}
	}

	return nil
}

func main() {
	if err := run(os.Args[1:], os.Getenv); err != nil {
		log.Fatal(err)
	}
}
