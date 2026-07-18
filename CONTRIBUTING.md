# Contributing to PCP

Thank you for your interest in contributing to Payment Control Plane.

## Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes following the guidelines below
4. Write tests (90% coverage required)
5. Update documentation and CHANGELOG.md
6. Submit a pull request

## Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add payment routing engine
fix: resolve merchant duplicate check
docs: update API documentation
test: add merchant service tests
refactor: extract common error handling
ci: add security scanning step
```

## Code Standards

- Follow Go conventions and `gofmt`
- SOLID, DRY, KISS principles
- No `TODO` comments in source code (use TODO.md instead)
- All exports must have documentation comments
- Error handling must be explicit — no ignored errors

## Pull Request Checklist

- [ ] All tests pass (`make test`)
- [ ] Coverage ≥ 90%
- [ ] Linter passes (`make lint`)
- [ ] Docker build succeeds
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] TODO.md updated if applicable
- [ ] API documentation updated (if API changes)

## Testing Requirements

Every feature requires:
- Unit tests
- Integration tests (where applicable)
- API tests (for endpoint changes)

## Architecture

Follow the Clean Architecture layering. See [ARCHITECTURE.md](ARCHITECTURE.md).

- **Domain layer**: No external dependencies
- **Application layer**: Depends only on domain
- **Infrastructure layer**: Implements domain ports
- **Interface layer**: HTTP/gRPC handlers

## Getting Help

- Open a [Discussion](https://github.com/paymentbridge/pcp/discussions)
- Check [docs/DECISIONS.md](docs/DECISIONS.md) for architecture rationale
