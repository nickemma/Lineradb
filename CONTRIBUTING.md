# Contributing to LineraDB

Thank you for your interest in contributing to LineraDB! This project is primarily a **learning journey** to understand distributed systems from first principles, but community involvement makes it better.

---

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Architecture Guidelines](#architecture-guidelines)
- [Communication](#communication)

---

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [Nicholas Emmanuel](mailto:nicholasemmanuel321@gmail.com).

---

## How Can I Contribute?

### 1. **Reporting Bugs**

Found a bug? Help us fix it!

- **Check existing issues** first to avoid duplicates
- Use the [bug report template](.github/ISSUE_TEMPLATE/1-bug.yml)
- Include:
  - Clear description of the problem
  - Steps to reproduce
  - Expected vs. actual behavior
  - System information (OS, Go version, etc.)
  - Logs or error messages

### 2. **Suggesting Features**

Have an idea? We'd love to hear it!

- Use the [feature request template](.github/ISSUE_TEMPLATE/2-feature.yml)
- Explain:
  - The problem this solves
  - Your proposed solution
  - Alternative approaches you considered
  - Any trade-offs involved

### 3. **Asking Questions**

Confused about something?

- Use the [question template](.github/ISSUE_TEMPLATE/3-question.yml)
- Check [docs/](docs/) first - your answer might be there
- Be specific about what you're trying to understand

### 4. **Improving Documentation**

Documentation is critical for a learning project!

- Fix typos, clarify explanations, add examples
- Update outdated information
- Add diagrams or visualizations
- Improve code comments

### 5. **Contributing Code**

Ready to dive in?

**Great first issues:**

- Look for issues labeled `good first issue`
- Documentation improvements
- Test coverage improvements
- Utility functions

**Advanced contributions:**

- Distributed systems algorithms (Raft, 2PC, MVCC)
- Storage engine optimizations
- Query planner improvements
- Chaos engineering tests

---

## Development Setup

### Prerequisites

- **Go 1.25** - [Install Go](https://go.dev/doc/install)
- **Rust 1.92** (for storage engine) - [Install Rust](https://rustup.rs/)
- **Docker** (optional) - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)

### Local Setup

```bash
# 1. Fork the repository on GitHub

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/lineradb.git
cd lineradb

# 3. Add upstream remote
git remote add upstream https://github.com/yourusername/lineradb.git

# 4. Install dependencies
go mod download

# 5. Build the project
make build

# 6. Run tests
make test

# 7. Run with race detector (important!)
make test-race
```

### Verify Your Setup

```bash
# Should output version
./bin/lineradb-server --version

# Should pass all tests
make test

# Should have no linting errors
make lint
```

---

## Keeping Your Fork Up to Date

If you've forked the repository, sync it regularly to stay current with upstream changes.

```bash
# Add upstream remote (only once)
git remote add upstream https://github.com/yourusername/lineradb.git

# Fetch latest changes
git fetch upstream
```

---

## Pull Request Process

### Before You Start

1. **Open an issue first** - Discuss your idea before writing code
2. **Check existing PRs** - Someone might already be working on it
3. **Start small** - Small, focused PRs are easier to review

### Creating a Pull Request

```bash
# 1. Create a feature branch from develop
git checkout develop
git pull upstream develop
git checkout -b feature/your-feature-name

# 2. Make your changes
# ... edit files ...

# 3. Test thoroughly
make test
make test-race
make lint

# 4. Commit with clear messages
git add .
git commit -m "feat(consensus): implement Raft leader election

- Add leader election algorithm
- Handle split-brain scenarios
- Add unit tests for election timeout
- Update ARCHITECTURE.md with consensus details

Closes #123"

# 5. Push to your fork
git push origin feature/your-feature-name

# 6. Open a PR on GitHub
# Base: develop (not main!)
# Compare: your feature branch -> develop
```

### PR Requirements

Your PR must:

- ✅ Pass all CI checks (tests, linting, builds)
- ✅ Include tests for new functionality
- ✅ Update documentation if needed
- ✅ Follow our coding standards
- ✅ Have a clear description explaining the changes
- ✅ Reference related issues (e.g., "Closes #123")

### PR Review Process

1. **Automated checks** run first (CI/CD)
2. **Code review** by maintainers
3. **Feedback** - Address comments and requested changes
4. **Approval** - Once approved, your PR will be merged to `develop`
5. **Release** - Changes in `develop` are periodically merged to `main`

---

## Coding Standards

### Go Code Style

```go
// Good: Clear naming, explicit error handling
func (s *RaftServer) RequestVote(ctx context.Context, req *VoteRequest) (*VoteResponse, error) {
    if req.Term < s.currentTerm {
        return &VoteResponse{VoteGranted: false}, nil
    }

    // Grant vote if we haven't voted yet
    if s.votedFor == nil || *s.votedFor == req.CandidateID {
        s.votedFor = &req.CandidateID
        return &VoteResponse{VoteGranted: true}, nil
    }

    return &VoteResponse{VoteGranted: false}, nil
}

// Bad: Unclear naming, ignoring errors
func (s *RaftServer) rv(c context.Context, r *VoteRequest) *VoteResponse {
    if r.Term < s.currentTerm {
        return &VoteResponse{VoteGranted: false}
    }
    s.votedFor = &r.CandidateID
    return &VoteResponse{VoteGranted: true}
}
```

**Guidelines:**

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` (run `make lint`)
- Explicit error handling (no `_` for errors unless justified)
- Clear variable names (avoid single letters except loops)
- Document exported functions and types
- Keep functions focused (do one thing well)

### Rust Code Style (Coming Soon)

- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)
- Use `rustfmt` and `clippy`
- Prefer `Result` over panics
- Document unsafe code thoroughly

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `test` - Adding or updating tests
- `refactor` - Code refactoring (no functional change)
- `perf` - Performance improvements
- `chore` - Build process, dependencies, tooling

**Examples:**

```bash
feat(clock): implement hybrid logical clock

- Add HLC timestamp generation
- Implement happens-before relationship
- Add clock synchronization across nodes
- Include unit tests for clock skew scenarios

Closes #45

---

fix(consensus): prevent split-brain in leader election

Previously, two nodes could become leaders if network
partition happened during election. Now using term
numbers correctly to prevent this.

Fixes #78

---

docs(architecture): add consensus protocol documentation

- Explain Raft leader election
- Document failure scenarios
- Add sequence diagrams for key flows
```

---

## Testing Guidelines

### Test Levels

```
test/
├── unit/          # Fast, isolated tests (< 1s each)
├── integration/   # Multi-component tests (< 10s each)
├── e2e/           # Full system tests (< 1min each)
├── chaos/         # Fault injection tests (minutes to hours)
└── benchmark/     # Performance tests
```

### Writing Tests

```go
// Good: Clear test name, arranged in Given-When-Then
func TestRaftLeaderElection_WhenLeaderFails_NewLeaderIsElected(t *testing.T) {
    // Given: A 3-node Raft cluster with a leader
    cluster := NewTestCluster(t, 3)
    defer cluster.Shutdown()

    leader := cluster.WaitForLeader(t)

    // When: The leader crashes
    cluster.StopNode(leader.ID)

    // Then: A new leader is elected within timeout
    newLeader := cluster.WaitForLeader(t)
    if newLeader.ID == leader.ID {
        t.Fatal("New leader should be different from crashed leader")
    }
}

// Bad: Unclear what's being tested
func TestRaft(t *testing.T) {
    c := NewTestCluster(t, 3)
    c.StopNode(0)
    time.Sleep(5 * time.Second)
    // What are we testing?
}
```

### Test Requirements

- ✅ Test both **happy path** and **failure scenarios**
- ✅ Use **table-driven tests** for multiple inputs
- ✅ Include **race detector** tests (`go test -race`)
- ✅ Add **chaos tests** for distributed components
- ✅ Aim for **>80% code coverage** (check with `make test`)

### Running Tests

```bash
# Run all tests
make test

# Run with race detector (critical!)
make test-race

# Run specific test
go test -v -run TestRaftLeaderElection ./internal/consensus/...

# Run benchmarks
go test -bench=. ./internal/storage/...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Architecture Guidelines

LineraDB follows **Hexagonal Architecture** with clear module boundaries.

### Module Structure

```
internal/<module>/
├── domain/         # Pure business logic (no external dependencies)
├── application/    # Use cases/orchestration
├── repository/     # Interfaces (ports)
└── infrastructure/ # Concrete implementations (adapters)
```

**Rules:**

- `domain/` never imports `infrastructure/`
- `application/` imports `domain/` and `repository/` (interfaces only)
- `infrastructure/` implements `repository/` interfaces
- Dependencies point inward (domain is the core)

### Adding a New Module

```bash
# 1. Create module structure
mkdir -p internal/yourmodule/{domain,application,repository,infrastructure}

# 2. Define domain entities
# internal/yourmodule/domain/entity.go

# 3. Define use cases
# internal/yourmodule/application/service.go

# 4. Define repository interface
# internal/yourmodule/repository/repo.go

# 5. Implement adapter
# internal/yourmodule/infrastructure/impl.go

# 6. Add tests
mkdir -p test/unit/yourmodule
# test/unit/yourmodule/entity_test.go

# 7. Update documentation
# docs/ARCHITECTURE.md - Add your module
```

### Design Principles

1. **Separation of Concerns** - Each module has one clear responsibility
2. **Dependency Inversion** - Depend on interfaces, not implementations
3. **Explicit Constraints** - Document physical limits (latency, CAP, etc.)
4. **Fail Explicitly** - Never silently ignore errors
5. **Test Failure Modes** - Distributed systems fail; test for it

---

## Communication

### Where to Ask Questions

- **GitHub Issues** - Bug reports, feature requests, questions
- **GitHub Discussions** - General discussion, ideas, help
- **Email** - [Nicholas Emmanuel](mailto:nicholasemmanuel321@gmail.com) for private matters

### Response Times

This is a **solo learning project**, so responses may take:

- Bug reports: 1-3 days
- Feature requests: 1-7 days
- Questions: 1-3 days
- PR reviews: 1-7 days

**Please be patient!** I'm balancing this with other responsibilities.

---

## 🙏 Recognition

All contributors will be listed in:

- GitHub's contributors page
- Release notes (for significant contributions)
- `CONTRIBUTORS.md` (coming soon)

Thank you for helping make LineraDB better! 🚀

---

## 📚 Resources

- [Designing Data-Intensive Applications](https://dataintensive.net/) by Martin Kleppmann
- [Database Internals](https://www.databass.dev/) by Alex Petrov
- [Raft Paper](https://raft.github.io/raft.pdf) by Ongaro & Ousterhout
- [Jepsen](https://jepsen.io/) for distributed systems testing
- [CockroachDB Architecture](https://www.cockroachlabs.com/docs/stable/architecture/overview.html)

---

<div align="center">

**Thank you for contributing to LineraDB!**

[⬆ Back to Top](#contributing-to-lineradb)

</div>
