# ghdag

:octocat: ghdag is a tiny workflow engine for GitHub issue and pull request.

## Getting Started

### Generate a workflow file

``` console
$ ghdag init myworkflow
2021-02-23T23:29:48+09:00 [INFO] ghdag version 0.0.1
2021-02-23T23:29:48+09:00 [INFO] Creating myworkflow.yml
$ cat myworkflow.yml
---
tasks:
  -
    id: sample-task
    if: '"good first issue" in labels'
    do:
      comment: 'Good :+1:'
    ok:
      run: echo 'commented'
    ng:
      run: echo 'fail comment'
    desc: This task comment when labels contains 'good first issue'
```

And edit myworkflow.yml.

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

#### Environment variables in global scope (`env:`)

#### Tasks (`tasks:`)

#### Id (`tasks[*].id:`)

#### Name (`tasks[*].name:`)

Task name.

#### Run condition (`tasks[*].if:`)

##### Expression syntax

ghdag use [antonmedv/expr](https://github.com/antonmedv/expr).

See [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).

##### Variables

| Variable name | Type | Description |
| --- | --- | --- |
| `number` | `int` | Number of the issue (pull request) |
| `title` | `string` | Title of the issue (pull request) |
| `body` | `string` | Body of the issue (pull request) |
| `url` | `string` | URL of the issue (pull request) |
| `author` | `string` | Author of the issue (pull request) |
| `labels` | `array` | Labels that are set for the issue |
| `assignees` | `array` | Assignees of the issue (pull request) |
| `reviewers` | `array` | Reviewers of the pull request |
| `is_issue` | `bool` | `true` if the target type of the workflow is "Issue" |
| `is_pull_request` | `bool` | `true` if the target type of the workflow is "Pull request" |
| `is_approved` | `bool` | `true` if the pull request has received an approving review  |
| `is_review_required` | `bool` | `true` if a review is required before the pull request can be merged |
| `is_change_requested` | `bool` | `true` if changes have been requested on the pull request |
| `hours_elapsed_since_created` | `int` | Hours elspsed since the issue (pull request) created |
| `hours_elapsed_since_updated` | `int` | Hours elspsed since the issue (pull request) updated |
| `number_of_comments` | `int` | Number of comments |
| `latest_comment_author` | `string` | Author of latest comment |

#### Environment variables in the scope of each task (`tasks[*].env:`)

#### Actions (`tasks[*].do:`, `tasks[*].ok:`, `tasks[*].ng:`)

#### Action: Execute command (`tasks[*].<action_type>.run:`)

#### Action: Update labels (`tasks[*].<action_type>.labels:`)

#### Action: Update assignees (`tasks[*].<action_type>.assignees:`)

#### Action: Update reviewers (`tasks[*].<action_type>.reviewers:`)

#### Action: Add comment (`tasks[*].<action_type>.comment:`)

#### Action: Change state (`tasks[*].<action_type>.state:`)

#### Action: Notify message using Slack (`tasks[*].<action_type>.notify:`)

#### Action: Call next tasks (`tasks[*].<action_type>.next:`)

### Environment variables

| Environment variable | Nameription | Default on GitHub Actions |
| --- | --- | --- |
| `GITHUB_TOKEN` | A GitHub access token. | - |
| `GITHUB_REPOSITORY` | The owner and repository name | `owner/repo` of the repository where GitHub Actions are running |
| `GITHUB_API_URL` | The GitHub API URL | `https://api.github.com` |
| `GITHUB_GRAPHQL_URL` | The GitHub GraphQL API URL | `https://api.github.com/graphql` |
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
