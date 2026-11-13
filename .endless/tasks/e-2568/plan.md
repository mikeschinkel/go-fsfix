# Approach

Rationale and rejected alternatives are in the analysis. This is how.

## Order of work
1. Wait for, or develop against through go.work, go-dt with E-2567 (Mailbox, EmailAddress, EmailDisplayName) and E-2599 (GitBranchName), and go-gitutil with E-2600 (NewRepo/Open, Init, Clone, Add, Commit, worktree add). The release has no replace directives in go-fsfix's root go.mod.
2. Baseline: go 1.27.0 in every go.mod, golangci-lint v2.15.0 clean, Endless per-checkout files gitignored.
3. Defaults and args, then RepoFixture's git step, then Repo(), then AddClone and AddWorktree, then AddFiles, then Reconcile, then recreate after Cleanup, then tests and README.

## Files
- `repo_fixture.go`: `RepoFixtureArgs` gains Branch, Message, Author, Committer, CommitTime. `createWithParent` drops the `.git` mkdir and its "shell out to git init" TODO. After the directory, the declared files and the children are made, it runs one git step through go-gitutil: init on the branch, set the repo-local neutralising config, forced add of everything, one commit with author, committer and date. It keeps the `*gitutil.Repo` for `Repo()`.
- New `defaults.go` (or beside the args): the exported Default* constants and vars.
- New `git_env.go` or equivalent: the one place that builds the isolation environment (`GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1`) and the author/committer identity and dates handed to go-gitutil. go-fsfix never execs git itself; how the environment reaches git is E-2600's API.
- New `clone_fixture.go`: `CloneFixture` (embeds `*RepoFixture`), `CloneFixtureArgs`, `(*RepoFixture).AddClone`. The clone is appended to the origin's parent's ChildFixtures; its create step clones instead of init/add/commit.
- New `worktree_fixture.go`: `WorktreeFixture`, `WorktreeFixtureArgs`, `(*RepoFixture).AddWorktree`, appended to the origin's parent's ChildFixtures.
- `root_fixture.go`, `dir_fixture.go`, `repo_fixture.go`: `AddFiles(t, map[dt.RelFilepath]string)` on each, building FileFixtures from the keys.
- New `reconcile.go`: `ReconcileArgs{RemoveUndeclared bool}` and `Reconcile(t, *ReconcileArgs)` on RootFixture, DirFixture and RepoFixture (nil args means the defaults). It walks the declared tree, writes every declared file, and, when asked, deletes undeclared files, skipping every `.git` directory or file.
- `root_fixture.go`: `Create` after `Cleanup` makes a new temp directory and rebuilds the whole tree, including every repo, clone and worktree.
- `test/`: new tests per item, verifying through go-gitutil's read side; no test execs git.
- `README.md`: real-repo section, Create versus Reconcile versus recreate, the defaults and isolation, and drop the "simulate" wording.
- `.gitignore`: the Endless per-checkout entries.
