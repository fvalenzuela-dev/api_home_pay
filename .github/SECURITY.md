# Security Policy

## Reporting Security Vulnerabilities

Please report security vulnerabilities to the repository maintainer directly.

## Known False Positives

The following items are **false positives** and do not represent real security vulnerabilities:

### Test Credentials

Several files contain test credentials for **local development and testing only**:

- `cmd/api/main_test.go`
- `internal/config/config_test.go`
- `internal/database/database_test.go`
- `internal/repository/*_integration_test.go`

These files use hardcoded credentials like `postgres://test:test@localhost:5432/test` which are:
- **Local test databases only** - not connected to any production system
- **Intentionally invalid or test data** - used for unit/integration testing
- **Never exposed** - only used in test environments

### Example

```go
// This is a test database URL - not a real credential
t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
```

These credentials:
- Cannot connect to any production system
- Are intentionally weak/dummy values
- Are only used in `testing.Short()` or local test scenarios

## GitHub Secret Scanning

The repository has GitHub Advanced Security enabled. Secret scanning alerts for these test credentials can be safely ignored as they do not represent real secrets.
