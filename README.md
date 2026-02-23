# Github App Signed Commits Actions
This GH action allow you to sign your commits with a Github App

## Inputs
| Variable | Description | Type | Required | Default |
| -------- | ----------- | ---- | -------- | ------- |
| `github-app-id` | The Github App ID. | `string` | Yes | - |
| `github-app-private-key` | The Github App private key in PEM format. Can be plain PEM or base64-encoded PEM. | `string` | No* | - |
| `github-app-private-key-file` | Path to the Github App private key PEM file. | `string` | No* | - |
| `repository` | GitHub repository in the format owner/repo | `string` | Yes | - |
| `branch` | Target branch to commit to | `string` | Yes | `main` |
| `head` | Head branch to commit from | `string` | No | Same as `branch` |
| `message` | Commit message | `string` | No | `chore: autopublish ${date}` |
| `force-push` | Force push to the branch | `bool` | No | `false` |
| `tags` | Comma-separated tags to create for the commit | `string` | No | - |
| `add-new-files` | Add new files to the commit | `bool` | No | `true` |
| `coauthors` | Coauthors in the format 'Name1 <email1>, Name2 <email2>' | `string` | No | - |

\* Either `github-app-private-key` or `github-app-private-key-file` must be provided.

### Private Key Format

The `github-app-private-key` input accepts the private key in two formats:
- **Plain PEM**: Standard PEM format (multiline string beginning with `-----BEGIN RSA PRIVATE KEY-----`)
- **Base64-encoded PEM**: The entire PEM content encoded as base64 (useful in CI/CD environments where multiline secrets are difficult to handle)

The action will automatically detect and decode base64-encoded keys.

## Example usage

### Using private key from secrets
```yaml
uses: arcezd/github-app-commit-action@v1
with:
  github-app-id: ${{ secrets.GH_APP_ID }}
  github-app-private-key: ${{ secrets.GH_APP_PRIVATE_KEY }}
  repository: ${{ github.repository }}
  branch: "new-branch"
  head: "main"
  message: "feat: automated commit"
```

### Using private key file
```yaml
uses: arcezd/github-app-commit-action@v1
with:
  github-app-id: ${{ secrets.GH_APP_ID }}
  github-app-private-key-file: "./path/to/private-key.pem"
  repository: ${{ github.repository }}
  branch: "new-branch"
```

### Creating tags
```yaml
uses: arcezd/github-app-commit-action@v1
with:
  github-app-id: ${{ secrets.GH_APP_ID }}
  github-app-private-key: ${{ secrets.GH_APP_PRIVATE_KEY }}
  repository: ${{ github.repository }}
  branch: "main"
  tags: "v1.0.0, production"
  message: "release: version 1.0.0"
```

## TODO
- [ ] Support executable permissions for uploaded files
- [ ] Fix on-behalf-of commits
- [ ] Add support to specify the list of files to commit
- [ ] Fix commit when renaming files

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
