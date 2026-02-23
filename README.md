# Github App Signed Commits Actions
This GH action allow you to sign your commits with a Github App

## Inputs
| Variable | Description | Type |
| -------- | ----------- | ---- |
| `github-app-id` | **Required**. The Github App ID. | `string` |
| `github-app-private-key-file` | The Github App private key filename. | `string` |
| `repository` | **Required**. GitHub repository in the format owner/repo | `string` |
| `branch` | Target branch to commit to (default "main")| `string` |
| `head` | head branch to commit from. Default is the same as branch | `string` |
| `message` | Commit message (default "chore: autopublish ${date}") | `string` |
| `add-new-files` | Add new files to the commit. (default true) | `bool` |
| `coauthors` | Coauthors in the format 'Name1 <email1>, Name2 <email2>' | `string` |

## Example usage
```yaml
uses: arcezd/github-app-commit-action@v1
with:
  repository: ${{ github.repository }}
  branch: "new-branch"
  head: "main"
```

## TODO
- [ ] Support executable permissions for uploaded files
- [ ] Fix on-behalf-of commits
- [ ] Add support to specify the list of files to commit
- [ ] Fix commit when renaming files

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
