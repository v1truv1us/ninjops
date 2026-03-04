# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in ninjops, please report it responsibly.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of these methods:

1. **Email**: Send details to security@ninjops.dev (if available)
2. **GitHub Security Advisory**: Use GitHub's private vulnerability reporting feature

### What to Include

Please include the following information:

- Type of vulnerability (e.g., token exposure, injection, etc.)
- Full paths of source file(s) related to the vulnerability
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability

### Response Timeline

- **Initial Response**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Fix Development**: Depends on severity and complexity
- **Disclosure**: After fix is released

## Security Best Practices

When using ninjops:

### Secrets Management

1. **Never commit secrets** to version control
2. **Use environment variables** for API tokens and keys
3. **Never log secrets** - ninjops redacts tokens in logs by default

### API Tokens

```bash
# Good: Environment variable
export NINJOPS_NINJA_API_TOKEN="your-token-here"

# Bad: In config file (don't do this)
# ninja.api_token = "your-token-here"  # NEVER do this
```

### Local Server Security

1. **Binds to 127.0.0.1 by default** - No external access
2. **No authentication by default** - Designed for local single-user
3. **Use caution with --listen flag** - Only use on trusted networks

```bash
# Safe: Localhost only (default)
ninjops serve

# Potentially unsafe: Exposed to network
ninjops serve --listen 0.0.0.0  # Only on trusted networks!
```

### File Permissions

1. **Config files**: Should be readable only by owner (`chmod 600`)
2. **State files**: Contain reference IDs, protect accordingly
3. **QuoteSpec files**: May contain sensitive business info

```bash
chmod 600 ~/.ninjops/config.toml
chmod 700 .ninjops/
```

## Known Security Considerations

### Token Redaction

Ninjops automatically redacts tokens in error messages:
- Shows only first 4 characters + `****`
- Applies to API tokens and API keys
- Cannot be disabled

### Offline Provider

The offline provider processes data locally:
- No external API calls
- No data leaves your machine
- Suitable for sensitive environments

### External AI Providers

When using OpenAI or Anthropic providers:
- QuoteSpec data is sent to external APIs
- Review their privacy policies
- Consider data sensitivity before enabling

## Security Updates

Security updates will be released as:
- Patch versions (e.g., 1.0.1)
- Documented in GitHub Releases
- Announced in security advisories

## Contact

For security concerns:
- GitHub Security Advisories (preferred)
- Email: security@ninjops.dev

Thank you for helping keep ninjops secure!
