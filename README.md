# ghdag

:octocat: `ghdag` is a tiny workflow engine for GitHub issue and pull request.

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
    if: 'is_issue && len(labels) == 0 && title endsWith "?"'
    do:
      comment: 'Good :+1:'
    ok:
      run: echo 'commented'
    ng:
      run: echo 'fail comment'
    name: Set 'question' label
```

And edit myworkflow.yml.

### Run workflow on your machine

``` console
$ export GITHUB_TOKEN=xxXxXXxxXXxx
$ export GITHUB_REPOGITORY=k1LoW/myrepo
$ ghdag run myworkflow.yml
2021-02-28T00:26:41+09:00 [INFO] ghdag version 0.0.1
2021-02-28T00:26:41+09:00 [INFO] Start session
2021-02-28T00:26:41+09:00 [INFO] Fetch open issues and pull requests from k1LoW/myrepo
2021-02-28T00:26:42+09:00 [INFO] 3 issues and pull requests are fetched
2021-02-28T00:26:42+09:00 [INFO] 1 tasks are loaded
2021-02-28T00:26:42+09:00 [INFO] [#14 << sample-task] [DO] Add comment: Good :+1:
commented
2021-02-28T00:26:43+09:00 [INFO] Session finished
$
```

### Run workflow on GitHub Actions

``` yaml
name: ghdag workflow
on:
  issues:
    types: [opened]
  issue_comment:
    types: [created]
  pull_request:
    types: [opened]

jobs:
  run-workflow:
    name: Run workflow
    runs-on: ubuntu-latest
    container: ghcr.io/k1low/ghdag:latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
        with:
          token: ${{ secrets.GITHUB_TOKEN }}
      - name: Run ghdag
        run: ghdag run myworkflow.yml
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Workflow syntax

`ghdag` requires a YAML file to define your workflow configuration.

#### `env:`

A map of environment environment variables in global scope (available to all tasks in the workflow).

**Example**

``` yaml
env:
  SERVER: production
  GITHUB_TOKEN: ${GHDAG_GITHUB_TOKEN}
```

#### `tasks:`

A workflow run is made up of one or more tasks. Tasks run in sequentially.

**Example**

``` yaml
tasks:
  -
    id: first
    if: 'is_issue && len(labels) == 0 && title endsWith "?"'
    do:
      labels: [question]
    ok:
      run: echo 'Set labels'
    ng:
      run: echo 'failed'
    name: Set 'question' label
  -
    id: second
    if: len(assignees) == 0
    do:
      assignees: [k1LoW]
    name: Assign me
```

#### `tasks[*].id:`

Each task have an id to be called by another task.

#### `tasks[*].name:`

Name of task.

#### `tasks[*].if:`

The task will not be performed unless the condition in the `if` section is met.

If the `if` section is missing, the task will not be performed unless it is called by another task.

##### Expression syntax

`ghdag` use [antonmedv/expr](https://github.com/antonmedv/expr).

See [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md).

##### Available variables

The variables available in the `if` section are as follows

| Variable name | Type | Description |
| --- | --- | --- |
| `number` | `int` | Number of the issue (pull request) |
| `title` | `string` | Title of the issue (pull request) |
| `body` | `string` | Body of the issue (pull request) |
| `url` | `string` | URL of the issue (pull request) |
| `author` | `string` | Author of the issue (pull request) |
| `labels` | `array` | Labels that are set for the issue |
| `assignees` | `array` | Assignees of the issue (pull request) |
| `reviewers` | `array` | Reviewers of the pull request (including code owners) |
| `code_owners` | `array` | Code owners of the pull request |
| `is_issue` | `bool` | `true` if the target type of the workflow is "Issue" |
| `is_pull_request` | `bool` | `true` if the target type of the workflow is "Pull request" |
| `is_approved` | `bool` | `true` if the pull request has received an approving review  |
| `is_review_required` | `bool` | `true` if a review is required before the pull request can be merged |
| `is_change_requested` | `bool` | `true` if changes have been requested on the pull request |
| `mergeable` | `bool` | `true` if the pull request can be merged. |
| `hours_elapsed_since_created` | `int` | Hours elspsed since the issue (pull request) created |
| `hours_elapsed_since_updated` | `int` | Hours elspsed since the issue (pull request) updated |
| `number_of_comments` | `int` | Number of comments |
| `latest_comment_author` | `string` | Author of latest comment |
| `latest_comment_body` | `string` | Body of latest comment |

#### `tasks[*].env:`

A map of environment environment variables in the scope of each task.

#### `tasks[*].do:`, `tasks[*].ok:`, `tasks[*].ng:`

A task has 3 actions ( called `do`, `ok` and `ng` ) with predetermined steps to be performed.

1. Perform `do:` action.
2. If the `do:` action succeeds, perform the `ok:` action.
2. If the `do:` action fails, perform the `ng:` action.

##### Available builtin environment variables

| Environment variable | Description |
| --- | --- |
| `GHDAG_TASK_ID` | Task ID |
| `GHDAG_ACTION_OK_ERROR` | Error message when `ok:` action failed |
| `GHDAG_TASK_*` | Variables available in the `if:` section. ( ex. `number` -> `GHDAG_TASK_NUMBER` ) |

#### `tasks[*].<action_type>.run:`

Action: Execute command.

#### `tasks[*].<action_type>.labels:`

Action: Update labels.

#### `tasks[*].<action_type>.assignees:`

Action: Update assignees.

#### `tasks[*].<action_type>.reviewers:`

Action: Update reviewers.

#### `tasks[*].<action_type>.comment:`

Action: Add comment.

#### `tasks[*].<action_type>.state:`

Action: Change state.

#### `tasks[*].<action_type>.notify:`

Action: Notify message using Slack.

#### `tasks[*].<action_type>.next:`

Action: Call next tasks.

## Environment variables for configuration

| Environment variable | Description | Default on GitHub Actions |
| --- | --- | --- |
| `GITHUB_TOKEN` | A GitHub access token. | - |
| `GITHUB_REPOSITORY` | The owner and repository name | `owner/repo` of the repository where GitHub Actions are running |
| `GITHUB_API_URL` | The GitHub API URL | `https://api.github.com` |
| `GITHUB_GRAPHQL_URL` | The GitHub GraphQL API URL | `https://api.github.com/graphql` |
| `SLACK_API_TOKEN` | A Slack OAuth access token | - |
| `SLACK_WEBHOOK_URL` | A Slack incoming webhook URL | - |
| `SLACK_CHANNEL` | A Slack channel to be notified | - |
| `GITHUB_ASSIGNEES_SAMPLE` | Number of users to randomly select from those listed in the `assignees:` section. | - |
| `GITHUB_REVIEWERS_SAMPLE` | Number of users to randomly select from those listed in the `reviewers:` section. | - |
| `SLACK_MENTIONS` | Mentions to be given to notifications | - |
| `SLACK_MENTIONS_SAMPLE` | Number of users to randomly select from those listed in `SLACK_MENTIONS`. | - |

#### Required scope of `SLACK_API_TOKEN`

- `channel:read`
- `chat:write`
- `chat:write.public`
- `users:read`
- `usergroups:read`

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
