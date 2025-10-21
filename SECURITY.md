# Security Policy

## Supported Versions

TFE follows semantic versioning. We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of TFE seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do NOT Open a Public Issue

Security vulnerabilities should not be reported publicly. Please do not create a GitHub issue for security problems.

### 2. Report Privately

**Email:** [ggprompts@gmail.com](mailto:ggprompts@gmail.com?subject=Security%20-%20TFE) (use subject: "Security - TFE")

Or open a private security advisory on GitHub.

**Include in your report:**
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if you have one)
- Your contact information for follow-up

### 3. What to Expect

- **Acknowledgment:** We will acknowledge receipt of your vulnerability report within 48 hours
- **Assessment:** We will investigate and assess the severity within 5 business days
- **Updates:** We will keep you informed of our progress toward fixing the vulnerability
- **Credit:** With your permission, we will credit you in the release notes when the fix is published

### 4. Response Timeline

- **Critical vulnerabilities** (remote code execution, authentication bypass): Fix within 7 days
- **High severity** (privilege escalation, data exposure): Fix within 14 days
- **Medium severity** (information disclosure, DoS): Fix within 30 days
- **Low severity**: Fix in next regular release

## Security Best Practices for TFE Users

When using TFE, follow these best practices:

### File Operations
- Be cautious when executing commands via the command prompt (`:` key)
- Review file paths before operations, especially with symbolic links
- Don't run TFE with unnecessary elevated privileges

### Command Execution
- The command prompt executes commands in your current shell
- Commands are quoted and sanitized, but use caution with complex commands
- Review commands from untrusted sources before execution

### External Tools
- TFE integrates with system tools (editors, browsers, clipboard)
- Ensure these tools are from trusted sources
- Keep your system and tools up to date

### Directories and Files
- Be cautious when navigating untrusted directories
- Maliciously named files (with special characters) are handled safely, but verify before operations
- Preview mode may render content - be aware when viewing untrusted files

## Known Security Considerations

### Terminal Escape Sequences
- Preview mode renders file content as plain text
- Terminal escape sequences in files are displayed but not executed
- Markdown rendering uses Glamour with safe defaults

### Command Prompt
- Uses `bash -c` with proper quoting via `shellQuote()`
- Paths and commands are sanitized before execution
- Command history is stored locally in memory (not persisted)

### External Editor Integration
- When opening files in external editors, paths are properly quoted
- TFE suspends while editor runs, then resumes
- Editor choice follows system defaults (micro, nano, vim, vi)

### File System Access
- TFE operates with your user's file system permissions
- No privilege escalation attempts
- Respects system file permissions and ownership

## Disclosure Policy

- We follow responsible disclosure practices
- Security fixes will be released as quickly as possible
- Public disclosure will occur after a fix is available
- Credit will be given to reporters (unless they prefer anonymity)

## Security Update Notifications

To stay informed about security updates:

1. Watch the [TFE repository](https://github.com/GGPrompts/tfe) for releases
2. Check the [CHANGELOG.md](CHANGELOG.md) for security-related entries
3. Subscribe to GitHub security advisories for this repository

## Past Security Issues

None reported as of initial v1.0 release.

## Contact

For security concerns: [ggprompts@gmail.com](mailto:ggprompts@gmail.com?subject=Security%20-%20TFE)

For general issues: [GitHub Issues](https://github.com/GGPrompts/tfe/issues)

---

**Thank you for helping keep TFE and its users safe!**
