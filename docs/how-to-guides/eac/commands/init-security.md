# Init Command Security Guide

## Overview

This guide covers security best practices for managing API keys and configuring AI providers with the `init` command. Proper API key management is critical to protect your credentials and prevent unauthorized access to AI services.

## Getting API Keys

### Anthropic API Key (claude-api)

1. Visit [console.anthropic.com](https://console.anthropic.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Click "Create Key"
5. Copy the key (starts with `sk-ant-api03-`)
6. Set environment variable: `ANTHROPIC_API_KEY`

**Pricing:** Pay-per-use based on tokens consumed. See [Anthropic Pricing](https://www.anthropic.com/pricing).

**Security Notes:**

- API keys are shown only once during creation
- Store the key immediately in a secure location
- Never share keys via email, chat, or version control
- Each developer should have their own key

### Claude Pro Subscription (claude-cli)

1. Visit [claude.ai](https://claude.ai/)
2. Subscribe to Claude Pro
3. Install Claude CLI (if needed)
4. Log in to Claude CLI
5. No API key required

**Pricing:** Fixed monthly subscription fee.

**Security Notes:**

- Authentication is handled through Claude CLI
- No API key management required
- Each developer needs their own subscription
- More secure than managing API keys

### OpenAI API Key (openai)

1. Visit [platform.openai.com](https://platform.openai.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Click "Create new secret key"
5. Copy the key (starts with `sk-proj-` or `sk-`)
6. Set environment variable: `OPENAI_API_KEY`

**Pricing:** Pay-per-use based on tokens consumed. See [OpenAI Pricing](https://openai.com/pricing).

**Security Notes:**

- Keys are shown only once during creation
- Use project-scoped keys when possible
- Set usage limits in OpenAI dashboard
- Rotate keys regularly

### Google API Key (gemini)

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with Google account
3. Click "Create API Key"
4. Copy the key (starts with `AIza`)
5. Set environment variable: `GOOGLE_API_KEY`

**Pricing:** Free tier available, then pay-per-use. See [Gemini Pricing](https://ai.google.dev/pricing).

**Security Notes:**

- Restrict API key to specific IPs when possible
- Use API restrictions in Google Cloud Console
- Monitor usage in Google Cloud Console
- Enable billing alerts

## Security Best Practices

### Critical Security Rules

1. **Never commit API keys to version control**
2. **Never hardcode API keys in source code**
3. **Never share API keys via email, chat, or documentation**
4. **Always use environment variables for API keys**
5. **Always add configuration files to .gitignore**

### Protecting API Keys

```bash
# Add to .gitignore BEFORE initializing
echo ".r2r/eac/eac-config.yml" >> .gitignore

# Verify it's not tracked
git status .r2r/

# If accidentally committed, remove from history
git rm --cached .r2r/eac/eac-config.yml
git commit -m "Remove sensitive config file"
```

### Key Rotation

Rotate API keys regularly to minimize security risks:

```bash
# 1. Generate new key from provider dashboard
# 2. Update environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-NEW-KEY..."

# 3. Test new key
eac work commit --all --debug

# 4. Revoke old key in provider dashboard
# 5. Update CI/CD secrets if applicable
```

**Recommended rotation schedule:**

- Development keys: Every 90 days
- Production keys: Every 30-60 days
- After team member departure: Immediately
- After suspected compromise: Immediately

### Separate Keys for Environments

Use different API keys for different environments:

```bash
# Development
export ANTHROPIC_API_KEY="sk-ant-api03-dev-..."

# Staging
export ANTHROPIC_API_KEY="sk-ant-api03-staging-..."

# Production
export ANTHROPIC_API_KEY="sk-ant-api03-prod-..."
```

**Benefits:**

- Isolate usage and costs by environment
- Revoke compromised keys without affecting other environments
- Set different rate limits per environment
- Track usage separately

## API Key Management

### Secure Storage Options

**For Development:**

1. **Environment Variables (Recommended)**
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export ANTHROPIC_API_KEY="sk-ant-api03-..."
   ```

2. **Password Manager**
   - Store keys in 1Password, LastPass, or similar
   - Copy when needed, never save to disk
   - Use secure notes feature

3. **OS Keychain**
   - macOS Keychain Access
   - Windows Credential Manager
   - Linux Secret Service

**For Production:**

1. **CI/CD Secrets**
   - GitHub Secrets
   - GitLab CI/CD Variables
   - Azure Key Vault
   - AWS Secrets Manager

2. **Environment Management**
   - Docker secrets
   - Kubernetes secrets
   - HashiCorp Vault

### Never Store Keys In

- Source code files
- Configuration files committed to git
- Shell history
- Log files
- Error messages
- Documentation
- README files
- Comments in code

### Checking for Exposed Keys

```bash
# Check if config file is tracked by git
git ls-files .r2r/eac/eac-config.yml

# If tracked, remove it
git rm --cached .r2r/eac/eac-config.yml

# Check git history for exposed keys
git log -p | grep -i "api.*key"

# If keys found in history, consider them compromised
# Rotate immediately and use git-filter-repo to clean history
```

## Team Security Guidelines

### Individual Developer Setup

Each team member should:

1. Obtain their own API key (never share)
2. Set up environment variables on their machine
3. Initialize EAC with their credentials
4. Never commit configuration files

```bash
# Developer A
export ANTHROPIC_API_KEY="sk-ant-api03-devA-..."
eac init --ai claude-api

# Developer B
export ANTHROPIC_API_KEY="sk-ant-api03-devB-..."
eac init --ai claude-api
```

### Team Best Practices

1. **Document provider choice** in team wiki (not keys!)
2. **Share .gitignore rules** to exclude config files
3. **Use separate keys** per developer
4. **Establish rotation policy** for production keys
5. **Monitor usage** to detect anomalies
6. **Review access** when team members leave

### Onboarding New Team Members

```bash
# 1. Add .gitignore rule (if not already present)
echo ".r2r/eac/eac-config.yml" >> .gitignore
git add .gitignore
git commit -m "Ensure AI config is ignored"

# 2. Direct new member to get their own key
# 3. Share provider choice and setup instructions
# 4. New member initializes on their machine
export ANTHROPIC_API_KEY="sk-ant-api03-..."
eac init --ai claude-api
```

### Offboarding Team Members

When team members leave:

1. **Immediately revoke their personal API keys** in provider dashboard
2. **Rotate shared/production keys** they had access to
3. **Update CI/CD secrets** with new keys
4. **Review recent usage** for anomalies
5. **Document the rotation** in security log

## Secure Environment Setup

### macOS/Linux

```bash
# Add to ~/.bashrc or ~/.zshrc
# NEVER commit this file if it's in your project directory
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Secure the file
chmod 600 ~/.bashrc

# Reload
source ~/.bashrc

# Verify
echo $ANTHROPIC_API_KEY
```

### Windows (PowerShell Profile)

```powershell
# Edit PowerShell profile
notepad $PROFILE

# Add to profile
$env:ANTHROPIC_API_KEY = "sk-ant-api03-..."

# Secure the profile file (restrict access)
$acl = Get-Acl $PROFILE
$acl.SetAccessRuleProtection($true, $false)
Set-Acl $PROFILE $acl

# Reload
. $PROFILE
```

### Docker

```dockerfile
# Dockerfile - DO NOT hardcode keys
FROM alpine
# API key passed at runtime via --env

# docker-compose.yml
services:
  app:
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}

# .env file (add to .gitignore!)
ANTHROPIC_API_KEY=sk-ant-api03-...
```

### GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Initialize AI provider
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: eac init --ai claude-api
```

**Security:**

- Store keys in repository Settings > Secrets
- Use environment-specific secrets when possible
- Restrict secret access to necessary workflows
- Audit secret access regularly

## Security Checklist

Use this checklist when setting up AI provider configuration:

### Initial Setup

- [ ] Generated API key from provider dashboard
- [ ] Stored key in password manager or secure location
- [ ] Added `.r2r/eac/eac-config.yml` to `.gitignore`
- [ ] Set environment variable (not hardcoded)
- [ ] Ran `eac init --ai <provider>`
- [ ] Verified configuration file is not tracked by git
- [ ] Tested AI features work correctly
- [ ] Deleted any temporary files containing keys

### Team Setup

- [ ] Each team member has their own API key
- [ ] Documented provider choice (without keys) in team wiki
- [ ] Shared .gitignore configuration with team
- [ ] Established key rotation policy
- [ ] Set up usage monitoring
- [ ] Configured usage/spending limits in provider dashboard

### Production Setup

- [ ] Using separate API key for production
- [ ] Stored production key in secrets manager
- [ ] Configured CI/CD to use secret environment variables
- [ ] Set up usage alerts and monitoring
- [ ] Documented key rotation procedure
- [ ] Established incident response plan for compromised keys
- [ ] Restricted API key permissions (if supported by provider)

### Ongoing Maintenance

- [ ] Review API usage monthly
- [ ] Rotate keys according to schedule
- [ ] Update documentation when processes change
- [ ] Audit team member access quarterly
- [ ] Remove keys for departed team members
- [ ] Check for exposed keys in git history

## Emergency: Compromised API Key

If you believe an API key has been compromised:

1. **Immediately revoke the key** in provider dashboard
2. **Generate a new key**
3. **Update all systems** using the compromised key
4. **Review recent usage** for unauthorized activity
5. **Check billing** for unexpected charges
6. **Investigate** how the key was exposed
7. **Document** the incident and remediation
8. **Update processes** to prevent recurrence

```bash
# Quick response procedure
# 1. Revoke old key in provider dashboard
# 2. Generate new key
# 3. Update environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-NEW-KEY..."

# 4. Re-initialize
eac init --ai claude-api

# 5. Update CI/CD secrets
# 6. Test functionality
eac work commit --all --debug

# 7. Monitor for unusual activity
```

## See Also

- [Init Command Guide](init-command.md) - Quick start and basic usage
- [Init Reference](../../../reference/commands/init-reference.md) - Complete technical details
- [Anthropic Security Best Practices](https://docs.anthropic.com/claude/docs/security-best-practices)
- [OpenAI API Key Safety](https://platform.openai.com/docs/guides/safety-best-practices)
