# Security Policy

## Reporting Vulnerabilities

**Do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them responsibly via email to: **security@paymentbridge.dev**

Include:
- Description of the vulnerability
- Steps to reproduce
- Impact assessment
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide a detailed response within 5 business days.

## Security Practices

### What PCP Does NOT Do

- PCP **never** processes raw card numbers (PAN)
- PCP **never** stores CVV, PIN, or sensitive authentication data
- PCP is **not** a payment gateway — it orchestrates through providers

### Data Protection

- All API communication over TLS
- Secrets managed via environment variables (never committed)
- JWT tokens with expiration and issuer validation
- API keys generated with cryptographic randomness (32 bytes)
- Database credentials isolated per environment

### Infrastructure Security

- Docker images built from minimal base (Alpine)
- Non-root container execution
- Trivy vulnerability scanning in CI
- Dependency auditing via Nancy
- SBOM generation (planned)
- Signed container images (planned)

### Authentication & Authorization

- JWT-based authentication
- API key authentication
- Per-merchant API key isolation
- Role-based access control (planned)

### CI/CD Security

- GitHub Actions with minimal permissions
- Secret scanning enabled
- SAST via golangci-lint security linters
- Container scanning via Trivy
- Dependency vulnerability scanning

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅ Current |

## Disclosure Policy

We follow [responsible disclosure](https://en.wikipedia.org/wiki/Responsible_disclosure). Reporters will be credited (with permission) in the security advisory.
