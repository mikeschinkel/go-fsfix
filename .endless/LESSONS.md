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
