# ghdag

:octocat: ghdag is a tiny workflow engine for GitHub issue and pull request.

## Usage



## Environments

| Environment variable | Description | Default on GitHub Actions |
| --- | --- | --- |
| `GITHUB_TOKEN` | A GitHub access token. | - |
| `GITHUB_REPOSITORY` | The owner and repository name | `owner/repo` of the repository where GitHub Actions are running |
| `GITHUB_API_URL` | The GitHub API URL | `https://api.github.com` |
| `SLACK_API_TOKEN` | A Slack OAuth access token | - |
| `SLACK_WEBHOOK_URL` | A Slack incoming webhook URL | - |
| `SLACK_CHANNEL` | A notify Slack channel | - |
| `GITHUB_ASSIGNEES_SAMPLE` | Number of users to randomly select from those listed in the `assignees:` section. | - |

### Required scope of `SLACK_API_TOKEN`

- `channel:read`
- `chat.write`
- `chat.write.public`
- `users.read`
