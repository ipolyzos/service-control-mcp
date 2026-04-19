# Branching and Release Workflow

## Branch topology

```
main          ← stable, tagged releases only
  └── develop ← integration branch; all PRs target here
        ├── feature/<topic>
        ├── fix/<topic>
        └── chore/<topic>
```

`main` is always at a tagged release commit. `develop` is always green
on CI. No one commits directly to either.

## Day-to-day development

```sh
# Start work
git checkout -b feature/my-topic develop

# ... implement, test, commit ...

# Open PR targeting develop
gh pr create --base develop
```

Every PR must:
- Pass all required CI checks (build, test, lint, coverage gate)
- Have at least one commit with `Signed-off-by`
- Include a `CHANGELOG.md` entry under `[Unreleased]` for user-visible changes

## Release flow

Releases are cut from `develop` via a version tag on `main`.

```sh
# 1. Ensure develop CI is green and CHANGELOG is updated
# 2. Fast-forward main to develop
git checkout main
git merge --ff-only origin/develop
git push origin main

# 3. Tag the release (matches GITHUB_REF_NAME in the release workflow)
git tag v<major>.<minor>.<patch> -m "Release v<major>.<minor>.<patch>"
git push origin v<major>.<minor>.<patch>
```

Pushing a `v*.*.*` tag triggers the release workflow, which:
1. Runs the full test suite
2. Cross-compiles binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`
3. Generates `sha256sums.txt`
4. Creates a GitHub Release with all artifacts attached

## Versioning policy

This project follows [Semantic Versioning 2.0](https://semver.org/):

| Change type | Version bump |
|---|---|
| Backward-compatible new tool or option | `MINOR` |
| Bug fix with no interface change | `PATCH` |
| Breaking change to tool input/output schema | `MAJOR` |
| Breaking change to plugin interface | `MAJOR` |

Pre-release identifiers (`-alpha.1`, `-beta.1`, `-rc.1`) in the tag name
cause the release workflow to mark the GitHub Release as a pre-release.

## Hotfix flow

For critical fixes against a released version:

```sh
git checkout -b fix/critical-issue main
# ... fix with reproducing test first ...
git push origin fix/critical-issue
gh pr create --base main   # exceptional; document reason in PR
# After merge to main, back-port to develop
git checkout develop
git cherry-pick <sha>
git push origin develop
```

Hotfixes that go directly to `main` must be back-ported to `develop`
before the next regular release.
