# ghdag

[![build](https://github.com/k1LoW/ghdag/actions/workflows/ci.yml/badge.svg)](https://github.com/k1LoW/ghdag/actions) [![ghdag workflow](https://github.com/k1LoW/ghdag/actions/workflows/ghdag_workflow.yml/badge.svg)](https://github.com/k1LoW/ghdag/actions/workflows/ghdag_workflow.yml)

:octocat: `ghdag` is a tiny workflow engine for GitHub issue and pull request.

Key features of `ghdag` are:

- Simple definition of workflows to improve the lifecycle of issues and pull requests.
- Built-in GitHub and Slack actions.
- Optimized for running on GitHub Actions, it is easier to configure when running on GitHub Actions.

See **[Examples](examples.md)** page for practical usage.

## Getting Started

### Generate a workflow file

Execute `ghdag init` for generating a workflow YAML file for `ghdag` ( and a workflow YAML file for GitHub Actions ).

``` console
$ ghdag init myworkflow
2021-02-23T23:29:48+09:00 [INFO] ghdag version 0.2.3
2021-02-23T23:29:48+09:00 [INFO] Creating myworkflow.yml
Do you generate a workflow YAML file for GitHub Actions? (y/n) [y]: y
2021-02-23T23:29:48+09:00 [INFO] Creating .github/workflows/ghdag_workflow.yml
$ cat myworkflow.yml
---
# generate by ghdag init
tasks:
  -
    id: set-question-label
    if: 'is_issue && len(labels) == 0 && title endsWith "?"'
    do:
      labels: [question]
    ok:
      run: echo 'Set labels'
    ng:
      run: echo 'failed'
    name: Set 'question' label
$ cat .github/workflows/ghdag_workflow.yml
---
# generate by ghdag init
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

And edit myworkflow.yml.

### Run workflow on your machine

``` console
$ export GITHUB_TOKEN=xxXxXXxxXXxx
$ export GITHUB_REPOGITORY=k1LoW/myrepo
$ ghdag run myworkflow.yml
2021-02-28T00:26:41+09:00 [INFO] ghdag version 0.2.3
2021-02-28T00:26:41+09:00 [INFO] Start session
2021-02-28T00:26:41+09:00 [INFO] Fetch all open issues and pull requests from k1LoW/myrepo
2021-02-28T00:26:42+09:00 [INFO] 3 issues and pull requests are fetched
2021-02-28T00:26:42+09:00 [INFO] 1 tasks are loaded
2021-02-28T00:26:42+09:00 [INFO] [#14 << set-question-label] [DO] Replace labels: question
Set labels
2021-02-28T00:26:43+09:00 [INFO] Session finished
$
```

### Run workflow on GitHub Actions

``` console
$ git add myworkflow.yml
$ git add .github/workflows/ghdag_workflow.yml
$ git commit -m'Initial commit for ghdag workflow'
$ git push origin
```

And, see `https://github.com/[owner]/[repo]/actions`

#### Issues and pull requests targeted by ghdag

:memo: The issues and pull requests that `ghdag` fetches varies depending on the environment in which it is run.

| Environment or [Event that trigger workflows](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target) for GitHub Actions | Issues and pull requests fetched by `ghdag` |
| --- | --- |
| On your machine | **All** `opened` and `not draft` issues and pull requests |
| `issues` | A **single** `opened` issue that triggered the event |
| `issue_comment` | A **single** `opened` and `not draft` issue or pull request that triggered the event |
| `pull_request` `pull_request_*` | A **single** `opened` and `not draft` pull request that triggered the event |
| Other events | **All** `opened` and `not draft` issues and pull requests |

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
| `state` | `string` | State of the issue (pull request) |
| `title` | `string` | Title of the issue (pull request) |
| `body` | `string` | Body of the issue (pull request) |
| `url` | `string` | URL of the issue (pull request) |
| `author` | `string` | Author of the issue (pull request) |
| `labels` | `array` | Labels that are set for the issue |
| `assignees` | `array` | Assignees of the issue (pull request) |
| `reviewers` | `array` | Reviewers of the pull request (including code owners) |
| `code_owners` | `array` | Code owners of the pull request |
| `reviewers_who_approved` | `array` | Reviewers who approved the pull request (including code owners) |
| `code_owners_who_approved` | `array` | Code owners who approved the pull request |
| `is_issue` | `bool` | `true` if the target type of the workflow is "Issue" |
| `is_pull_request` | `bool` | `true` if the target type of the workflow is "Pull request" |
| `is_approved` | `bool` | `true` if the pull request has received an approving review  |
| `is_review_required` | `bool` | `true` if a review is required before the pull request can be merged |
| `is_change_requested` | `bool` | `true` if changes have been requested on the pull request |
| `mergeable` | `bool` | `true` if the pull request can be merged. |
| `changed_files` | `int` | Number of changed files in this pull request |
| `hours_elapsed_since_created` | `int` | Hours elspsed since the issue (pull request) created |
| `hours_elapsed_since_updated` | `int` | Hours elspsed since the issue (pull request) updated |
| `number_of_comments` | `int` | Number of comments |
| `latest_comment_author` | `string` | Author of latest comment |
| `latest_comment_body` | `string` | Body of latest comment |
| `login` | `string` | User of `GITHUB_TOKEN` |
| `is_called` | `true` | `true` if the target is called |
| `caller_action_run_stdout` | `string` | Latest caller STDOUT of the `run` action |
| `caller_action_run_stderr` | `string` | Latest caller STDERR of the `run` action |
| `caller_action_labels_updated` | `array` | Latest caller update result of the `labels:` action |
| `caller_action_assignees_updated` | `array` | Latest caller update result of the `assgnees:` action |
| `caller_action_reviewers_updated` | `array` | Latest caller update result of the `reviewers:` action |
| `caller_action_comment_created` | `string` | Latest caller created comment of the `comment:` action |
| `caller_action_state_changed` | `string` | Latest caller changed state of the `state:` action |
| `caller_action_notify_sent` | `string` | Latest caller sent message of the `notify:` action |
| `caller_action_do_error` | `string` | Latest caller error message when `do:` action failed |
| `year` | `int` | Year of current time (UTC) |
| `month` | `int` | Month of current time (UTC) |
| `day` | `int` | Day of current time (UTC) |
| `hour` | `int` | Hour of current time (UTC) |
| `weekday` | `int` | Weekday of current time (UTC) (Sunday = 0, ...) |
| `github.event_name` | `string` | Event name of GitHub Actions ( ex. `issues`, `pull_request` )|
| `github.event` | `object` | Detailed data for each event of GitHub Actions (ex. `github.event.action`, `github.event.label.name` ) |
| `env.<env_name>` | `string` | The value of a specific environment variable |

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
| `GHDAG_CALLER_TASK_ID` | Task ID of the caller via the `next:` action |
| `GHDAG_ACTION_RUN_STDOUT` | STDOUT of the `run` action |
| `GHDAG_ACTION_RUN_STDERR` | STDERR of the `run` action |
| `GHDAG_ACTION_LABELS_UPDATED` | Update result of the `labels:` action |
| `GHDAG_ACTION_ASSIGNEES_UPDATED` | Update result of the `assgnees:` action |
| `GHDAG_ACTION_REVIEWERS_UPDATED` | Update result of the `reviewers:` action |
| `GHDAG_ACTION_COMMENT_CREATED` | Created comment of the `comment:` action |
| `GHDAG_ACTION_STATE_CHANGED` | Changed state of the `state:` action |
| `GHDAG_ACTION_NOTIFY_SENT` | Sent message of the `notify:` action |
| `GHDAG_ACTION_DO_ERROR` | Error message when `do:` action failed |
| `GHDAG_TASK_*` | [Variables available in the `if:` section](https://github.com/k1LoW/ghdag#available-variables). ( ex. `number` -> `GHDAG_TASK_NUMBER` ) |

#### `tasks[*].<action_type>.run:`

Execute command using `sh -c`.

**Example**

``` yaml
do:
  run: echo 'execute command'
```

#### `tasks[*].<action_type>.labels:`

Update the labels of the target issue or pull request.

**Example**

``` yaml
do:
  labels: [question, bug]
```

#### `tasks[*].<action_type>.assignees:`

Update the assignees of the target issue or pull request.

**Example**

``` yaml
do:
  assignees: [alice, bob, charlie]
env:
  GITHUB_ASSIGNEES_SAMPLE: 1
```

#### `tasks[*].<action_type>.reviewers:`

Update the reviewers for the target pull request.

However, [Code owners](https://docs.github.com/en/github/creating-cloning-and-archiving-repositories/about-code-owners) has already been registered as a reviewer, so it is excluded.

**Example**

``` yaml
do:
  reviewers: [alice, bob, charlie]
env:
  GITHUB_REVIEWERS_SAMPLE: 2
```

#### `tasks[*].<action_type>.comment:`

Create new comment to the target issue or pull request.

**Example**

``` yaml
do:
  labels: [question]
ok:
  comment: Thank you for your question :+1:
```

#### `tasks[*].<action_type>.state:`

Change state the the target issue or pull request.

**Example**

``` yaml
if: hours_elapsed_since_updated > (24 * 30)
do:
  state: close
```

##### Changeable states

| Target | Changeable states |
| --- | --- |
| Issue | `close` |
| Pull request | `close` `merge` |

#### `tasks[*].<action_type>.notify:`

Send notify message to Slack channel.

**Example**

``` yaml
ng:
  notify: error ${GHDAG_ACTION_OK_ERROR}
env:
  SLACK_CHANNEL: workflow-alerts
  SLACK_MENTIONS: bob
```

##### Required environment variables

- ( `SLACK_API_TOKEN` and `SLACK_CHANNEL` ) or `SLACK_WEBHOOK_URL`

#### `tasks[*].<action_type>.next:`

Call next tasks in the same session.

**Example**

``` yaml
tasks:
  -
    id: set-question-label
    if: 'is_issue && len(labels) == 0 && title endsWith "?"'
    do:
      labels: [question]
    ok:
      next [notify-slack, add-comment]
  -
    id: notify-slack
    do:
      notify: A question has been posted.
  -
    id: add-comment
    do:
      comment: Thank you your comment !!!
```

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
| `GITHUB_ASSIGNEES_SAMPLE` | Number of users to randomly select from those listed in the `assignees:` action. | - |
| `GITHUB_REVIEWERS_SAMPLE` | Number of users to randomly select from those listed in the `reviewers:` action. | - |
| `GITHUB_COMMENT_MENTIONS` | Mentions to be given to comment | - |
| `GITHUB_COMMENT_MENTIONS_SAMPLE` | Number of users to randomly select from those listed in `GITHUB_COMMENT_MENTIONS`. | - |
| `SLACK_MENTIONS` | Mentions to be given to Slack message | - |
| `SLACK_MENTIONS_SAMPLE` | Number of users to randomly select from those listed in `SLACK_MENTIONS`. | - |
| `GHDAG_SAMPLE_WITH_SAME_SEED` | Sample using the same random seed as the previous action/task or not. | - |
| `SLACK_USERNAME` | Custom `username` of slack message. Require `chat:write.customize` scope. | |
| `SLACK_ICON_EMOJI` | Custom `icon_emoji` of slack message. Require `chat:write.customize` scope. | |
| `SLACK_ICON_URL` | Custom `icon_url` of slack message. Require `chat:write.customize` scope. | |
| `GITHUB_ASSIGNEES` | Additional Assignees to the list in the `assignees:` action | - |
| `GITHUB_REVIEWERS` | Additional Reviewers to the list in the `reviewers:` action | - |
| `GHDAG_ACTION_LABELS_BEHAVIOR` | Behavior of the `labels:` action ( `replace` (=default), `add`, `remove` ) | - |
| `GHDAG_ACTION_COMMENT_MAX` | Maximum number of consecutive comments by the same login ( default: 5 ) | - |

#### Required scope of `SLACK_API_TOKEN`

- `channel:read`
- `chat:write`
- `chat:write.public`
- `users:read`
- `usergroups:read`
- `chat:write.customize` ( optional )

## Link the GitHub user or team name to the Slack user or team account name

Provides the feature `linkedNames:` to link GitHub and Slack transparently even if they have different account names.

**Example**

Send a Slack message with a mention ( `@bob_marly` ) to the author ( `@bob` ) of the pull request.

``` yaml
tasks:
  -
    id: notify-slack
    if: 'is_pull_request && is_approved && mergeable'
    do:
      notify: Your pull request is ready for merge !
    env:
      SLACK_MENTIONS: ${GHDAG_TASK_AUTHOR}
linkedNames:
  -
    github: bob
    slack: bob_marly
```

## Use ghdag as the one-shot command on GitHub Actions

`ghdag` can be used not only as a workflow engine, but also as a utility command in jobs on GitHub Actions.

``` console
$ ghdag do --help
do action.

Usage:
  ghdag do [command]

Available Commands:
  assignees   update the assignees of the target issue or pull request
  comment     create the comment of the target issue or pull request
  labels      update the labels of the target issue or pull request
  notify      send notify message to slack channel
  reviewers   update the reviewers of the target issue or pull request
  run         execute command using `sh -c`
  state       change state of the target issue or pull request

Flags:
  -h, --help   help for do

Use "ghdag do [command] --help" for more information about a command.
$ ghdag if --help
check condition.

Usage:
  ghdag if [CONDITION] [flags]

Flags:
  -h, --help         help for if
  -n, --number int   issue or pull request number
```

**Example**

``` yaml
name: labels

on:
  issues:
    types: [created]

jobs:
  set-labels:
    name: Set labels
    runs-on: ubuntu-latest
    container: ghcr.io/k1low/ghdag:latest
    steps:
      - name: Set labels
        run: ghdag do labels question
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

``` yaml
name: build

jobs:
  job-tests:

# snip

      - name: Send notification to channel
        if: failure()
        run: ghdag do notify 'Test faild'
        env:
          SLACK_API_TOKEN: ${{ secrets.SLACK_API_TOKEN }}
          SLACK_CHANNEL: operations
          SLACK_MENTIONS: bob
```

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

**apk:**

Use [apk-add-from-url](https://github.com/k1LoW/apk-add-from-url)

``` console
$ export GHDAG_VERSION=X.X.X
$ curl -L https://git.io/apk-add-from-url | sh -s -- https://github.com/k1LoW/ghdag/releases/download/v$GHDAG_VERSION/ghdag_$GHDAG_VERSION-1_amd64.apk
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

**docker:**

```console
$ docker pull ghcr.io/k1low/ghdag:latest
```
