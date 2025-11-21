# CI/CD Documentation

## Overview

This document describes the Continuous Integration and Continuous Deployment (CI/CD) pipeline for the Go backend application. The pipeline is built using GitHub Actions and supports automated testing, building, security scanning, and deployment.

## Table of Contents

1. [Workflows](#workflows)
2. [Deployment Process](#deployment-process)
3. [Environment Setup](#environment-setup)
4. [Secret Management](#secret-management)
5. [Rollback Procedures](#rollback-procedures)
6. [Troubleshooting](#troubleshooting)

## Workflows

### 1. CI Pipeline (`ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop`

**Jobs:**
- **Lint**: Runs golangci-lint for code quality checks
- **Test**: Executes tests with PostgreSQL service
  - Runs on Go versions 1.22 and 1.23
  - Generates coverage report
  - Requires minimum 60% test coverage
  - Uploads coverage to Codecov
- **Build**: Compiles binaries for multiple platforms
  - Linux, Darwin, Windows
  - amd64 and arm64 architectures
  - Uploads artifacts for download
- **Notify**: Sends build status to Slack

**Status Badge:**
```markdown
![CI Pipeline](https://github.com/your-org/backend/workflows/CI%20Pipeline/badge.svg)
```

### 2. Security Scan (`security.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop`
- Weekly schedule (Sundays at midnight)
- Manual trigger via workflow_dispatch

**Jobs:**
- **Vulnerability Scan**: Uses govulncheck to detect known vulnerabilities
- **Secret Scan**: TruffleHog scans for leaked secrets
- **Static Analysis**: gosec performs security-focused static analysis
- **Dependency Review**: Checks dependencies for security issues (PR only)
- **CodeQL Analysis**: Deep semantic code analysis

**Automated Actions:**
- Creates GitHub issues for vulnerabilities found during scheduled scans
- Uploads SARIF reports to GitHub Security tab

### 3. Docker Build (`docker.yml`)

**Triggers:**
- Push to `main` branch
- Version tags (e.g., `v1.0.0`)
- Manual trigger

**Features:**
- Multi-architecture builds (amd64, arm64)
- Pushes to GitHub Container Registry and Docker Hub
- Uses Docker BuildKit caching for faster builds
- Runs Trivy vulnerability scanner on images
- Generates SBOM (Software Bill of Materials)
- Auto-updates Docker Hub description

**Image Tags:**
- `latest` (main branch)
- `v1.0.0` (semantic version)
- `v1.0` (major.minor)
- `v1` (major)
- `main-abc1234` (branch-sha)

### 4. Deployment (`deploy.yml`)

**Triggers:**
- Manual workflow dispatch with environment selection
- Push to version tags (auto-deploys to staging)

**Environments:**
- **Staging**: Automatic deployment, no approval required
- **Production**: Requires manual approval, blue-green deployment

**Deployment Steps:**
1. Set up environment configuration
2. Run database migrations
3. Deploy application
4. Run health checks
5. Switch traffic (production only)
6. Run smoke tests
7. Clean up old deployments
8. Rollback on failure

**Approval Process:**
- Staging: No approval required
- Production: Requires approval from designated reviewers

### 5. Database Migration (`migrate.yml`)

**Triggers:**
- Manual workflow dispatch with environment and action selection

**Actions:**
- `up`: Apply pending migrations
- `down`: Rollback migrations (specify steps)
- `status`: Show current migration version
- `create`: Generate new migration files

**Safety Features:**
- Validates migration file format
- Creates backup before production migrations
- Verifies database integrity after migration
- Creates GitHub issue on failure
- Supports rollback on failure

## Deployment Process

### Staging Deployment

1. Push code to `main` branch or create a version tag
2. CI pipeline runs automatically
3. If CI passes, Docker image is built and pushed
4. Staging deployment is triggered automatically
5. Health checks verify deployment success

```bash
# Example: Deploy specific version to staging
gh workflow run deploy.yml -f environment=staging -f version=v1.0.0
```

### Production Deployment

1. Ensure staging deployment is successful
2. Trigger production deployment manually
3. Approve deployment in GitHub UI
4. Blue-green deployment executes
5. Health checks validate new deployment
6. Traffic switches to new version
7. Old version is kept for quick rollback

```bash
# Example: Deploy to production
gh workflow run deploy.yml -f environment=production -f version=v1.0.0
```

### Deployment Flow Diagram

```
┌─────────────┐
│ Code Change │
└──────┬──────┘
       │
       ▼
┌──────────────┐
│  CI Pipeline │ ◄── Tests, Build, Security
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Docker Build │ ◄── Multi-arch images
└──────┬───────┘
       │
       ├──────────────┐
       ▼              ▼
┌──────────┐   ┌─────────────┐
│ Staging  │   │ Production  │
│ (Auto)   │   │ (Approval)  │
└──────────┘   └─────────────┘
       │              │
       ▼              ▼
┌──────────────────────────┐
│   Health Check & Tests   │
└──────────────────────────┘
```

## Environment Setup

### Required GitHub Secrets

#### Staging
```
STAGING_DATABASE_URL
STAGING_HOST
STAGING_DEPLOY_KEY
CODECOV_TOKEN (optional)
SLACK_WEBHOOK_URL (optional)
```

#### Production
```
PRODUCTION_DATABASE_URL
PRODUCTION_HOST
PRODUCTION_DEPLOY_KEY
DOCKERHUB_USERNAME (optional)
DOCKERHUB_TOKEN (optional)
SLACK_WEBHOOK_URL (optional)
PAGERDUTY_INTEGRATION_KEY (optional)
```

### Setting Up Secrets

1. Navigate to repository Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add each required secret

```bash
# Using GitHub CLI
gh secret set PRODUCTION_DATABASE_URL -b "postgres://..."
```

### Environment Variables

Copy and customize environment configuration:

```bash
# Staging
cp config/staging.env.example config/staging.env
# Edit config/staging.env with actual values

# Production
cp config/production.env.example config/production.env
# Edit config/production.env with actual values
```

## Secret Management

### Best Practices

1. **Never commit secrets** to version control
2. Use **GitHub Secrets** for CI/CD pipelines
3. Rotate secrets regularly (every 90 days)
4. Use different secrets for each environment
5. Limit secret access to necessary personnel
6. Use **encrypted secrets** for production
7. Enable **secret scanning** in repository settings

### Secret Rotation

```bash
# Example: Rotate JWT secret
./scripts/rotate-secret.sh production JWT_SECRET

# Update GitHub secret
gh secret set JWT_SECRET -b "new-secret-value"

# Deploy with new secret
gh workflow run deploy.yml -f environment=production -f version=current
```

## Rollback Procedures

### Automatic Rollback

Deployments automatically rollback on:
- Failed health checks
- Smoke test failures
- Container startup failures

### Manual Rollback

#### Using Workflow

```bash
# Trigger rollback workflow
gh workflow run deploy.yml -f environment=production -f version=v1.0.0
```

#### Using Rollback Script

```bash
# SSH to server
ssh production-server

# Run rollback script
cd /app
./scripts/rollback.sh production

# Verify rollback
./scripts/health-check.sh production
```

### Rollback Steps

1. **Identify issue**: Check logs and monitoring
2. **Confirm rollback**: Verify previous version
3. **Execute rollback**: Use script or workflow
4. **Verify health**: Run health checks
5. **Rollback migrations**: If necessary
6. **Notify team**: Send rollback notification
7. **Investigate**: Find and fix root cause

### Database Rollback

```bash
# Check current migration version
./scripts/migrate.sh production status

# Rollback N migrations
./scripts/migrate.sh production down N

# Verify database state
./scripts/health-check.sh production
```

## Troubleshooting

### Common Issues

#### 1. Build Failures

**Symptom**: CI pipeline fails during build
**Solution**:
```bash
# Check build logs
gh run view --log

# Run build locally
go build -o backend ./cmd/server

# Check for missing dependencies
go mod verify
go mod tidy
```

#### 2. Test Failures

**Symptom**: Tests fail in CI but pass locally
**Solution**:
```bash
# Run tests with race detector
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Check database connection
docker-compose up -d postgres
export DATABASE_URL="postgres://testuser:testpass@localhost:5432/testdb"
go test ./...
```

#### 3. Low Test Coverage

**Symptom**: CI fails due to < 60% coverage
**Solution**:
```bash
# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Add missing tests
# Re-run CI
```

#### 4. Deployment Failures

**Symptom**: Deployment fails during health check
**Solution**:
```bash
# Check container logs
docker logs backend-production

# Run health check manually
./scripts/health-check.sh production

# Check environment variables
docker exec backend-production env | grep DATABASE_URL

# Rollback if necessary
./scripts/rollback.sh production
```

#### 5. Migration Failures

**Symptom**: Database migration fails
**Solution**:
```bash
# Check migration status
./scripts/migrate.sh production status

# Validate migration files
ls -l db/migrations/

# Force migration version (use with caution)
migrate -database "$DATABASE_URL" -path db/migrations force VERSION

# Rollback migration
./scripts/migrate.sh production down 1
```

#### 6. Docker Build Failures

**Symptom**: Docker build fails or times out
**Solution**:
```bash
# Build locally with verbose output
docker build -t backend:test . --no-cache --progress=plain

# Check Dockerfile syntax
docker run --rm -i hadolint/hadolint < Dockerfile

# Verify .dockerignore
cat .dockerignore
```

### Monitoring and Alerts

#### Check Workflow Status

```bash
# List recent workflow runs
gh run list --workflow=ci.yml

# View specific run
gh run view RUN_ID --log

# Watch workflow in real-time
gh run watch RUN_ID
```

#### Notifications

Configure Slack notifications:
1. Create Slack webhook: https://api.slack.com/messaging/webhooks
2. Add webhook URL to GitHub secrets: `SLACK_WEBHOOK_URL`
3. Notifications will be sent for:
   - CI pipeline status
   - Security scan issues
   - Deployment status
   - Rollback events

### Getting Help

1. Check workflow logs: `gh run view --log`
2. Review documentation: This file
3. Check GitHub Actions status: https://www.githubstatus.com/
4. Contact DevOps team: #devops-support

## Maintenance

### Regular Tasks

- **Weekly**: Review security scan results
- **Monthly**: Update dependencies and base images
- **Quarterly**: Rotate secrets and review access
- **Yearly**: Review and update CI/CD pipeline

### Pipeline Updates

```bash
# Update workflow files in .github/workflows/
git add .github/workflows/
git commit -m "Update CI/CD pipeline"
git push

# Test changes in feature branch first
git checkout -b update-cicd
# Make changes
git push -u origin update-cicd
# Create PR and verify workflows
```

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Documentation](https://docs.docker.com/)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [PostgreSQL Migrations](https://github.com/golang-migrate/migrate)

---

**Last Updated**: 2025-11-21
**Maintained By**: DevOps Team
