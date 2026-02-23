package github_helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	TOKEN_TTL          = int64(5)
	initJwt, initToken sync.Once
	ghAppSignedToken   string
	ghAppToken         *GitHubAppToken = nil
	initJwtErr         error
)

func initAppToken(appId string, privatePem []byte) error {
	initJwt.Do(func() {
		token, err := GenerateToken(appId, privatePem)
		if err != nil {
			initJwtErr = err
			return
		}
		ghAppSignedToken = token
	})

	return initJwtErr
}

func SetGithubAppToken(token *GitHubAppToken) error {
	if token == nil {
		return fmt.Errorf("GitHub App Token not provided")
	}

	initToken.Do(func() {
		ghAppToken = token
	})

	return nil
}

// generate jwt token
func GenerateToken(githubAppId string, privatePem []byte) (string, error) {
	// parse private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)
	if err != nil {
		return "", fmt.Errorf(`error parsing private key, %v`, err)
	}

	// create the jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix(), // Convert int64 to string
		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"iss": githubAppId, //871426, //49337552, //
	})

	// sign the token using the private key
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf(`error signing jwt token, %v`, err)
	}

	// sign and get the complete encoded token as a string using the secret
	return tokenString, nil
}

func GenerateInstallationAccessToken(token string, installationId int) (string, error) {
	response, err := CallGithubAPI(token, "POST", fmt.Sprintf("/app/installations/%d/access_tokens", installationId), nil)
	if err != nil {
		return "", err
	}

	// parse the response
	var tokenInfo TokenInfo
	err = json.Unmarshal([]byte(response), &tokenInfo)
	if err != nil {
		return "", err
	}
	return tokenInfo.Token, nil
}

func GetReference(ref string) (GithubRefResponse, error) {
	if ghAppToken == nil {
		return GithubRefResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	var respObj GithubRefResponse
	response, err := CallGithubAPI(ghAppToken.Token, "GET", fmt.Sprintf("/repos/%s/%s/git/%s", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo, ref), nil)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func CreateTree(tree GithubTreeRequest) (GithubTreeResponse, error) {
	if ghAppToken == nil {
		return GithubTreeResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	githubTreeResponse := GithubTreeResponse{}
	response, err := CallGithubAPI(ghAppToken.Token, "POST", fmt.Sprintf("/repos/%s/%s/git/trees", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo), tree)
	if err != nil {
		return githubTreeResponse, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &githubTreeResponse)
	if err != nil {
		return githubTreeResponse, err
	}
	return githubTreeResponse, nil
}

func CreateCommit(commit GithubCommitRequest) (GithubCommitResponse, error) {
	if ghAppToken == nil {
		return GithubCommitResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	var respObj GithubCommitResponse
	response, err := CallGithubAPI(ghAppToken.Token, "POST", fmt.Sprintf("/repos/%s/%s/git/commits", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo), commit)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func CreateReference(request GithubRefRequest) (GithubRefResponse, error) {
	if ghAppToken == nil {
		return GithubRefResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	var respObj GithubRefResponse
	response, err := CallGithubAPI(ghAppToken.Token, "POST", fmt.Sprintf("/repos/%s/%s/git/refs", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo), request)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func UpdateReference(request GithubRefRequest, ref string, createBranch bool) (GithubRefResponse, error) {
	if ghAppToken == nil {
		return GithubRefResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	if createBranch {
		// Remove "heads/" prefix if present, then add "refs/heads/"
		branchName := ref
		if strings.HasPrefix(ref, "heads/") {
			branchName = ref[6:] // Remove "heads/" prefix
		}
		respObj, err := CreateReference(GithubRefRequest{
			Ref: fmt.Sprintf("refs/heads/%s", branchName),
			Sha: request.Sha,
		})
		if err != nil {
			return GithubRefResponse{}, err
		}
		// Return early when creating branch - no need to update it
		return respObj, nil
	}
	var respObj GithubRefResponse
	response, err := CallGithubAPI(ghAppToken.Token, "PATCH", fmt.Sprintf("/repos/%s/%s/git/refs/%s", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo, ref), request)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func CreateBlob(blob GithubBlobRequest) (GithubBlobResponse, error) {
	if ghAppToken == nil {
		return GithubBlobResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	var respObj GithubBlobResponse
	response, err := CallGithubAPI(ghAppToken.Token, "POST", fmt.Sprintf("/repos/%s/%s/git/blobs", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo), blob)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func CreateTag(tag GithubTagRequest) (GithubTagResponse, error) {
	if ghAppToken == nil {
		return GithubTagResponse{}, fmt.Errorf("GitHub App Token not initialized")
	}
	var respObj GithubTagResponse
	response, err := CallGithubAPI(ghAppToken.Token, "POST", fmt.Sprintf("/repos/%s/%s/git/tags", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo), tag)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func GetAppInstallationDetails(jwt string, repo GitHubRepo) (GithubAppInstallationResponse, error) {
	var respObj GithubAppInstallationResponse
	response, err := CallGithubAPI(jwt, "GET", fmt.Sprintf("/repos/%s/%s/installation", repo.Owner, repo.Repo), nil)
	if err != nil {
		return respObj, err
	}

	// parse the response
	err = json.Unmarshal([]byte(response), &respObj)
	if err != nil {
		return respObj, err
	}
	return respObj, nil
}

func CallGithubAPI(token string, method string, path string, data interface{}) (string, error) {
	// define the request
	req, err := http.NewRequest(method, fmt.Sprintf("https://api.github.com%s", path), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// set body
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		req.Body = io.NopCloser(bytes.NewReader(b))
	}

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// close the response body
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// check http status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", &GitHubAPIError{StatusCode: resp.StatusCode, Body: string(b)}
	}

	return string(b), nil
}
