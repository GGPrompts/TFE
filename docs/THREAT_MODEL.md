# TFE Security & Threat Model

## Overview

**TFE is a local terminal file manager where the user is the operator, not an attacker.**

TFE operates under the principle that the user has legitimate access to their system and is using the application for its intended purpose: browsing and managing files in a terminal environment.

---

## Threat Model

### Trust Boundary

**TFE trusts:**
- The user running the application
- The local operating system
- The terminal emulator
- The file system permissions enforced by the OS

**TFE does NOT trust:**
- (Nothing - all actors in the threat model are trusted)

### Assumptions

1. **User has legitimate access**: The user is authorized to use the system and access the files they can see
2. **OS enforces permissions**: The operating system correctly enforces file system permissions
3. **Terminal is secure**: The terminal emulator is not compromised
4. **Local execution**: TFE runs locally, not as a networked service
5. **Single user context**: Each instance serves one user in their own terminal session

---

## Why Common "Security Issues" Don't Apply

Many concepts that are security vulnerabilities in web applications or networked services are **not vulnerabilities** in a local terminal application like TFE.

### 1. "Command Injection" in Command Prompt

**Status: ✅ Not a vulnerability**

**Why:**
- User is already in a terminal with full shell access
- Users can run ANY command they want directly (e.g., `rm -rf ~`)
- Sanitizing commands would just add friction with no security benefit
- This is like saying `bash` has a "command injection vulnerability"

**Example:**
```bash
# In TFE command prompt
$ rm -rf ~/important_files

# This is equivalent to:
$ bash
$ rm -rf ~/important_files
```

**Conclusion**: The command prompt is a convenience feature, not a security boundary.

---

### 2. "Path Traversal" in File Navigation

**Status: ✅ Not a vulnerability**

**Why:**
- That's the entire point of a file browser
- Users can navigate to any path they have permissions for (e.g., `../../etc/passwd`)
- Blocking this would make TFE completely useless
- The OS enforces file permissions, not TFE

**Example:**
```bash
# User can navigate to:
/home/user/projects/
../../etc/passwd            # OS allows if user has read permission
/var/log/syslog            # OS allows if user has read permission
/root/.ssh/id_rsa          # OS denies if user lacks permission (Permission Denied)
```

**Conclusion**: TFE respects OS file permissions. If the user can access a file in their terminal, they can access it in TFE.

---

### 3. History/Favorites File Permissions (0644 vs 0600)

**Status: ✅ Minor privacy consideration - Not a security vulnerability**

**Why:**
- No privilege escalation possible
- No data theft from other users
- Other users can see command history, but they already share the system
- Users on shared systems can manually `chmod 600` if desired

**Files affected:**
- `~/.config/tfe/command_history.json` (0644)
- `~/.config/tfe/favorites.json` (0644)

**Risk level**: Very low
- **Impact**: Other users on the same system could read your command history
- **Likelihood**: Low (most systems are single-user)
- **Mitigation**: Users can manually run `chmod 600 ~/.config/tfe/*` if concerned

**Conclusion**: This is a privacy consideration, not a security flaw.

---

## Actual Security Bugs Fixed

These are real bugs that were fixed, though they were resource leaks rather than exploitable vulnerabilities:

### File Handle Leak (2025-10-24)

**Status: ✅ Fixed**

**Issue:**
- Some `os.Open()` calls were missing `defer file.Close()`
- This caused file handles to leak over time
- Could eventually exhaust system resources (file descriptor limit)

**Impact:**
- Resource exhaustion after opening many files
- No privilege escalation or data exposure
- Would eventually cause "too many open files" error

**Fix:**
- Added `defer file.Close()` to all file operations
- Files affected: `file_operations.go:119,647,2147` and `trash.go:347,354`

**Classification**: Resource leak bug, not a security vulnerability

---

## What TFE Does NOT Protect Against

TFE is **not designed to defend against**:

1. **Malicious user attacking themselves**: If a user runs `rm -rf ~` in the command prompt, TFE won't stop them
2. **Compromised terminal**: If the terminal emulator is malicious, it can intercept all TFE output
3. **Kernel/OS vulnerabilities**: TFE relies on OS file permission enforcement
4. **Physical access attacks**: Someone with physical access can do anything
5. **Multi-user attacks on the same account**: Users sharing the same account can read each other's TFE config

---

## What TFE DOES Protect Against

TFE does implement basic safety features:

1. **Trash instead of delete**: Files go to trash by default, not permanent deletion
2. **OS permission respect**: TFE will fail gracefully if user lacks permissions
3. **No privilege escalation**: TFE runs with user's privileges, never attempts to elevate
4. **Graceful error handling**: Permission denied errors are caught and displayed

---

## Security Best Practices Followed

While TFE doesn't need traditional "security hardening," it follows good practices:

1. **No eval() or arbitrary code execution**: Commands are passed directly to shell, not eval'd
2. **Resource cleanup**: All files are properly closed with defer
3. **Error handling**: Errors are caught and displayed, not ignored
4. **No hardcoded credentials**: No authentication system exists
5. **No network access**: TFE is entirely local (except optional git operations)

---

## Reporting Security Issues

If you believe you've found a security issue in TFE, please consider:

1. **Is it really a vulnerability?** Review this threat model first
2. **Can it be exploited remotely?** TFE has no network surface
3. **Does it allow privilege escalation?** TFE runs as the user
4. **Is it OS-level?** OS bugs should be reported to the OS vendor

**Real security issues** (privilege escalation, remote code execution, etc.) should be reported to:
- GitHub Issues: https://github.com/GGPrompts/TFE/issues
- Or via email if you prefer private disclosure

**Not security issues** (design decisions explained in this document):
- Command injection in command prompt
- Path traversal in file browser
- World-readable config files

---

## Comparison with Other Applications

| Application | User is Operator | Command Execution | File System Access | Security Model |
|-------------|------------------|-------------------|-------------------|----------------|
| **TFE** | Yes | Direct shell | Full user access | Trust user |
| **bash** | Yes | Direct shell | Full user access | Trust user |
| **ranger** | Yes | Direct shell | Full user access | Trust user |
| **Web browser** | No | Sandboxed | Restricted | Sandbox untrusted code |
| **Web server** | No | None (ideally) | Restricted | Defense in depth |

TFE's security model is similar to bash and ranger: trust the user, respect OS permissions.

---

## Summary

**TFE's threat model assumes the user is trustworthy and has legitimate access to their system.**

Common "vulnerabilities" in web applications (command injection, path traversal) are **features** in a terminal file manager. The operating system enforces security boundaries, not TFE.

Real security issues (privilege escalation, unauthorized access) would be taken seriously, but the application's design does not create exploitable attack surfaces.

**Bottom line**: If you wouldn't report it as a vulnerability in `bash` or `ls`, it's probably not a vulnerability in TFE.
