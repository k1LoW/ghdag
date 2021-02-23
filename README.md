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

## Workflow configuration

### Syntax

#### Tasks (`tasks:`)

#### Id (`tasks[*].id:`)

#### Name (`tasks[*].name:`)

#### Run condition (`tasks[*].if:`)

##### Expression syntax

ghdag use [antonmedv/expr](https://github.com/antonmedv/expr).

See [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).

##### Variables

#### Environment variables (`tasks[*].env:`)

#### Actions (`tasks[*].do:`, `tasks[*].ok:`, `tasks[*].ng:`)

#### Execute command (`tasks[*].<action_type>.run:`)

#### Update labels (`tasks[*].<action_type>.labels:`)

#### Update assignees (`tasks[*].<action_type>.assignees:`)

#### Add comment (`tasks[*].<action_type>.comment:`)

#### Change state (`tasks[*].<action_type>.state:`)

#### Notify message using Slack (`tasks[*].<action_type>.notify:`)

#### Call next tasks (`tasks[*].<action_type>.next:`)

### Environment variables

| Environment variable | Nameription | Default on GitHub Actions |
| --- | --- | --- |
| `GITHUB_TOKEN` | A GitHub access token. | - |
| `GITHUB_REPOSITORY` | The owner and repository name | `owner/repo` of the repository where GitHub Actions are running |
| `GITHUB_API_URL` | The GitHub API URL | `https://api.github.com` |
| `SLACK_API_TOKEN` | A Slack OAuth access token | - |
| `SLACK_WEBHOOK_URL` | A Slack incoming webhook URL | - |
| `SLACK_CHANNEL` | A notify Slack channel | - |
| `GITHUB_ASSIGNEES_SAMPLE` | Number of users to randomly select from those listed in the `assignees:` section. | - |

#### Required scope of `SLACK_API_TOKEN`

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
