# Contributing to GoLaserCut

Thank you for your interest in contributing!

## How to contribute

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Run tests: `go test ./...`
5. Ensure the build passes: `go build ./...`
6. Submit a Pull Request

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- All HTML elements must have unique `id` attributes (see coding standards)
- Separate HTML templates and JavaScript into distinct files
- Write tests for new functionality

## Pull Request guidelines

- Keep PRs focused on a single change
- Include tests for new features or bug fixes
- Write clear commit messages explaining the "why"
- Ensure all tests pass before submitting

## Reporting issues

Open an issue on GitHub with:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
