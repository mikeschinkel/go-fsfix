# Design settled in E-2547

## RepoFixture always creates a real repository
- No opt-in flag and no NoGit escape. A test that wants only a `.git` marker builds a DirFixture with a `.git` DirFixture child.
- Creation order stays directory, declared files, children; then one git step: init with branch, forced add of every file, one commit. Every declared file is committed; ignore rules cannot drop one.
- After Create the repo is clean with exactly one commit. A test wanting untracked or modified files makes them after Create (see Reconcile) and further commits through Repo().
- RepoFixture nested inside a RepoFixture is allowed. The inner repo commits first (walk order); the outer forced add records it as an embedded repository (a gitlink). go-gitutil's add sets advice.addEmbeddedRepo=false. A proper AddSubmodule with .gitmodules waits until a test needs it.

## Args are flat (Mike's principle)
    type RepoFixtureArgs struct {
        Files        []*FileFixture
        Permissions  int
        ModifiedTime time.Time
        Branch       string      // "" means DefaultBranch
        Message      string      // "" means DefaultCommitMessage
        Author       dt.Mailbox  // zero means DefaultAuthor
        Committer    dt.Mailbox  // zero means Author
    }

## Exported defaults, composed
    const (
        DefaultBranch        = "main"
        DefaultCommitMessage = "fsfix: initial commit"
        DefaultAuthorName    = "go-fsfix"
        DefaultAuthorEmail   dt.EmailAddress = "fsfix@example.com"
    )
    var DefaultAuthor = dt.NewMailbox(DefaultAuthorName, DefaultAuthorEmail)

## Repo accessor
- RepoFixture keeps the *gitutil.Repo it used and returns it from `Repo()`, guarded by ensureCreated. go-gitutil is already in the consumer's build, so the accessor adds no dependency.

## AddClone and AddWorktree
- `origin.AddClone(t, name, args) *RepoFixture`: a clone is a sibling of the origin on disk (a clone inside the working tree would be an embedded repo). Implementation appends the clone to the origin's parent's ChildFixtures; since children create in order and the origin was added first, the clone runs after the origin's commit with no special ordering code.
- `repo.AddWorktree(t, name, &WorktreeFixtureArgs{Branch}) *WorktreeFixture`: same sibling placement; a linked worktree has a `.git` file rather than a directory, so it is its own small type.
- Both delegate to go-gitutil's clone and worktree-add.

## AddFiles map form
- `AddFiles(t, map[dt.RelFilepath]string)` on RootFixture, DirFixture and RepoFixture, beside AddFileFixtures. Keys are file paths; directories come from the keys (createFile already runs MkdirAll). An empty directory still needs AddDirFixture. Builder calls stay for permissions, times and ContentFunc.

## Reconcile instead of new write verbs
- `Reconcile(t)` on RootFixture (and the other fixtures) walks the tree and rewrites every declared file so disk matches the declaration: missing files appear, modified ones revert, declarations added after Create get created. Useful at the top of a loop iteration.
- Reconcile never commits. Git state is set once at Create, so a file added after Create is untracked and a committed file rewritten after Create shows as modified, the two states gomion's tests need.
- The silent case today (AddFileFixture after Create appends to FileFixtures but writes nothing, dir_fixture.go) is resolved by Reconcile, not by a fatal.

## Dependency and release
- go-fsfix requires the whole go-gitutil module; x/mod comes with it. go directive rises to 1.27.0; go-dt to the release that has E-2567.
- Release order: go-dt (E-2567), go-gitutil (E-2548's implementation, no replace directives), then go-fsfix. Local development uses go.work.

## Tests and docs
- go-fsfix's test module verifies through go-gitutil's read side (branch, log, status) plus one assertion needing no git: GitPath() contains HEAD and objects. No git invocation from a test file.
- README gains a real-repo section, documents Create versus Reconcile, and drops the "simulate" wording.

## Not in this task
- GoModFixture: E-2564. go-cfgstore integration: none (zero lines saved). AddSubmodule: when needed.
