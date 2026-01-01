# Contributing to Sliver C2 TUI

Thank you for considering contributing to Sliver C2 TUI! We welcome contributions from the community.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)

---

## Code of Conduct

This project follows a simple code of conduct:

- Be respectful and constructive
- Focus on what is best for the community
- Show empathy towards other community members
- Accept constructive criticism gracefully

---

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title** - Descriptive summary of the issue
- **Steps to reproduce** - Detailed steps to reproduce the behavior
- **Expected behavior** - What you expected to happen
- **Actual behavior** - What actually happened
- **Environment** - OS, terminal, Go version, etc.
- **Screenshots** - If applicable

### Suggesting Enhancements

Enhancement suggestions are welcome! Include:

- **Clear title** - Concise description of the feature
- **Use case** - Why is this feature needed?
- **Proposed solution** - How should it work?
- **Alternatives** - Other solutions you've considered

### Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Sliver C2 server (for testing)
- Modern terminal with Unicode support

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/sliver-tui.git
cd sliver-tui/go

# Download dependencies
go mod download

# Build
go build -o sliver-graph main.go

# Run
./sliver-graph
```

### Project Structure

```
go/
├── main.go                  # Main application entry point
├── internal/
│   ├── alerts/              # Alert management system
│   │   └── alerts.go
│   ├── client/              # Sliver client connection
│   │   └── sliver.go
│   ├── config/              # Configuration (themes, views)
│   │   ├── domains.go
│   │   ├── themes.go
│   │   └── views.go
│   ├── models/              # Data models
│   │   └── agent.go
│   ├── tracking/            # Activity and change tracking
│   │   ├── activity.go
│   │   └── changes.go
│   └── tree/                # Tree view builder
│       └── builder.go
├── go.mod
└── go.sum
```

---

## Coding Guidelines

### Go Style

Follow standard Go conventions:

- Run `go fmt` before committing
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and concise
- Handle errors explicitly

### Code Organization

- **Keep functions small** - Aim for single responsibility
- **Use descriptive names** - Make code self-documenting
- **Add comments** - Explain "why", not "what"
- **Avoid globals** - Pass dependencies explicitly
- **Test your code** - Add tests for new features

### Example

```go
// Good
func (m *model) renderAgentBox(agent Agent) string {
    // Render individual agent in box format
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Render(agent.Hostname)
}

// Avoid
func (m *model) r(a Agent) string {
    return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(a.Hostname)
}
```

---

## Commit Messages

We follow conventional commits for clear history:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

### Examples

```bash
feat(dashboard): add network intel page

- Add subnet grouping functionality
- Display compromised hosts by network
- Include statistics for each subnet

Closes #123
```

```bash
fix(alerts): correct click detection coordinates

Alert clicks were off by 2 lines due to viewport offset.
Fixed by adjusting click coordinate calculation.

Fixes #456
```

---

## Pull Request Process

### Before Submitting

1. **Update your fork**
   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

2. **Test thoroughly**
   - Build succeeds
   - No regressions
   - New features work as expected

3. **Format code**
   ```bash
   go fmt ./...
   ```

4. **Update documentation**
   - Update README if needed
   - Add inline comments
   - Update CHANGELOG

### Submitting

1. **Push to your fork**
   ```bash
   git push origin feature/your-feature
   ```

2. **Create Pull Request**
   - Use descriptive title
   - Reference related issues
   - Explain what and why
   - Add screenshots if UI changes

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How has this been tested?

## Screenshots
If applicable

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings
- [ ] Tests added/updated
```

### Review Process

1. Maintainer reviews PR
2. Feedback provided if needed
3. Make requested changes
4. PR approved and merged

---

## Development Tips

### Testing Locally

```bash
# Build and run
go build -o sliver-graph main.go
./sliver-graph

# Build with race detector
go build -race -o sliver-graph main.go

# Run with verbose output
./sliver-graph -v
```

### Debugging

- Use `fmt.Printf()` for quick debugging
- Check logs in terminal output
- Test with different terminal sizes
- Try all themes and views

### Common Issues

**Build fails:**
- Run `go mod tidy` to clean dependencies
- Check Go version (`go version`)
- Ensure all imports are correct

**Runtime errors:**
- Check Sliver client configuration
- Verify server is running
- Check file permissions

---

## Questions?

- Open an issue for questions
- Join discussions on GitHub
- Email: ianmusyoka101@gmail.com

---

Thank you for contributing! 🎉
