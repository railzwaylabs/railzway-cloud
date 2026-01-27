# Conventional Commits Guide for Railzway Cloud

This project uses [Conventional Commits](https://www.conventionalcommits.org/) with **semantic-release** for automated versioning and deployment.

## Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

| Type | Description | Version Bump | Example |
|------|-------------|--------------|---------|
| `feat` | New feature | **Minor** (1.x.0) | `feat: add OAuth client management` |
| `fix` | Bug fix | **Patch** (1.0.x) | `fix: resolve database connection timeout` |
| `perf` | Performance improvement | **Patch** (1.0.x) | `perf: optimize Nomad job generation` |
| `refactor` | Code refactoring | **Patch** (1.0.x) | `refactor: extract auth logic to service` |
| `revert` | Revert previous commit | **Patch** (1.0.x) | `revert: rollback OAuth changes` |
| `docs` | Documentation only | **None** | `docs: update deployment guide` |
| `style` | Code style changes | **None** | `style: format code with gofmt` |
| `test` | Adding/updating tests | **None** | `test: add unit tests for deploy usecase` |
| `chore` | Maintenance tasks | **None** | `chore: update dependencies` |
| `ci` | CI/CD changes | **None** | `ci: add coverage check` |
| `build` | Build system changes | **None** | `build: update Dockerfile` |

### Breaking Changes

To trigger a **major version bump** (x.0.0), add `!` after the type or include `BREAKING CHANGE:` in the footer:

```bash
feat!: migrate to new database schema

BREAKING CHANGE: Database schema has changed, requires migration
```

## Examples

### Feature (Minor Release)
```bash
git commit -m "feat(api): add user profile endpoint

Adds GET /api/user/profile endpoint to retrieve user information
including name, email, and organization details."
```

### Bug Fix (Patch Release)
```bash
git commit -m "fix(auth): resolve OAuth redirect loop

Fixed issue where OAuth callback was redirecting to /login
instead of the intended destination."
```

### Breaking Change (Major Release)
```bash
git commit -m "feat(db)!: migrate to PostgreSQL 16

BREAKING CHANGE: Requires PostgreSQL 16+. Run migrations before deploying."
```

### Non-Release Commits
```bash
git commit -m "docs: update README with deployment instructions"
git commit -m "test: add integration tests for instance lifecycle"
git commit -m "chore: update Go dependencies"
```

## Scopes (Optional but Recommended)

Common scopes for this project:
- `api` - API endpoints and handlers
- `auth` - Authentication and authorization
- `db` - Database operations
- `nomad` - Nomad integration
- `deploy` - Deployment logic
- `billing` - Billing integration
- `ui` - Frontend changes

## Automated Release Flow

1. **Commit with conventional format** → Push to branch
2. **Create PR** → CI runs tests
3. **Merge to `main`** → Semantic Release analyzes commits
4. **Auto-create tag** → Based on commit types
5. **Auto-deploy** → Build Docker image and deploy to GCE
6. **Generate changelog** → Updated automatically

## Tips

✅ **DO:**
- Use imperative mood: "add feature" not "added feature"
- Keep subject line under 72 characters
- Reference issues: `fix(auth): resolve #123`
- Be specific and descriptive

❌ **DON'T:**
- Use vague messages: "fix bug", "update code"
- Mix multiple changes in one commit
- Forget the type prefix

## Quick Reference

```bash
# Feature
git commit -m "feat: add new API endpoint"

# Bug fix
git commit -m "fix: resolve memory leak"

# Breaking change
git commit -m "feat!: change API response format"

# Non-release
git commit -m "docs: update README"
```

---

**Remember:** Every merge to `main` with `feat`, `fix`, `perf`, `refactor`, or `revert` will trigger an automatic release and deployment! 🚀
