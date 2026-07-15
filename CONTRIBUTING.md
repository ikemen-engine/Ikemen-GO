# Contributing to Ikemen GO

We would love for you to contribute to Ikemen GO and help make it even better than it is today!
As a contributor, here are the guidelines we would like you to follow:

 - [Question or Problem?](#question)
 - [Issues and Bugs](#issue)
 - [Feature Requests](#feature)
 - [Branching and Release Strategy](#branching-strategy)
 - [Code Style Guidelines](#style)
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

## <a name="branching-strategy"></a> Branching and Release Strategy

- `develop` is the default branch for ongoing development and nightly builds.
- `release/X.Y` is created from `develop` when `X.Y.0` enters feature freeze. It is kept for the entire `X.Y` release line, while development of the next version continues on `develop`.
- Version tags such as `v1.0.0-rc.1`, `v1.0.0`, and `v1.0.1` are permanent.

The release process is:

1. Stabilize `release/X.Y` and tag selected commits as `vX.Y.0-rc.1`, `vX.Y.0-rc.2`, and so on.
2. Tag the approved commit as the stable `vX.Y.0` release.
3. Keep using the same branch for compatible fixes and tag them as `vX.Y.1`, `vX.Y.2`, and so on.

RCs and patch releases are tags on `release/X.Y`; they do not use separate permanent branches.

Maintainers publish a version by running the `releases` workflow from the matching `release/X.Y` branch and entering its tag. Leaving the tag empty creates a disposable test build from the selected branch; pushes to `develop` update the nightly build.

A fix that affects both lines should normally be merged into `develop` first and then cherry-picked with `git cherry-pick -x` through a PR to `release/X.Y`. Release-only fixes should be forward-ported when they also apply to `develop`. Do not merge the complete branches after they diverge.

A fix that affects both lines should normally be merged into `develop` first. A maintainer then decides whether it should be backported and creates, or asks the original contributor to create, a PR that cherry-picks the fix with `git cherry-pick -x` into `release/X.Y`. Release-only fixes should be forward-ported when they also apply to `develop`. Do not merge the complete branches after they diverge.

Only the latest stable minor release line is normally maintained. Exceptions may be announced by the maintainers.

## <a name="style"></a> Code Style Guidelines

To keep the codebase consistent and accessible to all contributors, please follow these rules:

- **Language:** All source code comments must be written in **English**.  
- **Clarity:** Write clear, concise comments that explain the intent of the code, not just what it does.  
- **Consistency:** Follow existing formatting and naming conventions in the project.  
- **Simplicity:** Prefer straightforward, readable code over clever but hard-to-understand solutions.  

Additional style or formatting rules may be introduced over time; please check existing code for guidance when in doubt.

## <a name="submit"></a> Submission Guidelines

All GitHub issues, discussions, and pull requests must be written in **English**.
Submissions in other languages may be closed or returned with a request to translate them before review.

Please keep in mind that Ikemen GO is developed by volunteers. Contributors choose what to work on based on availability, interest, and project priorities, so issues, discussions, and pull requests do not guarantee acceptance, prioritization, resolution, or a specific response timeline.

Maintainers may close topics that are out of scope, inactive, duplicated, not reproducible, already answered, or not planned.

### <a name="submit-issue"></a> Submitting an Issue

Before submitting an issue, please check the issue tracker to see if it has already been reported. Existing reports may also contain useful workarounds.

Test your content with the [nightly development release][nightly] and the [latest release][latest], as the problem may already be fixed. Note that only the most recent release is supported.

To resolve bugs, we must be able to reproduce them. A minimal reproduction is essential:
* Include a link to the problematic content.
* Provide clear steps to trigger the bug.
* When possible, use resources included with the engine (e.g. kfm/kfmz character or default screenpack) to maximize reproducibility.

A good minimal reproduction helps us quickly confirm whether an issue is a bug or a coding error, ensures we address the correct problem, and saves valuable development time. Issues without enough information to reproduce the problem cannot be addressed and will be closed.

To create a new issue, please choose from our [new issue templates][templates] and complete the relevant template.

### <a name="submit-pr"></a> Submitting a Pull Request (PR)

Before you submit your Pull Request (PR), please follow these guidelines:

1. Check [GitHub][pulls] for existing PRs that may be similar to your submission.
2. Ensure there's an issue that describes your fix or the feature you're adding. Design discussions should happen before starting your work.
3. [Fork][fork] the Ikemen GO repository.
4. In your fork, create a branch from the appropriate base branch.
   - Normal features and fixes target `develop`:
     ```shell
     git checkout -b my-branch-name develop
     ```
   - Approved backports and release-specific fixes target the applicable release branch:
     ```shell
     git checkout -b my-branch-name release/1.0
     ```

5. Make your changes and commit them:
   
   ```shell
   git commit --all
   ```

6. Push your branch to GitHub:
   
   ```shell
   git push origin my-branch-name
   ```

7. Submit the PR to the branch from which your work was created.

#### Contributor Responsibilities

Regardless of whether a contribution is self-written or AI-assisted, contributors must understand the code they submit and ensure it is testable and maintainable.

Contributors are also expected to help investigate and resolve bugs or regressions introduced by their changes, including those discovered after the PR has been merged. Changes that introduce unresolved bugs or regressions may be reverted to preserve the stability of the codebase.

#### Pull Request Review

All PRs are subject to review by the Ikemen GO dev team, which retains the right to decline any contributions.

#### Addressing Review Feedback

If changes are requested:

1. Make the required updates.
2. Commit the changes and push them to update your PR.

##### Updating the Commit Message

To update the commit message of the last commit:

1. Check out your branch and amend the commit message.
2. Force push to your repository to update the PR.

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
  └─⫸ Type: build|chore|docs|feat|fix|other|perf|refactor|style|test
```

### <a name="pr-type"> Type

The `<type>` portion of the title must be one of the following:
- **build**: Changes that affect the build system, external dependencies, CI configuration
- **chore**: Routine maintenance that does not change engine behavior
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
