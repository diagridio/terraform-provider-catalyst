# Acceptance Tests

This directory contains acceptance tests for the Catalyst Terraform Provider.

## Prerequisites

Before running acceptance tests, you need:

1. **CATALYST_API_KEY**: A valid service account API key for Catalyst
2. **CATALYST_API_ENDPOINT**: The API endpoint (e.g., `https://api.local.diagrid.io`)
3. **CATALYST_PROJECT_ID**: An existing project ID to use for tests that require a project
4. **CATALYST_TLS_SKIP_VERIFY** (optional): Set to `true` or `1` to skip TLS certificate verification for self-signed certificates

## Running Tests

### Set Environment Variables

```bash
export TF_ACC=1
export CATALYST_API_KEY='your-api-key-here'
export CATALYST_API_ENDPOINT='https://api.local.diagrid.io'
export CATALYST_PROJECT_ID='your-project-id'
export CATALYST_TLS_SKIP_VERIFY=true  # Optional: for self-signed certificates
```

### Run All Tests

```bash
go test -v ./internal/test/acceptance/ -timeout 30m
```

### Run Specific Test

```bash
go test -v ./internal/test/acceptance/ -run TestAccProjectResource -timeout 5m
```

### Run Tests by Category

```bash
# Resource tests only
go test -v ./internal/test/acceptance/ -run TestAcc.*Resource -timeout 30m

# Data source tests only
go test -v ./internal/test/acceptance/ -run TestAcc.*DataSource -timeout 30m
```
