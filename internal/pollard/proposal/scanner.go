package proposal

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ContextScanner extracts project context from documentation files.
type ContextScanner struct {
	projectPath string
	maxFileSize int64
	maxContent  int // max chars to keep per file
}

// NewContextScanner creates a scanner for the given project path.
func NewContextScanner(projectPath string) *ContextScanner {
	return &ContextScanner{
		projectPath: projectPath,
		maxFileSize: 1024 * 1024, // 1MB max
		maxContent:  4000,        // 4000 chars per file
	}
}

// Scan extracts project context from documentation files.
func (s *ContextScanner) Scan() (*ProjectContext, error) {
	ctx := &ProjectContext{
		Files: make(map[string]string),
	}

	// Read documentation files in priority order
	docFiles := []string{"CLAUDE.md", "AGENTS.md", "README.md"}
	for _, file := range docFiles {
		content, err := s.readFile(file)
		if err == nil && content != "" {
			ctx.Files[file] = truncate(content, s.maxContent)
		}
	}

	// Detect technologies from config files
	ctx.Technologies = s.detectTechnologies()

	// Detect project type
	ctx.DetectedType = s.detectProjectType()

	// Extract project name
	ctx.ProjectName = s.extractProjectName()

	// Extract description from CLAUDE.md or README.md
	ctx.Description = s.extractDescription(ctx.Files)

	// Try to detect domain
	ctx.Domain = s.detectDomain(ctx.Files, ctx.Technologies)

	return ctx, nil
}

// ScanWithSrc also scans source files for additional context.
func (s *ContextScanner) ScanWithSrc() (*ProjectContext, error) {
	ctx, err := s.Scan()
	if err != nil {
		return nil, err
	}

	// Add main package/entry point files
	mainFiles := []string{
		"main.go", "cmd/main.go",
		"index.ts", "src/index.ts",
		"app.py", "main.py",
		"package.json", "go.mod", "Cargo.toml",
	}

	for _, file := range mainFiles {
		content, err := s.readFile(file)
		if err == nil && content != "" {
			// Only include first 1000 chars for source files
			ctx.Files[file] = truncate(content, 1000)
			break // Only include one main file
		}
	}

	return ctx, nil
}

// readFile reads a file relative to the project path.
func (s *ContextScanner) readFile(relPath string) (string, error) {
	fullPath := filepath.Join(s.projectPath, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}

	if info.Size() > s.maxFileSize {
		return "", nil // Skip oversized files
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// detectTechnologies identifies technologies from config files.
func (s *ContextScanner) detectTechnologies() []string {
	var techs []string
	seen := make(map[string]bool)

	addTech := func(tech string) {
		if !seen[tech] {
			seen[tech] = true
			techs = append(techs, tech)
		}
	}

	// Go project
	if _, err := os.Stat(filepath.Join(s.projectPath, "go.mod")); err == nil {
		addTech("Go")
		// Check for common Go frameworks
		if content, err := s.readFile("go.mod"); err == nil {
			if strings.Contains(content, "github.com/spf13/cobra") {
				addTech("Cobra CLI")
			}
			if strings.Contains(content, "github.com/charmbracelet/bubbletea") {
				addTech("Bubble Tea TUI")
			}
			if strings.Contains(content, "github.com/gin-gonic/gin") {
				addTech("Gin")
			}
			if strings.Contains(content, "github.com/labstack/echo") {
				addTech("Echo")
			}
		}
	}

	// Node.js project
	if content, err := s.readFile("package.json"); err == nil {
		addTech("Node.js")
		if strings.Contains(content, "\"next\"") {
			addTech("Next.js")
		}
		if strings.Contains(content, "\"react\"") {
			addTech("React")
		}
		if strings.Contains(content, "\"typescript\"") {
			addTech("TypeScript")
		}
		if strings.Contains(content, "\"tailwindcss\"") {
			addTech("Tailwind CSS")
		}
		if strings.Contains(content, "\"prisma\"") {
			addTech("Prisma")
		}
		if strings.Contains(content, "\"express\"") {
			addTech("Express")
		}
	}

	// Python project
	if _, err := os.Stat(filepath.Join(s.projectPath, "requirements.txt")); err == nil {
		addTech("Python")
	}
	if _, err := os.Stat(filepath.Join(s.projectPath, "pyproject.toml")); err == nil {
		addTech("Python")
	}
	if content, err := s.readFile("pyproject.toml"); err == nil {
		if strings.Contains(content, "fastapi") {
			addTech("FastAPI")
		}
		if strings.Contains(content, "django") {
			addTech("Django")
		}
		if strings.Contains(content, "flask") {
			addTech("Flask")
		}
	}

	// Rust project
	if _, err := os.Stat(filepath.Join(s.projectPath, "Cargo.toml")); err == nil {
		addTech("Rust")
	}

	// Ruby project
	if _, err := os.Stat(filepath.Join(s.projectPath, "Gemfile")); err == nil {
		addTech("Ruby")
		if content, err := s.readFile("Gemfile"); err == nil {
			if strings.Contains(content, "rails") {
				addTech("Rails")
			}
		}
	}

	// Database detection
	if _, err := os.Stat(filepath.Join(s.projectPath, ".pollard/state.db")); err == nil {
		addTech("SQLite")
	}
	if s.fileContainsAny("docker-compose.yml", "postgres", "postgresql") {
		addTech("PostgreSQL")
	}
	if s.fileContainsAny("docker-compose.yml", "mongo") {
		addTech("MongoDB")
	}

	return techs
}

// detectProjectType determines the type of project.
func (s *ContextScanner) detectProjectType() string {
	// Check for monorepo patterns
	if s.hasMultipleTools() {
		return "monorepo"
	}

	// Check for CLI indicators
	if _, err := os.Stat(filepath.Join(s.projectPath, "cmd")); err == nil {
		return "cli"
	}

	// Check for web app indicators
	if _, err := os.Stat(filepath.Join(s.projectPath, "pages")); err == nil {
		return "web"
	}
	if _, err := os.Stat(filepath.Join(s.projectPath, "app")); err == nil {
		if _, err := os.Stat(filepath.Join(s.projectPath, "app/page.tsx")); err == nil {
			return "web"
		}
	}

	// Check for API indicators
	if s.fileContainsAny("main.go", "http.ListenAndServe", "gin.", "echo.") {
		return "api"
	}
	if s.fileContainsAny("package.json", "express", "fastify", "koa") {
		return "api"
	}

	// Check for library indicators
	if s.fileContainsAny("package.json", `"main":`, `"module":`) {
		if !s.fileContainsAny("package.json", "next", "react-dom") {
			return "library"
		}
	}

	return "unknown"
}

// hasMultipleTools checks if this is a monorepo with multiple tools.
func (s *ContextScanner) hasMultipleTools() bool {
	cmdDir := filepath.Join(s.projectPath, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return false
	}

	toolCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			toolCount++
		}
	}
	return toolCount > 1
}

// extractProjectName gets the project name from various sources.
func (s *ContextScanner) extractProjectName() string {
	// Try go.mod first
	if content, err := s.readFile("go.mod"); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "module ") {
				parts := strings.Split(strings.TrimPrefix(line, "module "), "/")
				return parts[len(parts)-1]
			}
		}
	}

	// Try package.json
	if content, err := s.readFile("package.json"); err == nil {
		nameRe := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
		if matches := nameRe.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}

	// Try Cargo.toml
	if content, err := s.readFile("Cargo.toml"); err == nil {
		nameRe := regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
		if matches := nameRe.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}

	// Fall back to directory name
	return filepath.Base(s.projectPath)
}

// extractDescription extracts a description from docs.
func (s *ContextScanner) extractDescription(files map[string]string) string {
	// Try CLAUDE.md Overview section
	if content, ok := files["CLAUDE.md"]; ok {
		if desc := extractSection(content, "## Overview"); desc != "" {
			return truncate(desc, 500)
		}
	}

	// Try README.md first paragraph
	if content, ok := files["README.md"]; ok {
		lines := strings.Split(content, "\n")
		var desc strings.Builder
		inParagraph := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" && inParagraph {
				break
			}
			if !strings.HasPrefix(line, "#") && line != "" {
				inParagraph = true
				desc.WriteString(line)
				desc.WriteString(" ")
			}
		}
		if desc.Len() > 0 {
			return truncate(strings.TrimSpace(desc.String()), 500)
		}
	}

	return ""
}

// detectDomain attempts to detect the project's domain.
func (s *ContextScanner) detectDomain(files map[string]string, techs []string) string {
	// Combine all file content for analysis
	var combined strings.Builder
	for _, content := range files {
		combined.WriteString(content)
		combined.WriteString(" ")
	}
	text := strings.ToLower(combined.String())

	// Check for domain keywords
	domains := map[string][]string{
		"developer-tools": {"cli", "developer", "ide", "editor", "code", "programming", "agent", "mcp"},
		"ai-ml":           {"machine learning", "ai", "llm", "neural", "model", "training", "inference"},
		"healthcare":      {"medical", "health", "patient", "clinical", "diagnosis", "treatment"},
		"fintech":         {"payment", "banking", "finance", "transaction", "trading", "crypto"},
		"e-commerce":      {"shop", "cart", "product", "checkout", "payment", "order"},
		"education":       {"learning", "course", "student", "education", "teaching", "quiz"},
		"social":          {"social", "community", "message", "chat", "network", "friend"},
		"gaming":          {"game", "player", "score", "level", "achievement"},
	}

	bestDomain := ""
	bestScore := 0

	for domain, keywords := range domains {
		score := 0
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestDomain = domain
		}
	}

	if bestScore >= 2 {
		return bestDomain
	}
	return ""
}

// fileContainsAny checks if a file contains any of the given strings.
func (s *ContextScanner) fileContainsAny(relPath string, strs ...string) bool {
	content, err := s.readFile(relPath)
	if err != nil {
		return false
	}
	for _, str := range strs {
		if strings.Contains(content, str) {
			return true
		}
	}
	return false
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func extractSection(content, header string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder
	inSection := false
	headerLevel := strings.Count(header, "#")

	for _, line := range lines {
		if strings.HasPrefix(line, header) {
			inSection = true
			continue
		}
		if inSection {
			// Check if we've hit a header of same or higher level
			if strings.HasPrefix(line, "#") {
				level := 0
				for _, c := range line {
					if c == '#' {
						level++
					} else {
						break
					}
				}
				if level <= headerLevel {
					break
				}
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}
