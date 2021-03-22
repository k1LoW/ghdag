## [v0.13.1](https://github.com/k1LoW/ghdag/compare/v0.13.0...v0.13.1) (2021-03-22)

* Remove variables `code_owners` `code_owners_who_approved` [#62](https://github.com/k1LoW/ghdag/pull/62) ([k1LoW](https://github.com/k1LoW))
* Fix reviewers count [#61](https://github.com/k1LoW/ghdag/pull/61) ([k1LoW](https://github.com/k1LoW))

## [v0.13.0](https://github.com/k1LoW/ghdag/compare/v0.12.1...v0.13.0) (2021-03-19)

* Dump `if:` section variables instead of github.event JSON [#60](https://github.com/k1LoW/ghdag/pull/60) ([k1LoW](https://github.com/k1LoW))
* Add `check` command [#59](https://github.com/k1LoW/ghdag/pull/59) ([k1LoW](https://github.com/k1LoW))

## [v0.12.1](https://github.com/k1LoW/ghdag/compare/v0.12.0...v0.12.1) (2021-03-19)

* Add variable `login` [#58](https://github.com/k1LoW/ghdag/pull/58) ([k1LoW](https://github.com/k1LoW))
* Add variables `env.*` [#57](https://github.com/k1LoW/ghdag/pull/57) ([k1LoW](https://github.com/k1LoW))

## [v0.12.0](https://github.com/k1LoW/ghdag/compare/v0.11.1...v0.12.0) (2021-03-17)

* Add `linkedNames:` for linking the GitHub user or team name to the Slack user or team account name. [#56](https://github.com/k1LoW/ghdag/pull/56) ([k1LoW](https://github.com/k1LoW))

## [v0.11.1](https://github.com/k1LoW/ghdag/compare/v0.11.0...v0.11.1) (2021-03-17)

* Fix reviewers: action [#55](https://github.com/k1LoW/ghdag/pull/55) ([k1LoW](https://github.com/k1LoW))

## [v0.11.0](https://github.com/k1LoW/ghdag/compare/v0.10.3...v0.11.0) (2021-03-17)

* [BREAKING] Remove variables `github_event_name` `github_event_action` [#54](https://github.com/k1LoW/ghdag/pull/54) ([k1LoW](https://github.com/k1LoW))
* Add variables `github.event_name` `github.event.*`, similar to GitHub Actions [#53](https://github.com/k1LoW/ghdag/pull/53) ([k1LoW](https://github.com/k1LoW))
* Remove debug print [#52](https://github.com/k1LoW/ghdag/pull/52) ([k1LoW](https://github.com/k1LoW))
* Fix GitHub event handling [#51](https://github.com/k1LoW/ghdag/pull/51) ([k1LoW](https://github.com/k1LoW))

## [v0.10.3](https://github.com/k1LoW/ghdag/compare/v0.10.2...v0.10.3) (2021-03-16)

* Fix env set bug [#50](https://github.com/k1LoW/ghdag/pull/50) ([k1LoW](https://github.com/k1LoW))

## [v0.10.2](https://github.com/k1LoW/ghdag/compare/v0.10.1...v0.10.2) (2021-03-16)

* Add variable `is_called` [#49](https://github.com/k1LoW/ghdag/pull/49) ([k1LoW](https://github.com/k1LoW))

## [v0.10.1](https://github.com/k1LoW/ghdag/compare/v0.10.0...v0.10.1) (2021-03-12)

* [BREAKING]Stop adding signatures ( `<!-- ghdag:*:* -->` ) to comments [#48](https://github.com/k1LoW/ghdag/pull/48) ([k1LoW](https://github.com/k1LoW))

## [v0.10.0](https://github.com/k1LoW/ghdag/compare/v0.9.0...v0.10.0) (2021-03-12)

* [BREAKING] Output log to STDERR [#47](https://github.com/k1LoW/ghdag/pull/47) ([k1LoW](https://github.com/k1LoW))
* Add `if` command [#46](https://github.com/k1LoW/ghdag/pull/46) ([k1LoW](https://github.com/k1LoW))

## [v0.9.0](https://github.com/k1LoW/ghdag/compare/v0.8.0...v0.9.0) (2021-03-11)

* Add variable `github_event_action` [#45](https://github.com/k1LoW/ghdag/pull/45) ([k1LoW](https://github.com/k1LoW))
* Add `GHDAG_ACTION_LABELS_BEHAVIOR` to set behavior of the `labels:` action [#44](https://github.com/k1LoW/ghdag/pull/44) ([k1LoW](https://github.com/k1LoW))

## [v0.8.0](https://github.com/k1LoW/ghdag/compare/v0.7.1...v0.8.0) (2021-03-11)

* Add `do` command to perform ghdag action as oneshot command [#40](https://github.com/k1LoW/ghdag/pull/40) ([k1LoW](https://github.com/k1LoW))

## [v0.7.1](https://github.com/k1LoW/ghdag/compare/v0.7.0...v0.7.1) (2021-03-10)

* Fix: invalid memory address or nil pointer dereference when using temporary SLACK_API_TOKEN [#43](https://github.com/k1LoW/ghdag/pull/43) ([k1LoW](https://github.com/k1LoW))

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
