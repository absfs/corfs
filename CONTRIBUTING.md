# Contributing to CorFS

Thank you for your interest in contributing to CorFS! We welcome contributions from the community.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:
- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- Go version and operating system
- Any relevant code samples or error messages

### Suggesting Features

We welcome feature suggestions! Please open an issue with:
- A clear description of the feature
- Use cases and benefits
- Any implementation ideas (optional)

### Pull Requests

1. **Fork the repository** and create your branch from `master`
2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Ensure all tests pass** with `go test -v -race -cover ./...`
5. **Run go vet** with `go vet ./...`
6. **Format your code** with `gofmt -w .`
7. **Update documentation** as needed
8. **Update CHANGELOG.md** with your changes
9. **Submit a pull request**

## Development Setup

```bash
# Clone the repository
git clone https://github.com/absfs/corfs.git
cd corfs

# Install dependencies
go mod download

# Run tests
go test -v -race -cover ./...

# Run benchmarks
go test -bench=. -benchmem

# Check code quality
go vet ./...
gofmt -l .
```

## Coding Standards

- Follow standard Go conventions and idioms
- Write clear, descriptive commit messages
- Add comments for exported functions and types
- Keep functions focused and concise
- Maintain test coverage above 80%
- Use meaningful variable and function names

## Testing

- Write unit tests for all new code
- Ensure tests are deterministic and fast
- Use table-driven tests where appropriate
- Test edge cases and error conditions
- Run tests with race detection enabled

## Commit Messages

Follow conventional commit format:
- `feat: add new caching feature`
- `fix: resolve cache invalidation issue`
- `docs: update README with examples`
- `test: add tests for File.Read`
- `refactor: simplify OpenFile logic`
- `chore: update dependencies`

## Questions?

Feel free to open an issue for any questions or concerns. We're here to help!

## License

By contributing to CorFS, you agree that your contributions will be licensed under the MIT License.
