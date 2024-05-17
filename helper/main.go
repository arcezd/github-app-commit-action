package github_helper

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

type GitHubOrg struct {
	Name  string
	Slug  string
	Email string
}

type GitUser struct {
	Name  string
	Email string
}

type CommitOptions struct {
	AddNewFiles        bool
	RemoveDeletedFiles bool
}

type GitCommit struct {
	HeadBranch *string
	Branch     string
	Coauthors  *[]GitUser
	OnBehalfOf *GitHubOrg
	Message    string
	Options    CommitOptions
}

type GitFile struct {
	FileName   string
	WasDeleted bool
	Sha        *string
}

func UploadFileToGitHubBlob(filename string) (GithubBlobResponse, error) {
	resp := GithubBlobResponse{}
	// check if file exists
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			fmt.Printf("File '%s' was deleted: %s\n", filename, err)
			return resp, err
		} else {
			// other error
			fmt.Printf("Error checking if file '%s' exists: %s\n", filename, err)
			return resp, err
		}
	} else {
		// read file content
		content, err := os.ReadFile(filename)
		if err != nil {
			return resp, err
		}
		base64Content := base64.StdEncoding.EncodeToString(content)
		req := GithubBlobRequest{
			Content:  base64Content,
			Encoding: "base64",
		}

		resp, err = CreateBlob(req)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}
}

func UploadFilesToGitHubBlob(files []string) ([]GitFile, error) {
	//ch := make(chan string, len(files))
	gitFiles := []GitFile{}
	for _, filename := range files {
		//go func() {
		//	res, _ := requester.Get(insrequester.RequestEntity{Endpoint: url})
		//	ch <- fmt.Sprintf("%s: %d", url, res.StatusCode)
		//}()

		// check if file exists
		fileBlobResp, err := UploadFileToGitHubBlob(filename)
		if err != nil {
			if os.IsNotExist(err) {
				gitFiles = append(gitFiles, GitFile{
					FileName:   filename,
					WasDeleted: true,
					Sha:        nil,
				})
			} else {
				return gitFiles, fmt.Errorf("error uploading file '%s' to GitHub: %s", filename, err)
			}
		} else {
			gitFiles = append(gitFiles, GitFile{
				FileName:   filename,
				WasDeleted: false,
				Sha:        &fileBlobResp.Sha,
			})
		}
	}
	return gitFiles, nil
}

func SignJWTAppToken(privatePem []byte) {
	initAppToken(privatePem)
}

func SignJWTAppTokenWithFilename(pemFilename string) {
	if pemFilename == "" {
		panic("PEM file not provided")
	}

	// read the private key from a file
	privatePem, err := os.ReadFile(pemFilename)
	if err != nil {
		panic(err)
	}
	SignJWTAppToken(privatePem)
}

func GenerateInstallationAppToken(repo GitHubRepo) GitHubAppToken {
	// get app installation details
	app, err := GetAppInstallationDetails(ghAppSignedToken, repo)
	if err != nil {
		panic(err)
	}

	// generate installation app token
	installationToken, err := GenerateInstallationAccessToken(ghAppSignedToken, app.Id)
	if err != nil {
		panic(err)
	}
	return GitHubAppToken{
		Repo:  repo,
		Token: installationToken,
	}
}

func CommitAndPush(repo GitHubRepo, commit GitCommit) {
	// get head reference
	if commit.HeadBranch == nil {
		commit.HeadBranch = &commit.Branch
	}
	githubRefResponse, err := GetReference(*commit.HeadBranch)
	if err != nil {
		panic(err)
	}

	// get files to commit
	var files []string
	if commit.Options.AddNewFiles {
		files, err = GetModifiedAndNewFiles()
	} else {
		files, err = GetModifiedFiles()
	}
	if err != nil {
		panic(err)
	}

	// upload files to github blobs
	gitFiles, err := UploadFilesToGitHubBlob(files)
	if err != nil {
		panic(err)
	}

	// create git tree
	treeFiles := []TreeItem{}
	for _, file := range gitFiles {
		treeFiles = append(treeFiles, TreeItem{
			Path: file.FileName,
			Mode: "100644",
			Type: "blob",
			Sha:  file.Sha,
		})
	}
	treeReq := GithubTreeRequest{
		BaseTree: githubRefResponse.Object.Sha,
		Tree:     treeFiles,
	}
	b, err := json.Marshal(treeReq)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	treeResp, err := CreateTree(treeReq)
	if err != nil {
		panic(err)
	}

	// add coauthors and on-behalf-of to commit message
	commitMessage := commit.Message
	if commit.Coauthors != nil || commit.OnBehalfOf != nil {
		commitMessage = fmt.Sprintf("%s\n\n", commit.Message)
		if commit.Coauthors != nil {
			for _, coauthor := range *commit.Coauthors {
				commitMessage = fmt.Sprintf("%sCo-authored-by: %s <%s>\n", commitMessage, coauthor.Name, coauthor.Email)
			}
		}
		if commit.OnBehalfOf != nil {
			commitMessage = fmt.Sprintf("%son-behalf-of: @%s <%s>", commitMessage, commit.OnBehalfOf.Slug, commit.OnBehalfOf.Email)
		}
	}

	// create commit
	commitReq := GithubCommitRequest{
		Message: commitMessage,
		Tree:    treeResp.Sha,
		Parents: []string{
			githubRefResponse.Object.Sha,
		},
	}
	commitResp, err := CreateCommit(commitReq)
	if err != nil {
		panic(err)
	}

	// update git reference
	refReq := GithubRefRequest{
		Sha: commitResp.Sha,
	}
	refResp, err := UpdateReference(refReq, commit.Branch, false)
	if err != nil {
		panic(err)
	}

	message := fmt.Sprintf("Commit '%s' pushed to branch '%s' with SHA '%s'\n", commit.Message, commit.Branch, refResp.Object.Sha)
	fmt.Print(message)
	AppendToGHActionsSummary(message)

	SendToGHActionsOutput("sha", refResp.Object.Sha)
}
