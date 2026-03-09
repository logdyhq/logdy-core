# Security Policy

## Supported Versions

Security fixes are applied to the latest released version on `main`.

## Reporting a Vulnerability

If you discover a security issue, please do not open a public issue with exploit details.

Use one of these channels:

- GitHub private vulnerability report (Security tab)
- Email: security@logdy.dev

Please include:

- Affected version
- Reproduction steps
- Impact assessment
- Proposed remediation (if available)

We will acknowledge receipt as quickly as possible and coordinate disclosure.

## Security Hardening Recommendations

For local-only usage:

- Bind UI to loopback: `--ui-ip 127.0.0.1`
- Disable updates check in restricted environments: `--no-updates`
- Keep analytics disabled unless explicitly needed

If binding to non-local interfaces:

- Always set `--ui-pass`
- Always set `--api-key`
- Restrict network access with host firewall and reverse proxy policy
- Prefer TLS termination in front of Logdy when exposed beyond localhost
