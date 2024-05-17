package github_helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	TOKEN_TTL          = int64(5)
	initJwt, initToken sync.Once
	ghAppSignedToken   string
	ghAppToken         *GitHubAppToken = nil
)

func initAppToken(privatePem []byte) {
	initJwt.Do(func() {
		token, err := GenerateToken(privatePem)
		if err != nil {
			panic(err)
		}
		ghAppSignedToken = token
	})
}

func SetGithubAppToken(token *GitHubAppToken) {
	initToken.Do(func() {
		if token == nil {
			panic("GitHub App Token not provided")
		}
		ghAppToken = token
	})
}

// generate jwt token
func GenerateToken(privatePem []byte) (string, error) {
	// parse private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)
	if err != nil {
		return "", fmt.Errorf(`error parsing private key, %v`, err)
	}

	// create the jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": time.Now().Unix(), // Convert int64 to string
		"exp": time.Now().Add(time.Minute * 5).Unix(),
		"iss": 871426, //49337552, //
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

func GetReference(branch string) (GithubRefResponse, error) {
	if ghAppToken == nil {
		panic("GitHub App Token not initialized")
	}
	var respObj GithubRefResponse
	response, err := CallGithubAPI(ghAppToken.Token, "GET", fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo, branch), nil)
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
		panic("GitHub App Token not initialized")
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
		panic("GitHub App Token not initialized")
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

func CreateReference(request GithubRefRequest, branch string) (GithubRefResponse, error) {
	if ghAppToken == nil {
		panic("GitHub App Token not initialized")
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

func UpdateReference(request GithubRefRequest, branch string, createBranch bool) (GithubRefResponse, error) {
	if ghAppToken == nil {
		panic("GitHub App Token not initialized")
	}
	if createBranch {
		_, err := CreateReference(GithubRefRequest{
			Ref: fmt.Sprintf("refs/heads/%s", branch),
			Sha: request.Sha,
		}, branch)
		if err != nil {
			return GithubRefResponse{}, err
		}
	}
	var respObj GithubRefResponse
	response, err := CallGithubAPI(ghAppToken.Token, "PATCH", fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", ghAppToken.Repo.Owner, ghAppToken.Repo.Repo, branch), request)
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
		panic("GitHub App Token not initialized")
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
		return "", fmt.Errorf("error calling github api, status code: %d, response: %s", resp.StatusCode, string(b))
	}

	return string(b), nil
}
