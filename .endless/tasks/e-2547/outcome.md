# Synthesis: how go-fsfix creates real git repository fixtures

Brainstorm with Mike, 2026-11-03 to 2026-11-06. The plan on this task listed eight open questions; the conversation answered all eight and reshaped several of them. Everything below is Mike's call unless marked as accepted on my advice.

## Decisions Mike made

1. **RepoFixture always creates a real repository.** No opt-in flag, no NoGit escape. Mike's reasoning: no test calls AddRepoFixture without wanting a repository, and a NoGit RepoFixture is just a DirFixture. A test that wants only a `.git` marker nests a `.git` DirFixture in a DirFixture. The plan's Q1 (nested GitArgs pointer, flat bool, separate constructor, NoGit) is moot.

2. **Args stay flat.** Mike's principle: `*Args` structs hold fields, not nested option objects. `RepoFixtureArgs` gains `Branch`, `Message`, `Author` and `Committer` beside its existing fields. Branch and message are facts about the repo that the call adding it configures.

3. **The identity type is `dt.Mailbox` in go-dt, not a go-gitutil-local type.** The plan had coined "Identity"; Mike rejected a package-local type with a generic name and chose the third option: name what the thing is and put it in go-dt for every package to share. It is RFC 5322's mailbox, a display name plus an email address. Mike chose a struct over a string because a mailbox has two distinct values, and chose `dt.EmailAddress` as its own type because an email address is a subset of string, so a function can take one and either validate it or rightly assume it is valid. Constructor `NewMailbox`, parsers `ParseMailbox` and `ParseEmailAddress`, accessors `LocalPart()` and `Domain()` on EmailAddress (LocalPart, not Name, so it does not collide with Mailbox.Name). net/mail's Address is not used directly because its String() renders RFC 5322 quoting and RFC 2047 encoding, which git would store literally; net/mail remains the parser. go-gitutil names its field or parameter `Ident`, of type `dt.Mailbox`.

4. **Defaults are exported constants, composed.** `DefaultBranch`, `DefaultCommitMessage`, `DefaultAuthorName`, `DefaultAuthorEmail`, and `var DefaultAuthor = dt.NewMailbox(DefaultAuthorName, DefaultAuthorEmail)`. A test overrides any of them through the args.

5. **`RepoFixture.Repo()` returns the `*gitutil.Repo` go-fsfix used.** Mike accepted option A once the cost was reduced to one accessor: with Mailbox in go-dt, the return type of `Repo()` is the only go-gitutil type in go-fsfix's exported API, and go-gitutil is already in every consumer's build.

6. **`AddClone` and `AddWorktree` on RepoFixture, delegating to go-gitutil.** Mike proposed them; they replace the plan's "three sibling DirFixtures and the test calls go-gitutil". A clone or worktree is a sibling of its source on disk (inside the working tree it would be an embedded repository). Appending it to the source's parent gives the right creation order for free since children create in the order added. A linked worktree has a `.git` file, not a directory, so it gets its own `WorktreeFixture` type. This grows E-2548's scope to clone and worktree-add; recorded in E-2548's context.

7. **Nested RepoFixtures are allowed, not refused.** Mike asked what refusing would buy; nothing, once the walk order is considered: the inner repo commits first, and the outer forced add records it as an embedded repository (a gitlink, a submodule minus `.gitmodules`). go-gitutil's add sets `advice.addEmbeddedRepo=false`. A proper `AddSubmodule` waits until a test needs one; Mike avoids submodules but expects to meet repos that use them.

8. **`Reconcile` instead of new write verbs.** Mike proposed it in place of the plan's `WriteFile`/`Write` and the late-`AddFileFixture` fatal. `Reconcile(t)` walks the tree and rewrites every declared file so disk matches the declaration: missing files appear, modified ones revert, declarations added after Create get created. It can run at the top of a loop iteration. It never commits, so a file added after Create is untracked and a committed file rewritten after Create shows as modified, the two states gomion's tests need. Per ED-1550 this is part of the one implementation task, not a separate mutation task.

9. **GoModFixture lives in go-fsfix, not beside the go.mod code.** Mike reversed the plan's placement: a fixture is test code, go-goutil (E-2557's extraction) is production code and should carry no test helpers, and test code pulling a dependency is the cheap direction. Recorded as decision ED-1651 (proposed) and filed as E-2564, blocked by E-2557 and E-2568. Whether any part of a go.mod description earns a go-dt type is E-2557's question; Mike doubts `ModulePath` since its values are DirPath-shaped. The two rules that distinguish a module path (dot in the first element for fetchability, `/vN` suffix matching the major version) are recorded in E-2557's context.

10. **go-fsfix does not integrate go-cfgstore, for now.** Mike had been open to it and noted go-fsfix may import go-cfgstore even though go-cfgstore's test module imports go-fsfix. Convinced by the arithmetic: the integrated form saves zero lines over a DirFixture plus go-cfgstore's own DirsProvider helper. Revisit if a richer case appears (seeded config files in layout).

11. **The release chain is accepted:** go-dt (E-2567), then go-gitutil (E-2548's implementation, with its replace directives gone and go-diffutils tagged), then go-fsfix (E-2568), then gomion. Mike: deliberating over ordering costs more than using blocks and doing it; streamlining chained landings is filed elsewhere. Local development uses go.work, so only tagging is serial.

## Accepted on my advice

- **Every declared file is committed in one commit at Create.** Mike accepted without claiming to understand the ramifications; they are: the repo is clean after Create with one commit; a test needing untracked or modified files makes them after Create via AddFileFixture plus Reconcile; the add is forced so a declared file an ignore rule would skip is tracked anyway; further commits go through `Repo()`. The alternative, a per-file `Uncommitted` flag, can be added later if a test wants a mixed starting state in one declaration.
- **go-fsfix requires the whole go-gitutil module** rather than asking E-2548 for a nested module. x/mod comes with it and is wanted for E-2564 anyway.
- **go-fsfix's own tests verify through go-gitutil's read side** (branch, log, status) plus one assertion needing no git: `GitPath()` contains HEAD and objects. No git invocation from a test file.
- **`AddFiles(t, map[dt.RelFilepath]string)`** as a method beside the builder calls, on RootFixture, DirFixture and RepoFixture. Keys are file paths; directories come from the keys.

## Things that stay out
- go-fsfix never shells out to git itself and learns nothing about go.mod generation or go-cfgstore layouts.
- E-2548 does not block this brainstorm; it blocks E-2568.

## Follow-ups spawned
- E-2567 (go-dt): Add EmailAddress and Mailbox types. Cleans up E-2547.
- E-2568 (go-fsfix): Make RepoFixture create a real git repository. The design above is its analysis. Blocked by E-2567 and E-2548; blocks E-2540 and E-2559; cleans up E-2547. When E-2548 spawns its implementation task, re-point E-2568's blocked-by at that task.
- E-2564 (go-fsfix): Add GoModFixture. Blocked by E-2557 and E-2568; cleans up E-2547; implements ED-1651.
- ED-1651 (proposed): Fixtures live in go-fsfix and depend on production packages, never the reverse.
- E-2548's context now carries the enlarged scope (init with branch, forced add-all, commit with Ident, clone, worktree-add, advice.addEmbeddedRepo=false, Repo usable from init).
- E-2557's context now carries the ModulePath question with examples.
