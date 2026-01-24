# Task Completion Checklist

When completing a task, run these steps:

1. **Format code**
   ```bash
   go fmt ./...
   ```

2. **Run linting**
   ```bash
   go vet ./...
   ```

3. **Run tests**
   ```bash
   go test ./...
   ```

4. **Build to verify**
   ```bash
   go build ./cmd/...
   ```

5. **Git workflow**
   ```bash
   git status
   git add <specific files>
   git commit -m "type(scope): description"
   git push
   ```

## Commit Message Format
```
type(scope): description

Types: feat, fix, chore, docs, test, refactor
Scopes: vauxhall, praude, tandemonium, pollard, tui, build
```
