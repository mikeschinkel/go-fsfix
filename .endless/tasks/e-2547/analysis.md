# Requirements from gomion (E-2546, 2026-11-03)

Settled with Mike in the gomion brainstorm. Design is this task's; these are what gomion's tests need to express.

## Real repo fixture
- RepoFixture must be able to produce a real git repository: `git init`, add every declared file (forced, so paths such as `.endless/worktrees/a/go.mod` and `dtx/testdata/x/go.mod` are tracked even if an ignore rule would skip them), one commit.
- Opt-in through RepoFixtureArgs, which AddRepoFixture already takes (Mike: an option on RepoFixtureArgs, or a WithArgs constructor). The empty-`.git` behaviour stays available.
- Default branch name and author/committer identity are constants so every fixture is identical on every machine; the test may override them through the same args (Mike: "default to constants but allow the test to provide them").
- The git operations are go-gitutil's (E-2548). go-fsfix depends on go-gitutil; Mike accepted the dependency ("a dependency is better than large cases of duplication"). No shelling out from go-fsfix itself, no git interface for callers to fill in.

## Declaring contents
- Add a map form over the builder: relative path to content, expanding to directories and files. Builder calls (AddFileFixture with FileFixtureArgs) stay for permissions, times and ContentFunc. Mike: "EXACTLY".
- Note: FileFixture names already accept nested relative paths (createFile does MkdirAll), so the map form is a thin convenience, not new capability.

## After creation
- Tests write into a created fixture: add an untracked file, modify a committed file, then run the code under test. Adding a file to an already-created Dir/Repo fixture must work (today FileFixture.Create can be called directly; confirm that is the intended path).

## Shapes gomion needs
1. Repo with committed files, three layouts: root module plus nested modules; root only; nested modules only (no root go.mod). Content is go.mod files, a .go file, a workflow file, decoy go.mod files.
2. Origin repo, clone of it, linked worktree on a new branch. The fixture hands back directories; clone and worktree-add are go-gitutil calls the test makes (not go-fsfix API).
3. Plain directory trees of go.mod and go.sum files with no git (five gomion test files today).

## Not go-fsfix's
- User config directory isolation: a plain DirFixture; go-cfgstore gains the DirsProvider helpers. go-fsfix must not learn go-cfgstore's layout.
- go.mod content generation (gomion's goModFile helper): not a go-fsfix concern; filed separately.
