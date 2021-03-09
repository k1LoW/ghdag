# Workflow Examples

## Randomly assign one member to the issue.

``` yaml
# myworkflow.yml
---
tasks:
  -
    id: assign-issue
    if: 'is_issue && len(assignees) == 0'
    do:
      assignees: [alice bob charlie]
    env:
      GITHUB_ASSIGNEES_SAMPLE: 1
```

<details>

<summary>GitHub Actions workflow YAML file</summary>

``` yaml
# .github/workflows/ghdag_workflow.yml
name: ghdag workflow
on:
  issues:
    types: [opened]
  issue_comment:
    types: [created]

jobs:
  run-workflow:
    name: 'Run workflow for A **single** `opened` issue that triggered the event'
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

</details>

## Assign a reviewer to a pull request with the 'needs review' label

``` yaml
# myworkflow.yml
---
tasks:
  -
    id: assign-pull-request
    if: 'is_pull_request && len(reviewers) == 0 && "needs review" in labels'
    do:
      reviewers: [alice bob charlie]
    env:
      GITHUB_REVIEWERS_SAMPLE: 1
```

<details>

<summary>GitHub Actions workflow YAML file</summary>

``` yaml
# .github/workflows/ghdag_workflow.yml
name: ghdag workflow
on:
  pull_request:
    types: [labeled]

jobs:
  run-workflow:
    name: 'Run workflow for A **single** `opened` issue that triggered the event'
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

</details>

## Close issues that have not been updated in 30 days

``` yaml
# myworkflow.yml
---
tasks:
  -
    id: close-issues-30days
    if: hours_elapsed_since_updated > (30 * 24)
    do:
      state: close
```

<details>

<summary>GitHub Actions workflow YAML file</summary>

``` yaml
# .github/workflows/ghdag_workflow.yml
name: ghdag workflow
on:
  schedule:
    # Run at 00:05 every day.
    - cron: 5 0 * * *

jobs:
  run-workflow:
    name: 'Run workflow for **All** `opened` and `not draft` issues and pull requests'
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

</details>

## Reminder for specific Issues

``` yaml
# myworkflow.yml
---
tasks:
  -
    id: remind-issue-5541
    if: 'number == 5541'
    do:
      notify: 'Issue #5541 is still open. Please try to resolve it.'
    env:
      SLACK_CHANNEL: operation
      SLACK_MENTIONS: [k1low]
```

<details>

<summary>GitHub Actions workflow YAML file</summary>

``` yaml
# .github/workflows/ghdag_workflow.yml
name: ghdag workflow
on:
  schedule:
    # Run at 10:00 every Monday.
    - cron: 0 10 * * 1

jobs:
  run-workflow:
    name: 'Run workflow for **All** `opened` and `not draft` issues and pull requests'
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
          SLACK_API_TOKEN: ${{ secrets.SLACK_API_TOKEN }}
```

</details>

## Slack Notification with Mentions in sync with GitHub assignees

``` yaml
# myworkflow.yml
---
tasks:
  -
    id: assign-issue
    if: 'is_issue && len(assignees) == 0'
    do:
      assignees: [alice bob charlie]
    ok:
      next: [notify-assignees]
    env:
      GITHUB_ASSIGNEES_SAMPLE: 1
  -
    id: notify-assignees
    do:
      notify: You are assigned
    env:
      SLACK_CHANNEL: operation
      SLACK_ICON_EMOJI: white_check_mark
      SLACK_USERNAME: assign-bot
      SLACK_MENTIONS: alice_liddel bob_marly charlie_sheen
      SLACK_MENTIONS_SAMPLE: 1
      GHDAG_SAMPLE_WITH_SAME_SEED: true
```

<details>

<summary>GitHub Actions workflow YAML file</summary>

``` yaml
# .github/workflows/ghdag_workflow.yml
name: ghdag workflow
on:
  issues:
    types: [opened]
  issue_comment:
    types: [created]

jobs:
  run-workflow:
    name: 'Run workflow for A **single** `opened` issue that triggered the event'
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
          SLACK_API_TOKEN: ${{ secrets.SLACK_API_TOKEN }}
```

</details>
