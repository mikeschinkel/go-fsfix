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
