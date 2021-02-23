# ghdag

:octocat: ghdag is a tiny workflow engine for GitHub issue and pull request.

## Getting Started

### Generate a workflow file

``` console
$ ghdag init myworkflow
2021-02-23T23:29:48+09:00 [INFO] ghdag version 0.0.1
2021-02-23T23:29:48+09:00 [INFO] Creating myworkflow.yml
```

### Run workflow on your machine

``` console
$ env GITHUB_TOKEN=xxXxXXxxXXxx \
GITHUB_REPOGITORY=owner/repo \
SLACK_API_TOKEN=xoxb-xXXxxXXXxxXXXxXXXxxxxxxx \
ghdag run myworkflow.yml
```

### Run workflow on GitHub Actions

:construction:

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

## Install

**deb:**

Use [dpkg-i-from-url](https://github.com/k1LoW/dpkg-i-from-url)

``` console
$ export GHDAG_VERSION=X.X.X
$ curl -L https://git.io/dpkg-i-from-url | bash -s -- https://github.com/k1LoW/ghdag/releases/download/v$GHDAG_VERSION/ghdag_$GHDAG_VERSION-1_amd64.deb
```

**RPM:**

``` console
$ export GHDAG_VERSION=X.X.X
$ yum install https://github.com/k1LoW/ghdag/releases/download/v$GHDAG_VERSION/ghdag_$GHDAG_VERSION-1_amd64.rpm
```

**homebrew tap:**

```console
$ brew install k1LoW/tap/ghdag
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/ghdag/releases)

**go get:**

```console
$ go get github.com/k1LoW/ghdag
```
