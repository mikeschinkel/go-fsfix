# Open questions: how go-fsfix creates real git repository fixtures

Written 2026-11-03 from a survey of go-fsfix, go-gitutil and the E-2548 plan. The requirements are in this task's analysis (from gomion's E-2546). The questions stay open for Mike to answer; nothing is decided except the "Decided without asking" list at the end. Where an option depends on an E-2548 answer, the question says so.

Terms:

- **empty-.git fixture**: what RepoFixture produces today: a directory tree plus an empty `.git` directory. `git rev-parse` fails in it.
- **real repo fixture**: a RepoFixture whose directory is a git repository created by go-gitutil: `init`, forced add of every declared file, one commit.
- **declared file**: a FileFixture added to a fixture before `RootFixture.Create` runs.
- **Repo**: go-gitutil's type for an opened repository (E-2548 Q1 decides how it is constructed).
- **identity**: author and committer name and email recorded on a commit.

## What the survey found

- go-fsfix has four fixture types: RootFixture (temp dir), DirFixture, RepoFixture (embeds DirFixture, adds the empty `.git`), FileFixture. Every `Add*` call records a child; nothing touches disk until `RootFixture.Create`, which walks the tree: a fixture creates its directory, then its declared files, then its children.
- Nothing outside go-fsfix uses RepoFixture. gomion's test module has a `replace` for go-fsfix but no `require`; go-fsfix's own test module and example are the only callers.
- FileFixture names may contain directories (`createFile` runs MkdirAll), so the map form is syntax over what exists.
- After `Create`, `AddFileFixture` still only records; the caller must call `FileFixture.Create` to write it. The README does not say so.
- go-fsfix requires go-dt v0.3.3 and declares go 1.25.3. go-gitutil requires go-dt v0.6.1, go-dt/dtx, go-diffutils and golang.org/x/mod, declares go 1.27.0, and its go.mod carries `replace` directives for the go-dt and go-diffutils siblings; go-diffutils has no release tag (required at v0.0.0). Taking the dependency raises go-fsfix's go-dt floor by three minor versions and its go directive, and go-gitutil cannot be required from the proxy until its replaces are gone and go-diffutils is tagged.
- go-gitutil's run paths, Streamer and Repo construction are E-2548's questions; this plan assumes only that go-gitutil will offer init-with-branch, forced add-all, and commit-with-identity, each taking a context and a directory.

## Q1. How does a RepoFixture opt into being a real repo fixture?

Mike said "an option on RepoFixtureArgs, or a WithArgs constructor". The empty-.git behaviour stays available.

**A. A nested args value**: `RepoFixtureArgs.Git *GitArgs` where `GitArgs{Branch, Identity, Message}`; nil keeps the empty-.git behaviour, a non-nil pointer (even `&GitArgs{}`) means real repo with the constants filled in.
- Pros: every git-only field sits inside the thing that turns git on, so none is meaningful-but-ignored on an empty-.git fixture; `&fsfix.GitArgs{}` reads as "real repo, defaults".
- Cons: a pointer whose nil-ness carries meaning; a caller who wants the defaults writes an empty literal rather than `true`.

**B. Flat fields**: `RepoFixtureArgs{Git bool, Branch, Identity, Message}`.
- Pros: `Git: true` is the smallest opt-in.
- Cons: Branch, Identity and Message are silently ignored when Git is false; nothing stops a fixture that sets Branch and forgets Git.

**C. A separate constructor**: `AddGitRepoFixture(t, name, *GitRepoFixtureArgs)` beside `AddRepoFixture`, same RepoFixture type returned.
- Pros: the call site names what it builds; args types cannot mix git and non-git fields.
- Cons: three `Add*Repo*` methods on each of RootFixture, DirFixture and RepoFixture (six new methods) for one bool of difference.

**D. Real repo becomes the default**: RepoFixture always runs git; `RepoFixtureArgs.NoGit bool` restores the empty `.git`.
- Pros: a type named RepoFixture produces a repository without being asked; the empty-.git case is the odd one and is spelled out.
- Cons: every RepoFixture pays init-add-commit (tens of milliseconds and a git binary) including fixtures that only want a `.git` marker for a repo-root search; go-fsfix's own tests and README change meaning.

Recommendation: A.

## Q2. Which declared files does the fixture commit?

gomion's tests want to add an untracked file and modify a committed file after creation; both work when every declared file is committed. The question is whether a declared file can be left uncommitted.

**A. All declared files, one commit.** Untracked or modified files are made after creation (Q6).
- Pros: one rule; the fixture tree on disk equals the commit.
- Cons: a test wanting "repo with one untracked file" declares the tracked files, creates, then adds the other file in a second step rather than in one declaration.

**B. Per-file opt-out**: `FileFixtureArgs.Uncommitted bool`; the fixture adds everything else, forced, and commits.
- Pros: the whole starting state, tracked and untracked, is one declaration.
- Cons: `git add --all --force` cannot exclude paths, so the fixture must enumerate the files to add instead (the forced-add operation takes explicit paths), and a FileFixture gains a field that means nothing outside a real repo fixture.

**C. Fixture-level `Commit bool` in GitArgs**: false means init only, every file untracked.
- Pros: covers "fresh repo, nothing committed" in one declaration.
- Cons: cannot express a mix; gomion has no test needing an empty repository with untracked files.

Recommendation: A.

## Q3. How much of go-gitutil shows in go-fsfix's exported API?

**A. go-gitutil types in the API**: `GitArgs.Identity` is go-gitutil's identity type; `RepoFixture.Repo()` returns the Repo go-fsfix used, so the test can commit, branch, clone or inspect without reopening.
- Pros: the test and the fixture share one Repo; no second `rev-parse`; one definition of identity across both packages.
- Cons: go-gitutil becomes part of go-fsfix's public surface, so a breaking go-gitutil release is a breaking go-fsfix release; the `Repo()` method must guard against use before `Create` (the existing `ensureCreated` guard).

**B. Plain fields only**: `GitArgs{Branch string, AuthorName, AuthorEmail string, Message string}`; no accessor. go-gitutil is an implementation detail; a test wanting a Repo opens one on `rf.Dir()`.
- Pros: go-gitutil can change freely behind go-fsfix; a reader of go-fsfix's API needs no second package.
- Cons: identity is defined twice (go-fsfix's fields and go-gitutil's type) and go-fsfix converts; every test that goes on to run git opens a Repo that go-fsfix already had.

Recommendation: A. Mike accepted the dependency; hiding it buys a conversion and a reopen per test.

## Q4. What happens when a real repo fixture contains another RepoFixture?

The creation walk is directory, files, children. A RepoFixture child of a real repo fixture is created inside the outer working tree. gomion's shapes do not nest repositories (nested modules are directories; origin, clone and worktree are siblings).

**A. Commit the outer repo after its children are created.** A nested empty-.git child is skipped by git (git never descends into a directory named `.git`), so its files commit into the outer repo; a nested real repo child becomes an embedded repository entry (git warns) and its files do not.
- Pros: no restriction; the walk order stays as it is.
- Cons: the two nested kinds produce different outer commits, and git's embedded-repository warning lands in test output.

**B. Commit the outer repo before its children are created**: insert the git step between files and children.
- Pros: child content is never in the outer commit, whatever kind of child it is.
- Cons: a DirFixture child's files are untracked in the outer repo, which contradicts "every declared file is committed" (Q2 A) for the nested-module layout gomion needs, unless the git step distinguishes DirFixture children from RepoFixture children.

**C. Refuse**: `t.Fatalf` when a real repo fixture has a RepoFixture descendant (either kind).
- Pros: the one layout with ambiguous meaning cannot be declared by accident; A or B can be chosen later when a test needs it.
- Cons: a future test of nested repositories (gomion's `.endless/worktrees` layout holds linked worktrees inside the main checkout) has to change go-fsfix first.

Recommendation: C.

## Q5. Does go-fsfix depend on all of go-gitutil?

go-gitutil carries diff parsing, status, staging and worktree code and depends on go-diffutils, go-dt/dtx and golang.org/x/mod. go-fsfix needs init, add and commit.

**A. Whole module.**
- Pros: one `require`; no change to go-gitutil's layout.
- Cons: go-fsfix, a test helper, takes go-diffutils and x/mod into every test module that imports it; go-fsfix cannot release until go-diffutils is tagged and go-gitutil's replaces are removed (an E-2548 prerequisite either way, but A makes go-fsfix wait on all of it).

**B. go-gitutil publishes the run, init, add, commit and clone operations as a nested module** (a `go.mod` in a subdirectory, as go-dt does with appinfo and dtx) that depends only on go-dt; the main module imports it.
- Pros: go-fsfix's footprint is go-dt plus the small module; the operations gomion's non-test code also needs are separable from diff parsing.
- Cons: go-gitutil becomes two modules to tag and version; Repo (E-2548 Q1) must live in the small module, so the main module's methods hang off a type defined elsewhere.

Recommendation: A, and revisit if the footprint ever bites. The split is E-2548's to make; asking here records that go-fsfix would use it.

## Q6. How does a test change a fixture after creation?

Needed: add an untracked file; modify a committed file. Today `AddFileFixture` after `Create` records without writing, and `FileFixture.Create(t, parent)` writes.

**A. `AddFileFixture` writes immediately when the parent is already created**; before creation it records as today. Modifying a committed file is `AddFileFixture` with the same name and new content.
- Pros: one call does the right thing in either phase; no new method; the FileFixture list stays the record of what exists.
- Cons: the same call has two behaviours by fixture state; re-adding a name already in the list leaves two entries for one path.

**B. Keep today's two-step** (`AddFileFixture` then `FileFixture.Create`) and document it.
- Pros: no behaviour change.
- Cons: every post-creation write is two calls, and forgetting the second one fails silently (the file is recorded but absent).

**C. Distinct methods on created fixtures**: `WriteFile(t, name, content) dt.Filepath` and, on FileFixture, `Write(t, content)`. `AddFileFixture` after creation becomes `t.Fatalf`.
- Pros: declaration and mutation are different verbs, so a test reads as "declare, create, then change"; calling the wrong one in the wrong phase fails loudly.
- Cons: a second way to put a file in a fixture, with no FileFixture record for files written this way (unless WriteFile also records).

Recommendation: C, with `WriteFile` recording a FileFixture so the list stays complete.

## Q7. Where does the map form attach?

Mike: a map from relative path to content, expanding to directories and files, beside the builder calls which stay for permissions, times and ContentFunc.

**A. A field on each `*FixtureArgs`**: `DirFixtureArgs.Contents map[dt.RelFilepath]string`, same on RepoFixtureArgs; RootFixture gets the same through a new `RootFixtureArgs` or a method.
- Pros: a fixture and its files are one literal; the common case needs no builder call.
- Cons: each args struct then holds two ways to declare files (`Files []*FileFixture` and the map), and RootFixture has no args struct today.

**B. A method `AddFiles(t, map[dt.RelFilepath]string)`** on RootFixture, DirFixture and RepoFixture, beside `AddFileFixtures`.
- Pros: parallel to the existing builder methods; works for RootFixture without a new args type; callable before creation (records) and, under Q6 A, after.
- Cons: a fixture with files is at least two statements.

**C. Extend `AddFileFixtures`'s variadic `any`** to accept a map as one more case.
- Pros: no new name.
- Cons: `AddFileFixtures(t, nil, map...)` passes a defaults argument the map form never uses, and the type switch already takes three kinds.

Recommendation: B.

## Q8. How does go-fsfix's own test verify a real repo fixture?

Mike: git belongs in go-gitutil, not in test files; never mock git.

**A. Through go-gitutil's read operations** (Repo, branch, log or status) in go-fsfix's test module.
- Pros: no git invocation in a test file; the same package that wrote the fixture reads it back.
- Cons: a defect shared by go-gitutil's write and read paths passes; the test proves consistency, not that `git` itself sees a commit.

**B. By running `git` from the test module** (`git log`, `git status --porcelain`).
- Pros: independent of go-gitutil, so it catches a defect on both sides of it.
- Cons: test-file git invocations, which Mike ruled out of gomion; go-fsfix's test module would be the one place in the ecosystem that shells out to git outside go-gitutil.

Recommendation: A, plus one assertion that needs no git: `rf.GitPath()` contains `HEAD` and `objects`.

## Decided without asking

- Defaults are exported constants in go-fsfix: branch `main`, author and committer `go-fsfix <fsfix@example.com>`, commit message `fsfix: initial commit`. go-gitutil holds no fixture defaults (E-2548 Q5 recommendation A).
- A real repo fixture runs git once, after its directory and declared files exist, with every declared file added forced so ignore rules cannot drop one.
- Map keys are file paths; directories come from the keys (MkdirAll). An empty directory still needs `AddDirFixture`.
- The origin, clone and linked-worktree shape is three sibling fixtures: one real repo fixture and two empty DirFixtures as clone and worktree targets; the test calls go-gitutil for clone and worktree-add. go-fsfix adds no API for it.
- User-config isolation is a plain DirFixture; go-fsfix learns nothing about go-cfgstore (from E-2546).
- go-fsfix's go directive rises to 1.27 and its go-dt requirement to whatever go-gitutil's release requires; its test module follows.
- Prerequisite, not go-fsfix's to do: go-gitutil releases the fixture operations with no `replace` directives in its go.mod (go-diffutils tagged first). go-fsfix cannot be released before that.
- The README gains a real repo fixture section and corrects the "simulation" wording.
