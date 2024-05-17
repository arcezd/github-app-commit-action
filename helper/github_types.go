package github_helper

import "time"

type GitHubRepo struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

type GitHubAppToken struct {
	Repo  GitHubRepo `json:"repo"`
	Token string     `json:"token"`
}

type GitHubUser struct {
	Date  string `json:"date"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Permissions struct {
	Actions  string `json:"actions"`
	Contents string `json:"contents"`
	Metadata string `json:"metadata"`
	Packages string `json:"packages"`
	Checks   string `json:"checks"`
}

type TokenInfo struct {
	Token               string      `json:"token"`
	ExpiresAt           time.Time   `json:"expires_at"`
	Permissions         Permissions `json:"permissions"`
	RepositorySelection string      `json:"repository_selection"`
}

type RefObject struct {
	Sha  string `json:"sha"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

type GithubRefResponse struct {
	Ref    string    `json:"ref"`
	NodeId string    `json:"node_id"`
	Url    string    `json:"url"`
	Object RefObject `json:"object"`
}

type GithubRefRequest struct {
	Ref   string `json:"ref,omitempty"`
	Sha   string `json:"sha"`
	Force bool   `json:"force,omitempty"`
}

type TreeItem struct {
	Path    string  `json:"path"` // file referenced in the tree.
	Mode    string  `json:"mode"` // Can be one of: 100644, 100755, 040000, 160000, 120000
	Type    string  `json:"type"` // Can be one of: blob, tree, or commit.
	Sha     *string `json:"sha"`
	Content string  `json:"content,omitempty"`
	Size    int     `json:"size,omitempty"`
	Url     string  `json:"url,omitempty"`
}

type GithubTreeResponse struct {
	Sha       string     `json:"sha"`
	Url       string     `json:"url"`
	Truncated bool       `json:"truncated"`
	Tree      []TreeItem `json:"tree"`
}

type GithubTreeRequest struct {
	BaseTree string     `json:"base_tree"`
	Tree     []TreeItem `json:"tree"`
}

type CommitParent struct {
	Sha     string `json:"sha"`
	Url     string `json:"url"`
	HtmlUrl string `json:"html_url"`
}

type CommitVerification struct {
	Verified  bool   `json:"verified"`
	Reason    string `json:"reason"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
}

type GithubAccount struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravatarId        string `json:"gravatar_id"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type GithubCommitResponse struct {
	Sha          string             `json:"sha"`
	NodeId       string             `json:"node_id"`
	Url          string             `json:"url"`
	Author       GitHubUser         `json:"author"`
	Committer    GitHubUser         `json:"committer"`
	Message      string             `json:"message"`
	Tree         TreeItem           `json:"tree"`
	Parents      []CommitParent     `json:"parents"`
	Verification CommitVerification `json:"verification"`
	HtmlUrl      string             `json:"html_url"`
}

type GithubCommitRequest struct {
	Message string   `json:"message"`
	Parents []string `json:"parents"`
	Tree    string   `json:"tree"`
}

type GithubBlobResponse struct {
	Url string `json:"url"` // ulr of the blob
	Sha string `json:"sha"` // sha of the blob
}

type GithubBlobRequest struct {
	Content  string `json:"content"`  // The content of the blob.
	Encoding string `json:"encoding"` // The encoding used for content. Currently, "utf-8" and "base64" are supported.
}

type GithubAppInstallationResponse struct {
	Id                     int           `json:"id"`
	Account                GithubAccount `json:"account"`
	RepositorySelection    string        `json:"repository_selection"`
	AccessTokenUrl         string        `json:"access_tokens_url"`
	HtmlUrl                string        `json:"html_url"`
	AppId                  int           `json:"app_id"`
	TargetId               int           `json:"target_id"`
	TargetType             string        `json:"target_type"`
	Permissions            Permissions   `json:"permissions"`
	Events                 []string      `json:"events"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	SingleFileName         string        `json:"single_file_name"`
	HasMultipleSingleFiles bool          `json:"has_multiple_single_files"`
	SingleFilePaths        []string      `json:"single_file_paths"`
	AppSlug                string        `json:"app_slug"`
	SuspendedAt            *time.Time    `json:"suspended_at"`
	SuspendedBy            *string       `json:"suspended_by"`
}
