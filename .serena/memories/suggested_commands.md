# Suggested Commands

## Build & Run
```bash
go build ./cmd/...          # Build all tools
./dev vauxhall              # Run Vauxhall web mode
./dev vauxhall --tui        # Run Vauxhall TUI mode
./dev praude                # Run Praude TUI
./dev tandemonium           # Run Tandemonium TUI
./dev pollard               # Run Pollard CLI
```

## Testing
```bash
go test ./...                           # Test all
go test ./internal/pollard/... -v       # Test specific package
go test ./internal/pollard/hunters -v   # Test hunters
```

## Formatting & Linting
```bash
go fmt ./...                # Format code
go vet ./...                # Lint code
```

## Git
```bash
git status                  # Check changes
git add <files>             # Stage files
git commit -m "type(scope): message"  # Commit
git push                    # Push to remote
```

## System Utils
- Linux system: standard `ls`, `cd`, `grep`, `find`, `cat`, etc.
