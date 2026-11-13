# Lessons Learned

A capture log of corrections, newest last. Entries are appended by
`endless lesson write`, which commits this file on the main checkout as it
writes — so recording a lesson never waits on a task landing, and never
dirties a worktree.

## Format

```
### [YYYY-MM-DD] One-line summary
<the lesson — what went wrong, why, and the rule that replaces it>
- **Project**: <project>
```

---

## Entries

### [2026-11-06] Say 'exporting symbols', not 'exposing them'
While brainstorming E-2547 I wrote 'exposing them' for go-gitutil types appearing in go-fsfix's exported API. Mike: say 'exporting symbols'; 'exposing' reads as a crime. Use Go's own word, export, when talking about public API surface.
- **Project**: go-fsfix

### [2026-11-06] Present a list of settled or open points as bullets, not a run-on paragraph
In E-2547 I closed a message with a paragraph listing a dozen settled points separated by semicolons. Mike: a run-on paragraph is much harder to scan than a bulleted list. When the content is parallel items (decisions, open questions, settled points), use bullets even if it costs a few lines.
- **Project**: go-fsfix

### [2026-11-09] A decision must keep the scope the user argued, not generalize it
In E-2547 I filed ED-1651 'Fixtures live in go-fsfix and depend on production packages, never the reverse' from Mike's argument that GoModFixture belongs in go-fsfix. Mike rejected it: he was arguing that go-fsfix integrates git because git fixtures are widely used, not that every production capability belongs in go-fsfix (an Oracle DB fixture would not, since most users do not need it), and the wording could be read as moving production code. Record the case the user made, with its reason, so the boundary shows; a decision broader than the argument gets cited as authority it never had.
- **Project**: go-fsfix

### [2026-11-11] Don't recommend a type that models a looser constraint as the fallback
Situation: planning E-2568 (go-fsfix RepoFixture), Q5 asked what type RepoFixtureArgs.Branch should be. I recommended a new branch-name type but offered go-gitutil's existing GitRef as a fallback, while admitting a ref is broader than a branch name. Mike objected that this proposes types that don't model reality as closely as we could. Reason: the go-dt principle is types that embrace the real constraint; a fallback to a looser type quietly abandons it. Next time: when an existing type is broader than the concept, propose fleshing out the correct type family (and ask where it should live) rather than offering the broader one as an acceptable fallback.
- **Project**: go-fsfix

### [2026-11-12] Include the discard-and-rebuild option when asking how to restore state
Situation: planning E-2568, I asked Mike whether RepoFixture.Reconcile should restore git state, offering only A (disk only) or B (reset and checkout to the initial commit). The scenario, a loop iteration whose code under test commits, was my own inference. I left out the obvious option, wiping the fixture and recreating it. Mike picked B, and it got relayed to E-2548 as a go-gitutil requirement. Mike then asked why, noting that tests should not depend on a prior test's state, so wipe-and-recreate might be the answer. Next time: when a question is about restoring state, include discarding and rebuilding as an option, and say plainly when the scenario behind a question is my inference rather than his requirement.
- **Project**: go-fsfix

### [2026-11-12] Don't add criteria to Mike's stated purpose for go-dt, and count discovery cost as a con
Situation: planning E-2568, I asked where git domain types (BranchName, Revision, ObjectID...) should live. To argue against go-dt, I said go-dt's purpose is types any Go programmer would accept 'without caring which tools they use'. Mike never stated that; his bar is 'what almost everyone would accept as correct'. I also gave option D (a dependency-free go-gitutil subpackage) 'Cons: none'. Mike pointed out a real con: developers must learn about another domain-types package besides go-dt, and must decide which one to import. Next time: quote Mike's criteria as he stated them, without additions. When an option creates a new home for domain types, count how developers would find it and how they would choose between it and go-dt as a con. Mike chose go-dt's root package for git types, with a Git prefix on every name (GitRevision, GitRefName, GitObjectID).
- **Project**: go-fsfix

### [2026-11-12] Enumerate the sentinel errors when planning go-dt types
Situation: planning E-2599 (git domain types for go-dt). My plan said only that each parse error wraps a type sentinel 'and so on' plus a rule sentinel, without naming them. Mike noticed the plan defined no sentinel errors. Reason: sentinels are exported API in go-dt, so a plan that leaves them unnamed leaves part of the API undesigned and unreviewable. Next time a plan adds go-dt types, list every new sentinel with its message, name the existing sentinels it reuses, and map each type to its rule sentinels.
- **Project**: go-fsfix
