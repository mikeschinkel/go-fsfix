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
        Branch       dt.GitBranchName // "" means DefaultBranch
        Message      string           // "" means DefaultCommitMessage
        Author       dt.Mailbox       // zero means DefaultAuthor
        Committer    dt.Mailbox       // zero means Author
        CommitTime   time.Time        // zero means DefaultCommitTime
    }

## Exported defaults, composed
    const (
        DefaultBranch        dt.GitBranchName    = "main"
        DefaultCommitMessage                     = "fsfix: initial commit"
        DefaultAuthorName    dt.EmailDisplayName = "go-fsfix"
        DefaultAuthorEmail   dt.EmailAddress     = "fsfix@example.com"
    )
    var DefaultAuthor = dt.Mailbox{DisplayName: DefaultAuthorName, Email: DefaultAuthorEmail} // go-dt has no NewMailbox (E-2567)
    var DefaultCommitTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

## Repo accessor
- RepoFixture keeps the *gitutil.Repo it used and returns it from `Repo()`, guarded by ensureCreated. go-gitutil is already in the consumer's build, so the accessor adds no dependency.

## AddClone and AddWorktree
- `origin.AddClone(t, name, &CloneFixtureArgs{...}) *CloneFixture`: CloneFixture embeds *RepoFixture, so a clone gets RepoFixture's methods and every *Args type stays named after the fixture type it builds (Mike's choice over AddCloneArgs, which would follow go-gitutil's verb-named style but split go-fsfix's convention). A clone is a sibling of the origin on disk (a clone inside the working tree would be an embedded repo). Implementation appends the clone to the origin's parent's ChildFixtures; since children create in order and the origin was added first, the clone runs after the origin's commit with no special ordering code.
- `repo.AddWorktree(t, name, &WorktreeFixtureArgs{Branch}) *WorktreeFixture`: same sibling placement; a linked worktree has a `.git` file rather than a directory, so it is its own small type.
- Both delegate to go-gitutil's clone and worktree-add.

## AddFiles map form
- `AddFiles(t, map[dt.RelFilepath]string)` on RootFixture, DirFixture and RepoFixture, beside AddFileFixtures. Keys are file paths; directories come from the keys (createFile already runs MkdirAll). An empty directory still needs AddDirFixture. Builder calls stay for permissions, times and ContentFunc.

## Reconcile instead of new write verbs
- Mike proposed it in E-2547: "maybe we need a Reconcile() method ... that can be called later to fixup anything still needing fixed up, or even called before each iteration of a loop during testing?"
- `Reconcile(t)` on RootFixture (and the other fixtures) walks the tree and rewrites every declared file so disk matches the declaration: missing files appear, modified ones revert, declarations added after Create get created.
- Reconcile never commits and never runs git. Git state is set once at Create, so a file added after Create is untracked and a committed file rewritten after Create shows as modified, the two states gomion's tests need.
- The silent case today (AddFileFixture after Create appends to FileFixtures but writes nothing, dir_fixture.go) is resolved by Reconcile, not by a fatal.

## Dependency and release
- go-fsfix requires the whole go-gitutil module; x/mod comes with it. go directive rises to 1.27.0; go-dt to the release that has E-2567 and E-2599.
- Release order: go-dt (E-2567, E-2599), go-gitutil (E-2600, no replace directives), then go-fsfix. Local development uses go.work.

## Tests and docs
- go-fsfix's test module verifies through go-gitutil's read side (branch, log, status) plus one assertion needing no git: GitPath() contains HEAD and objects. No git invocation from a test file.
- README gains a real-repo section, documents Create versus Reconcile, and drops the "simulate" wording.

# Decided while planning (2026-11-12, with Mike)

## Fixture git ignores the machine's git config (Mike: C)
- A machine's global or system config can change or break fixture commits: `commit.gpgsign=true` tries to sign (fails or prompts without a key or agent), `core.hooksPath` runs global hooks, `core.autocrlf` rewrites line endings, `init.templateDir` copies hooks into each new repo.
- Fixture creation runs git with `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_NOSYSTEM=1`, so the same declaration gives the same repo on every machine.
- The fixture repo's own config also gets `commit.gpgsign=false`, an empty `core.hooksPath` and `core.autocrlf=false`, so code under test that commits inside a fixture behaves the same on every machine too. Cost accepted: a test of config-reading behaviour sees those entries.
- Rejected: inheriting the machine's config (fixtures differ by machine); isolating creation only (code under test still hits the machine's signing and hooks).

## Commit dates are fixed (Mike: B)
- `DefaultCommitTime`, overridable by `RepoFixtureArgs.CommitTime`, applied as both author and committer date. The same declaration gives the same commit hash everywhere, so tests can assert hashes and keep golden output.
- Cost accepted: code under test that compares commit age to now sees an old commit unless the test sets CommitTime.
- Rejected: the current time (hashes change per run); reusing ModifiedTime (it already means the directory's file time).

## Reconcile can remove undeclared files, opt-in (Mike: C)
- `ReconcileArgs.RemoveUndeclared` deletes files on disk that no FileFixture declares, never touching `.git` or a nested repo's `.git`. Without it, Reconcile deletes nothing, so files a test wrote directly stay for inspection.
- Rejected: never removing (an iteration inherits the previous one's leftovers); always removing (deletes files a test wrote on purpose).

## A clean git state comes from recreating, not from reset (Mike: B)
- My first framing of this question offered only "disk only" or "reset and checkout inside Reconcile"; it omitted wipe-and-recreate. Mike noted tests should not depend on a prior test's state and asked whether recreating could take seconds.
- Measured on git 2.52 with global and system config isolated, init + forced add + commit: 41 ms for 10 files, 59 ms for 100, 95 ms for 1,000, 412 ms for 10,000. A loop pays this per iteration; reset plus checkout would also start git processes, so the saving for small fixtures is about half, not an order of magnitude. Windows process start is slower and unmeasured; it scales both alike.
- Mike: "If no, then B definitely." So a test needing a clean git state calls Cleanup and then Create again (a new temp directory, a fresh repo), or builds a fixture per iteration. Recreation restores everything (branches, tags, worktrees, untracked files), which reset and checkout would not, and go-gitutil needs no reset or checkout for fixtures.

## Branch type is dt.GitBranchName (Mike)
- go-gitutil's GitRef is used where git means a revision, broader than a branch name (`HEAD~1`, `refs/tags/v1`). Mike asked for git's concepts to be fully modelled and reused across packages, in go-dt's root package, each name prefixed with Git: E-2599 (GitRef, GitBranchName, GitTagName, GitRemoteName, GitRevision, GitObjectID). Name is dropped only from GitRef, since a ref already is a name; a branch, an annotated tag and a remote are each more than their name. A second domain-types package was rejected because developers would have to discover it and choose between it and go-dt.
- WorktreeFixtureArgs.Branch and CloneFixtureArgs.Branch use the same type.

## Decided without asking (Mike could object; none did)
- `AddClone` takes its own `CloneFixtureArgs` and returns a `*CloneFixture` (Branch to check out, Permissions, ModifiedTime). RepoFixtureArgs' Files, Message, Author, Committer and CommitTime mean nothing for a clone.
- A missing git binary fails the test (`t.Fatal`), never skips it; a skipped suite can pass while testing nothing. Mike confirmed (Q7: A).
- A worktree's branch defaults to the fixture's name, matching `git worktree add <path>`.
- A git failure during Create calls `t.Fatal`: later steps depend on it.
- Author and Committer stay separate; git sets both through environment variables or the `author.*`/`committer.*` config keys.
- Items state behaviour; go-gitutil's call names come from E-2600.

## Args stay pointers; constructors stop writing into them (Mike: A plus the fix)
- Every Add* method takes `*…FixtureArgs`, matching the existing three, so callers may pass nil. Nil and an empty struct mean the same, since every zero field means its default.
- Today newRepoFixture and newDirFixture write defaults (Permissions = 0755) back into the caller's struct, so a test reusing one args value sees it changed. Constructors now copy the args before filling defaults. This fixes the existing three without breaking callers; switching all five to values was rejected because it breaks every caller passing nil or &…Args{}.
