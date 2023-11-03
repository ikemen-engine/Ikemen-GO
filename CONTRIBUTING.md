# Contributing to Ikemen GO

We would love for you to contribute to Ikemen GO and help make it even better than it is today!
As a contributor, here are the guidelines we would like you to follow:

 - [Question or Problem?](#question)
 - [Issues and Bugs](#issue)
 - [Feature Requests](#feature)
 - [Branching Strategy](#branching-strategy)
 - [Submission Guidelines](#submit)
 - [PR Message Guidelines](#pr)
 - [Commit Message Guidelines](#commit)


## <a name="question"></a> Got a Question or Problem?

Do not open issues for general support questions as we want to keep GitHub issues for bug reports. Q&As are allowed on [discussions section][discussions], but in most cases we prefer that you use the *ikemen-help* section of [our Discord server][discord], which you can find an invitation to on the [Ikemen GO website][website]. This is because many problems can be solved by members of the community who do not use GitHub. Please remember to check the [wiki][wiki] page before asking a question and use the search bar before creating a new feature request topic.


## <a name="issue"></a> Found a Bug?

If you find a bug, you can help us by [submitting an issue](#submit-issue) to our [GitHub Repository][github].
Even better, you can [submit a Pull Request](#submit-pr) with a fix.


## <a name="feature"></a> Missing a Feature?

You can *request* a new feature by [starting a discussion](#discussions) about it in our GitHub Repository.
If you would like to *implement* a new feature, please consider the size of the change in order to determine the right steps to proceed:

* For a **Major Feature**, first open a discussion and outline your proposal so that it can be discussed.
  This process allows us to better coordinate our efforts, prevent duplication of work, and help you to craft the change so that it is successfully accepted into the project.

* **Small Features** can be crafted and directly [submitted as a Pull Request](#submit-pr).

## <a name="branching-strategy"></a> Branching Strategy

Our project utilizes a specific branching strategy to ensure a well-organized and stable codebase. Here's an outline of how our branches are structured:

- `master`: Represents the most stable version of the code. It's updated only when a new stable release is ready, after merging and tagging the `release` branch with a version number.
- `develop`: The active development branch where all feature branches are created and merged back into. This branch contains features that will be part of the next release cycle.
- `release`: Created off the `develop` branch when we're ready for a new release cycle. It's reserved for preparing the release and will only receive bug fixes.

## <a name="submit"></a> Submission Guidelines

### <a name="submit-issue"></a> Submitting an Issue

Before submitting an issue, it is recommended that you search the issue tracker to see if your problem has already been reported. If an issue exists, the discussion might provide you with readily available workarounds.

It is also advisable to test the problematic content against the latest build, preferably using the [nightly development release][nightly] in addition to the [latest release][latest], as the problem may have already been resolved. Once a new release is pushed, previous releases are no longer supported.

We strive to resolve all issues as quickly as possible, but before we can fix a bug, we must first reproduce and confirm it. To do so, we require a minimal reproduction. Ideally, the minimal reproduction should include a link to the problematic content and detailed information what has to be done for the bug to occur. If the content is related to a complicated piece of code, preparing a test case with resources shipped with the engine (e.g. kfm/kfmz character or default screenpack) increases the chances of the bug being fixed.

A minimal reproduction provides us with a wealth of important information without the need for additional questions. It enables us to quickly confirm a bug or identify a coding problem, as well as ensure that we are addressing the correct issue. Having a minimal reproduction saves our developers' time, allowing us to fix more bugs. We understand that it can be challenging to extract essential bits of code from a larger codebase, but isolating the problem is crucial for us to be able to fix it.

Unfortunately, we cannot investigate or fix bugs without a minimal reproduction. If we do not receive enough information to reproduce the issue, we will have to close the issue.

To file a new issue, you can select from our [new issue templates][templates] and fill out the issue template.

### <a name="submit-pr"></a> Submitting a Pull Request (PR)

Before you submit your Pull Request (PR), please follow these guidelines:

1. Check [GitHub][pulls] for existing PRs that may be similar to your submission.
2. Ensure there's an issue that describes your fix or the feature you're adding. Design discussions should happen before starting your work.
3. [Fork][fork] the Ikemen GO repository.
4. In your fork, create a new git branch from the appropriate base branch.
- **For features and fixes in the next release cycle**: Branch off from `develop` and merge your changes back into it. These will be included in the next release cycle.
   ```shell
   git checkout -b my-branch-name develop
   ```
- **For fixes in the current release cycle**: Direct your fixes to the `release` branch, which is strictly for regression fixes and preparations for the current release.
   ```shell
   git checkout -b my-branch-name release
   ```
- **For stable release updates**: The `master` branch is read-only and updated by github maintainers exclusively for deploying new stable releases.
   ```shell
   git checkout -b my-branch-name master
   ```

5. Make your changes and commit them:
   
   ```shell
   git commit --all
   ```

6. Push your branch to GitHub:
   
   ```shell
   git push origin my-branch-name
   ```

7. Submit a PR to the correct base branch on `Ikemen GO` (either `develop` or `release`).

#### Reviewing a Pull Request

All PRs are subject to review by the Ikemen GO team, which retains the right to decline any contributions.

#### Addressing Review Feedback

If changes are requested:

1. Make the required updates.
2. Commit the changes and push them to update your PR.

##### Updating the Commit Message

To update the commit message of the last commit:

1. Check out your branch and amend the commit message.
2. Force push to your repository to update the PR.

#### After Your Pull Request Is Merged

Once merged, you can delete your branch:

1. Delete the remote branch on GitHub.
2. Update your local `master` with the latest from the upstream repository.
3. Delete your local branch.

Remember to work from the `develop` branch for features and the `release` branch for bug fixes. The `master` branch is now for stable releases only.

Remember to base your work on the `develop` branch when contributing new features for the upcoming release cycle, and the release branch for urgent bug fixes targeting the current release. The master branch serves as the archive for stable releases only and is updated strictly with finalized releases.

## <a name="pr"></a> PR Message Format

We use [Conventional Commits][cc] specification for adding human and machine readable meaning to pull requests.
The expected PR title formatting is:
```
<type>(<scope>): <short summary>
  │       │             │
  │       │             └─⫸ Summary in present tense. Not capitalized. No period at the end.
  │       │
  │       └─⫸ Scope: Scope of changes, e.g.: input|sctrl|trigger etc. Optional, can be skipped.
  │
  └─⫸ Type: build|docs|feat|fix|other|perf|refactor|style|test
```

### <a name="pr-type"> Type

The `<type>` portion of the title must be one of the following:
- **build**: Changes that affect the build system, external dependencies, CI configuration
- **docs**: Documentation only changes
- **feat**: A new feature
- **fix**: A bug fix
- **other**: Changes that do not belong to any other category (e.g., fixes for already merged PRs, not meant to show up in the changelog)
- **perf**: A code change that improves performance
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- **test**: Adding missing tests or correcting existing tests

If you have difficulty determining the appropriate classification for your pull request, the reviewer will assist in doing so before the merge. Please note that the Ikemen GO team reserves the right to modify pull request titles and translate the content of pull request messages to enhance readability for the general audience and developer community.

### <a name="pr-scope"> Scope

The `(<scope>)` portion of the title refers to the scope of changes, such as input, sctrl, trigger, and so on. It is optional and can be skipped.

### <a name="pr-summary"></a> Summary

Use the summary field to provide a succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize the first letter
* no dot (.) at the end

### <a name="pr-body"></a> Message Body

Just as in the summary, use the imperative, present tense: "fix" not "fixed" nor "fixes".

Explain the motivation for the change in the commit message body. This commit message should explain _why_ you are making the change.
You can include a comparison of the previous behavior with the new behavior in order to illustrate the impact of the change.

### <a name="commit-footer"></a> Message Footer

The footer can contain information about breaking changes and deprecations and is also the place to reference GitHub issues, discussions, and other PRs that this commit closes or is related to.
For example:

```
BREAKING CHANGE: <breaking change summary>
DEPRECATED: <what is deprecated>
Fixes #<issue number>
```

## <a name="commit"></a> Commit Message Format

Unlike pull requests, which are used for automatic generation of changelogs, there is no strict convention for commit titles. It is optional to follow the Conventional Commits specification described in the [PR Message Format](#pr).


### Revert commits

If the commit reverts a previous commit, it should begin with `revert: `, followed by the header of the reverted commit.

The content of the commit message body should contain:

- information about the SHA of the commit being reverted in the following format: `This reverts commit <SHA>`,
- a description of the reason for reverting the commit message.

[cc]: https://www.conventionalcommits.org/
[discord]: https://discord.com/invite/QWxxwjE
[discussions]: https://github.com/ikemen-engine/Ikemen-GO/discussions
[fork]: https://docs.github.com/en/github/getting-started-with-github/fork-a-repo
[git]: https://git-scm.com/docs/git-rebase#interactive_mode
[github]: https://github.com/ikemen-engine/Ikemen-GO
[latest]: https://github.com/ikemen-engine/Ikemen-GO/releases/latest
[nightly]: https://github.com/ikemen-engine/Ikemen-GO/releases/tag/nightly
[pulls]: https://github.com/ikemen-engine/Ikemen-GO/pulls
[templates]: https://github.com/ikemen-engine/Ikemen-GO/issues/new/choose
[website]: https://ikemen-engine.github.io/
[wiki]: https://github.com/ikemen-engine/Ikemen-GO/wiki
