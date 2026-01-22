# Authentication Token Path Consistency Audit

**Date**: YYYY-MM-DD  
**Auditor**: [Name/Role]  
**Task**: 45.4 - Authentication token path consistency audit  
**Status**: [In Progress / Complete]

---

## Executive Summary

Brief overview of audit scope, findings, and recommendations.

**Files Reviewed**: [Count]  
**Issues Found**: [Count]  
**Critical Issues**: [Count]  
**Corrections Made**: [Count]

---

## Audit Scope

### Code Files
- [ ] `internal/graph/oauth2.go` - Token creation and path logic
- [ ] `cmd/onemount/main.go` - Auth path determination
- [ ] `cmd/common/config.go` - Configuration handling
- [ ] `internal/testutil/test_constants.go` - Test paths
- [ ] Other Go files mentioning auth tokens

### Test Files
- [ ] `internal/graph/oauth2_test.go`
- [ ] `internal/fs/*_integration_test.go`
- [ ] `internal/fs/*_real_test.go`
- [ ] Other test files using auth tokens

### Scripts
- [ ] `scripts/setup-auth-reference.sh`
- [ ] `scripts/*.sh` - All shell scripts
- [ ] `tests/manual/*.sh` - Manual test scripts
- [ ] `docker/scripts/*.sh` - Docker helper scripts

### Documentation
- [ ] `docs/guides/developer/authentication-token-paths.md`
- [ ] `docs/TEST_SETUP.md`
- [ ] `docs/testing/*.md`
- [ ] `README.md`
- [ ] `CONTRIBUTING.md`
- [ ] Other documentation mentioning auth

### Configuration
- [ ] `.env.auth`
- [ ] `docker/compose/docker-compose.auth.yml`
- [ ] `configs/default-config.yml`

---

## Consistency Checks

### 1. Token Path Construction

**Expected Logic**:
```go
cachePath = config.CacheDir + "/" + unit.UnitNamePathEscape(absMountPath)
authPath = cachePath + "/auth_tokens.json"
```

**Findings**:
- [ ] Code implementation matches expected logic
- [ ] Documentation accurately describes logic
- [ ] Test fixtures use correct paths
- [ ] Scripts reference correct paths

**Issues**:
- [List any inconsistencies found]

---

### 2. Environment Variables

**Expected Variables**:
- `ONEMOUNT_AUTH_PATH` - Override for test auth token location
- `ONEMOUNT_AUTH_TOKEN_PATH` - Canonical host location (Docker)
- `ONEMOUNT_AUTH_PATH_DOCKER` - Container path (Docker)

**Findings**:
- [ ] Variable names consistent across all files
- [ ] Documentation matches actual usage
- [ ] Scripts use correct variable names
- [ ] Docker compose files reference correct variables

**Issues**:
- [List any inconsistencies found]

---

### 3. Default Locations

**Expected Defaults**:
- Production: `~/.cache/onemount/<instance>/auth_tokens.json`
- Tests: `test-artifacts/.auth_tokens.json` or `$ONEMOUNT_AUTH_PATH`
- Docker: `/tmp/auth-tokens/auth_tokens.json`

**Findings**:
- [ ] Code defaults match documentation
- [ ] Test defaults match documentation
- [ ] Docker defaults match documentation
- [ ] Scripts document their expected paths

**Issues**:
- [List any inconsistencies found]

---

### 4. Path Escaping

**Expected Behavior**:
- Use `unit.UnitNamePathEscape()` for mount path escaping
- Systemd-style escaping (e.g., `/tmp/test-mount` → `tmp-test-mount`)

**Findings**:
- [ ] Consistent escaping method used
- [ ] Documentation explains escaping
- [ ] Examples show escaped paths correctly

**Issues**:
- [List any inconsistencies found]

---

### 5. Error Messages

**Expected Behavior**:
- Error messages should include expected token path
- Clear guidance on where to find/place tokens

**Findings**:
- [ ] Authentication errors mention token path
- [ ] File not found errors show expected location
- [ ] Help text explains token location (where appropriate)

**Issues**:
- [List any inconsistencies found]

---

### 6. Documentation Accuracy

**Expected Content**:
- Accurate description of token creation process
- Correct path construction examples
- Valid troubleshooting guidance
- No outdated information

**Findings**:
- [ ] All documentation is accurate
- [ ] Examples are correct and tested
- [ ] No contradictory information
- [ ] Cross-references are valid

**Issues**:
- [List any inconsistencies found]

---

## Detailed Findings

### Critical Issues

#### Issue #1: [Title]
**Severity**: Critical  
**Location**: [File:Line]  
**Description**: [Detailed description]  
**Impact**: [What breaks or confuses users]  
**Recommendation**: [How to fix]  
**Status**: [Fixed / Pending]

### High Priority Issues

#### Issue #2: [Title]
**Severity**: High  
**Location**: [File:Line]  
**Description**: [Detailed description]  
**Impact**: [What breaks or confuses users]  
**Recommendation**: [How to fix]  
**Status**: [Fixed / Pending]

### Medium Priority Issues

#### Issue #3: [Title]
**Severity**: Medium  
**Location**: [File:Line]  
**Description**: [Detailed description]  
**Impact**: [What breaks or confuses users]  
**Recommendation**: [How to fix]  
**Status**: [Fixed / Pending]

### Low Priority Issues

#### Issue #4: [Title]
**Severity**: Low  
**Location**: [File:Line]  
**Description**: [Detailed description]  
**Impact**: [What breaks or confuses users]  
**Recommendation**: [How to fix]  
**Status**: [Fixed / Pending]

---

## Corrections Made

### Code Changes

1. **File**: [path]
   - **Change**: [Description]
   - **Reason**: [Why this was needed]
   - **Commit**: [hash if applicable]

2. **File**: [path]
   - **Change**: [Description]
   - **Reason**: [Why this was needed]
   - **Commit**: [hash if applicable]

### Documentation Updates

1. **File**: [path]
   - **Change**: [Description]
   - **Reason**: [Why this was needed]

2. **File**: [path]
   - **Change**: [Description]
   - **Reason**: [Why this was needed]

### Script Updates

1. **File**: [path]
   - **Change**: [Description]
   - **Reason**: [Why this was needed]

---

## Recommendations

### Immediate Actions

1. [Action item with priority]
2. [Action item with priority]
3. [Action item with priority]

### Future Improvements

1. [Longer-term improvement]
2. [Longer-term improvement]
3. [Longer-term improvement]

### Code Quality

1. [Code quality recommendation]
2. [Code quality recommendation]

### Documentation

1. [Documentation recommendation]
2. [Documentation recommendation]

---

## Verification

### Test Results

- [ ] All unit tests pass after corrections
- [ ] All integration tests pass after corrections
- [ ] Manual testing confirms corrections work
- [ ] Documentation examples verified

### Review Checklist

- [ ] All findings documented
- [ ] All corrections made and tested
- [ ] Documentation updated
- [ ] Cross-references validated
- [ ] No new inconsistencies introduced

---

## Appendix

### Files Reviewed

Complete list of all files reviewed during audit:

**Code Files** ([count]):
- [file path]
- [file path]

**Test Files** ([count]):
- [file path]
- [file path]

**Scripts** ([count]):
- [file path]
- [file path]

**Documentation** ([count]):
- [file path]
- [file path]

**Configuration** ([count]):
- [file path]
- [file path]

### Search Patterns Used

Patterns used to find relevant files:
```bash
# Auth token references
grep -r "auth.*token" --include="*.go" --include="*.sh" --include="*.md"

# Path construction
grep -r "GetAuthTokensPath\|AuthTokensPath" --include="*.go"

# Environment variables
grep -r "ONEMOUNT_AUTH" --include="*.sh" --include="*.yml" --include="*.md"

# Token file references
grep -r "auth_tokens\.json\|\.auth_tokens\.json" --include="*.go" --include="*.sh" --include="*.md"
```

### References

- Task 45.1: Manual D-Bus integration testing (discovered token confusion)
- `docs/guides/developer/authentication-token-paths.md` - Comprehensive guide
- `internal/graph/oauth2.go` - Token creation implementation
- `cmd/onemount/main.go` - Path determination logic
