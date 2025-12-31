# Rules of Engagement - Sliver Dashboard Development

## Git Workflow

### Branch Structure
- **`go-bubbletea`** - Primary development branch (work happens here)
- **`dev`** - Development/testing branch (stable features)
- **`master`** - Production branch (release-ready code)

### Development Workflow

#### 1. All Development on `go-bubbletea`
```bash
git checkout go-bubbletea
# Make changes, test, build
```

#### 2. **NEVER Commit Without User Approval**
- ❌ **DO NOT** run `git commit` without explicit user permission
- ❌ **DO NOT** run `git push` without explicit user permission
- ✅ **ALWAYS** ask user to review changes first
- ✅ **ALWAYS** show `git diff` or `git status` for user review
- ✅ **WAIT** for user to say "commit this" or "push this"

#### 3. Commit Process (Only After User Approval)
```bash
# User must explicitly approve BEFORE running these commands
git add <files>
git commit -m "message"
```

#### 4. Cherry-Pick to Other Branches (Only After User Approval)
```bash
# User must explicitly instruct to cherry-pick
git checkout dev
git cherry-pick <commit-hash>

git checkout master
git cherry-pick <commit-hash>

git checkout go-bubbletea
```

#### 5. Push to Remote (Only After User Approval)
```bash
# User must explicitly say "push to remote" or "push it"
git push origin go-bubbletea
git push origin dev
git push origin master
```

---

## Code Modification Rules

### Before Making Changes
1. **Read relevant files** to understand current implementation
2. **Show user** what you plan to change
3. **Wait for approval** before modifying files

### Build & Test Protocol
- **Always build** after changes: `bash build.sh`
- **Show build output** to user
- **Report any errors** immediately
- **Never assume** build success means user wants to commit

### File Operations
- **Prefer editing** existing files over creating new ones
- **Ask first** before creating new files
- **Show diffs** after making changes
- **Use Read tool** before Edit/Write tools

---

## Communication Protocol

### What to Show User Before Committing
1. `git status` - Files changed
2. `git diff` - Actual changes made
3. Build result - Verify it compiles
4. Test results (if applicable)

### What to Ask User
- "Would you like me to commit these changes?"
- "Should I cherry-pick this to dev and master?"
- "Should I push to remote?"

### What NOT to Do
- ❌ Don't commit automatically after completing a task
- ❌ Don't push automatically after committing
- ❌ Don't assume user wants changes committed
- ❌ Don't batch multiple commits without approval for each

---

## Task Management

### Using TodoWrite Tool
- **Use proactively** for multi-step tasks
- **Mark in_progress** when starting a task
- **Mark completed** when done (but this does NOT mean commit!)
- **Update user** on progress

### Completion vs Commit
- ✅ **Task Complete** = Code changes done, builds successfully
- ⏸️ **Awaiting User Review** = User needs to review before commit
- ✅ **User Approved** = User explicitly says "commit this"
- 🚀 **Committed & Pushed** = Changes are in git (only after approval)

---

## Project-Specific Context

### Current Project: Sliver C2 Dashboard
- **Technology**: Go + Bubbletea TUI framework
- **Location**: `/home/kali/Desktop/git/sliver-graphs/go/`
- **Build Command**: `bash build.sh`
- **Run Command**: `./sliver-graph`

### Key Files
- `main.go` - Main application (3400+ lines)
- `internal/client/sliver.go` - Sliver API client
- `internal/models/agent.go` - Data structures
- `internal/config/` - Configuration helpers
- `internal/alerts/` - Alert system
- `internal/tracking/` - Activity tracking
- `internal/tree/` - Tree visualization

### Testing Pattern
```bash
# Build
bash build.sh

# Run
./sliver-graph

# Navigate
D - Dashboard view
F1-F5 - Dashboard pages
Tab/Shift+Tab - Cycle pages
q - Quit
```

---

## Emergency Procedures

### If Changes Were Committed Without Approval
1. **Notify user immediately**: "I committed changes without approval - I apologize"
2. **Show what was committed**: `git log -1 --stat`
3. **Offer to undo**: "Would you like me to undo this commit?"
4. **Wait for instructions**: User decides what to do

### Undo Last Commit (If User Requests)
```bash
# Undo commit but keep changes
git reset --soft HEAD~1

# Undo commit and discard changes
git reset --hard HEAD~1  # DANGEROUS - only if user explicitly requests
```

### If Pushed Without Approval
1. **Notify user immediately**
2. **DO NOT** force push without user approval
3. **Wait for user decision** on how to handle

---

## Summary - The Golden Rules

1. **Work on `go-bubbletea` branch only** (unless instructed otherwise)
2. **NEVER commit without explicit user approval**
3. **NEVER push without explicit user approval**
4. **ALWAYS show diffs/changes before asking to commit**
5. **ASK before cherry-picking to dev/master**
6. **BUILD and TEST before showing user**
7. **WAIT for "commit this" or "push this" commands**

---

## Example Correct Workflow

```
User: "Add feature X"
Assistant: [Makes changes]
Assistant: "I've implemented feature X. Let me show you what changed:"
Assistant: [Shows git diff and build result]
Assistant: "The build is successful. Would you like me to commit these changes?"
User: "Yes, commit it"
Assistant: [Commits with good message]
Assistant: "Committed. Would you like me to cherry-pick to dev and master?"
User: "Yes"
Assistant: [Cherry-picks to both branches]
Assistant: "Cherry-picked to dev and master. Should I push to remote?"
User: "Push it"
Assistant: [Pushes all branches]
Assistant: "✅ Pushed to all branches"
```

---

**Last Updated**: December 29, 2025  
**Version**: 1.0  
**Applies To**: All development on sliver-graphs project
