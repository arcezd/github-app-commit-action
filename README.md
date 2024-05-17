# Github App Signed Commits Actions
This GH action allow you to sign your commits with a Github App

## Inputs
| Variable | Description | Type |
| -------- | ----------- | ---- |
| `github-app-private-key-file` | The Github App private key filename. | `string` |
| `repository` | **Required**. The private key of the Github App | `string` |
| `branch` | **Required**. The private key of the Github App | `string` |
| `head` | **Required**. The private key of the Github App | `string` |
| `message` | **Required**. The private key of the Github App | `string` |
| `add-new-files` | **Required**. The private key of the Github App | `string` |
| `coauthors` | **Required**. The private key of the Github App | `string` |

## Example usage
```yaml
uses: arcezd/github-app-commit-action@v1
with:
  repository: ${{ github.repository }}
  branch: "new-branch"
  head: "main"
```

## TODO
- [ ] Add support for `GITHUB_APP_INSTALLATION_TOKEN` as input
- [ ] Support executable permissions for uploaded files
- [ ] Fix on-behalf-of commits
- [ ] Add support to specify the list of files to commit
- [ ] Fix commit when renaming files

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
