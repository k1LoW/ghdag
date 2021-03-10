## [v0.7.0](https://github.com/k1LoW/ghdag/compare/v0.6.0...v0.7.0) (2021-03-10)

* Add variables `reviewers_who_approved` `code_owners_who_approved` [#42](https://github.com/k1LoW/ghdag/pull/42) ([k1LoW](https://github.com/k1LoW))
* Set caller result variables to `if:` section [#41](https://github.com/k1LoW/ghdag/pull/41) ([k1LoW](https://github.com/k1LoW))
* Use os.ExpandEnv [#39](https://github.com/k1LoW/ghdag/pull/39) ([k1LoW](https://github.com/k1LoW))
* Propagate action result environment variables to next tasks [#38](https://github.com/k1LoW/ghdag/pull/38) ([k1LoW](https://github.com/k1LoW))

## [v0.6.0](https://github.com/k1LoW/ghdag/compare/v0.5.0...v0.6.0) (2021-03-09)

* Propagating seed and excludeKey between caller and callee [#37](https://github.com/k1LoW/ghdag/pull/37) ([k1LoW](https://github.com/k1LoW))
* Set result of action environment variables [#36](https://github.com/k1LoW/ghdag/pull/36) ([k1LoW](https://github.com/k1LoW))
* Assign assignees/reviewers from environment variables [#35](https://github.com/k1LoW/ghdag/pull/35) ([k1LoW](https://github.com/k1LoW))

## [v0.5.0](https://github.com/k1LoW/ghdag/compare/v0.4.0...v0.5.0) (2021-03-08)

* Refactor runner.Runner [#34](https://github.com/k1LoW/ghdag/pull/34) ([k1LoW](https://github.com/k1LoW))
* Fix: raise EOF error when TARGET_ENV="" [#33](https://github.com/k1LoW/ghdag/pull/33) ([k1LoW](https://github.com/k1LoW))
* Add `GITHUB_COMMENT_MENTIONS` to add mentions to comment. [#32](https://github.com/k1LoW/ghdag/pull/32) ([k1LoW](https://github.com/k1LoW))
* Exclude author from reviewers [#31](https://github.com/k1LoW/ghdag/pull/31) ([k1LoW](https://github.com/k1LoW))
* Add env.ToSlice() [#30](https://github.com/k1LoW/ghdag/pull/30) ([k1LoW](https://github.com/k1LoW))

## [v0.4.0](https://github.com/k1LoW/ghdag/compare/v0.3.0...v0.4.0) (2021-03-05)

* Update Dockerfile [#29](https://github.com/k1LoW/ghdag/pull/29) ([k1LoW](https://github.com/k1LoW))
* Support custom usernme, icon_emoji, icon_url [#28](https://github.com/k1LoW/ghdag/pull/28) ([k1LoW](https://github.com/k1LoW))
* Set environment variable `GHDAG_CALLER_TASK_ID` [#27](https://github.com/k1LoW/ghdag/pull/27) ([k1LoW](https://github.com/k1LoW))
* Ensure that `GHDAG_SAMPLE_WITH_SAME_SEED` is also effective between caller task and callee task [#26](https://github.com/k1LoW/ghdag/pull/26) ([k1LoW](https://github.com/k1LoW))
* Fix GHDAG_ACTION_OK_ERROR -> GHDAG_ACTION_DO_ERROR [#25](https://github.com/k1LoW/ghdag/pull/25) ([k1LoW](https://github.com/k1LoW))
* Set STDOUT/STDERR of run action to environment variables [#24](https://github.com/k1LoW/ghdag/pull/24) ([k1LoW](https://github.com/k1LoW))

## [v0.3.0](https://github.com/k1LoW/ghdag/compare/v0.2.3...v0.3.0) (2021-03-04)

* Add an environment variable `GHDAG_SAMPLE_WITH_SAME_SEED` for sampling using the same seed as in the previous task. [#23](https://github.com/k1LoW/ghdag/pull/23) ([k1LoW](https://github.com/k1LoW))
* Generate a workflow YAML file for GitHub Actions, when `ghdag init` is executed [#22](https://github.com/k1LoW/ghdag/pull/22) ([k1LoW](https://github.com/k1LoW))
* Add ghdag workflow [#21](https://github.com/k1LoW/ghdag/pull/21) ([k1LoW](https://github.com/k1LoW))

## [v0.2.3](https://github.com/k1LoW/ghdag/compare/v0.2.2...v0.2.3) (2021-03-03)

* Increase task queue size [#20](https://github.com/k1LoW/ghdag/pull/20) ([k1LoW](https://github.com/k1LoW))

## [v0.2.2](https://github.com/k1LoW/ghdag/compare/v0.2.1...v0.2.2) (2021-03-03)

* Fix sampleByEnv [#19](https://github.com/k1LoW/ghdag/pull/19) ([k1LoW](https://github.com/k1LoW))

## [v0.2.1](https://github.com/k1LoW/ghdag/compare/v0.2.0...v0.2.1) (2021-03-03)

* Fix mention [#18](https://github.com/k1LoW/ghdag/pull/18) ([k1LoW](https://github.com/k1LoW))

## [v0.2.0](https://github.com/k1LoW/ghdag/compare/v0.1.1...v0.2.0) (2021-03-03)

* Detect issue or pull request number using GITHUB_EVENT_PATH [#17](https://github.com/k1LoW/ghdag/pull/17) ([k1LoW](https://github.com/k1LoW))

## [v0.1.1](https://github.com/k1LoW/ghdag/compare/v0.1.0...v0.1.1) (2021-03-02)

* Fix GHDAG_TARGET_* environment variables are not set [#16](https://github.com/k1LoW/ghdag/pull/16) ([k1LoW](https://github.com/k1LoW))
* Fix sampling [#15](https://github.com/k1LoW/ghdag/pull/15) ([k1LoW](https://github.com/k1LoW))

## [v0.1.0](https://github.com/k1LoW/ghdag/compare/f4ae05b30c05...v0.1.0) (2021-03-02)

* Add variable `code_owners` [#14](https://github.com/k1LoW/ghdag/pull/14) ([k1LoW](https://github.com/k1LoW))
* Set environment variables for target variables [#13](https://github.com/k1LoW/ghdag/pull/13) ([k1LoW](https://github.com/k1LoW))
* Add circuit breaker for comments [#12](https://github.com/k1LoW/ghdag/pull/12) ([k1LoW](https://github.com/k1LoW))
* Skip action if the target is already in a state of being wanted [#11](https://github.com/k1LoW/ghdag/pull/11) ([k1LoW](https://github.com/k1LoW))
* Parse comment/notify with environment variables [#10](https://github.com/k1LoW/ghdag/pull/10) ([k1LoW](https://github.com/k1LoW))
* Add variable `mergeable` [#9](https://github.com/k1LoW/ghdag/pull/9) ([k1LoW](https://github.com/k1LoW))
* Sampling mentions using SLACK_MENTIONS_SAMPLE [#8](https://github.com/k1LoW/ghdag/pull/8) ([k1LoW](https://github.com/k1LoW))
* Support Slack mentions [#7](https://github.com/k1LoW/ghdag/pull/7) ([k1LoW](https://github.com/k1LoW))
* Add action `reviewers:` [#6](https://github.com/k1LoW/ghdag/pull/6) ([k1LoW](https://github.com/k1LoW))
* Add variables `reviewers` [#5](https://github.com/k1LoW/ghdag/pull/5) ([k1LoW](https://github.com/k1LoW))
* Use GitHub GraphQL API [#4](https://github.com/k1LoW/ghdag/pull/4) ([k1LoW](https://github.com/k1LoW))
* Add environment variables expansion [#3](https://github.com/k1LoW/ghdag/pull/3) ([k1LoW](https://github.com/k1LoW))
* Sampling assignees using GITHUB_ASSIGNEES_SAMPLE [#2](https://github.com/k1LoW/ghdag/pull/2) ([k1LoW](https://github.com/k1LoW))
* Add zerolog [#1](https://github.com/k1LoW/ghdag/pull/1) ([k1LoW](https://github.com/k1LoW))
