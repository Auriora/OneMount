# Code Review Checklist

**Version**: 1.0  
**Last Updated**: 2025-01-21  
**Purpose**: Ensure consistent, thorough code reviews

## Overview

This checklist helps reviewers ensure that code changes meet project standards for quality, documentation, testing, and maintainability.

## General Review Process

1. **Understand the Context**
   - Read the PR description
   - Review linked issues
   - Understand the problem being solved

2. **Review the Code**
   - Check code quality
   - Verify functionality
   - Look for potential issues

3. **Verify Documentation**
   - Check documentation updates
   - Verify godoc comments
   - Review ADRs if applicable

4. **Check Tests**
   - Verify test coverage
   - Review test quality
   - Ensure tests pass

5. **Provide Feedback**
   - Be constructive
   - Explain reasoning
   - Suggest improvements

## Code Quality Checklist

### Functionality

- [ ] Code solves the stated problem
- [ ] Code handles edge cases appropriately
- [ ] Error handling is comprehensive
- [ ] No obvious bugs or logic errors
- [ ] Code is efficient and performant

### Code Style

- [ ] Code follows project style guidelines
- [ ] Variable names are descriptive
- [ ] Functions are appropriately sized
- [ ] Code is properly formatted
- [ ] No unnecessary complexity

### Best Practices

- [ ] Follows SOLID principles
- [ ] Follows DRY principle
- [ ] No code duplication
- [ ] Appropriate use of design patterns
- [ ] Proper separation of concerns

### Security

- [ ] No security vulnerabilities
- [ ] No hardcoded credentials
- [ ] Proper input validation
- [ ] No SQL injection risks
- [ ] No XSS vulnerabilities

### Concurrency (if applicable)

- [ ] Proper use of mutexes/locks
- [ ] No race conditions
- [ ] No deadlock potential
- [ ] Goroutines properly managed
- [ ] Wait groups used for cleanup

## Documentation Checklist

### Architecture Documentation

- [ ] Architecture docs updated for structural changes
- [ ] Component diagrams updated
- [ ] Component interactions documented
- [ ] Interfaces documented
- [ ] ADR created for significant decisions

### Design Documentation

- [ ] Design docs updated for data model changes
- [ ] Class diagrams updated
- [ ] Design patterns documented
- [ ] API signatures current
- [ ] Design rationale documented

### API Documentation (Godoc)

- [ ] All public APIs have godoc comments
- [ ] Godoc comments start with function name
- [ ] Parameter descriptions included
- [ ] Return value descriptions included
- [ ] Error conditions documented
- [ ] Examples provided for complex APIs

**Godoc Comment Template**:
```go
// FunctionName performs [action] on [object].
//
// This function [detailed description of behavior].
//
// Parameters:
//   - param1: Description of param1
//   - param2: Description of param2
//
// Returns:
//   - returnValue: Description of return value
//   - error: Description of error conditions
//
// Example:
//   result, err := FunctionName(arg1, arg2)
//   if err != nil {
//       // Handle error
//   }
//   // Use result
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error)
```

### User Documentation

- [ ] User docs updated for behavior changes
- [ ] Installation instructions current
- [ ] Configuration guides current
- [ ] Troubleshooting guides updated
- [ ] FAQ updated if needed

### Code Comments

- [ ] Complex logic is commented
- [ ] Non-obvious decisions explained
- [ ] TODO/FIXME comments have issue numbers
- [ ] Comments are accurate and helpful
- [ ] No commented-out code

## Testing Checklist

### Test Coverage

- [ ] New code has unit tests
- [ ] New code has integration tests (if applicable)
- [ ] Edge cases are tested
- [ ] Error conditions are tested
- [ ] Test coverage is adequate (>80%)

### Test Quality

- [ ] Tests are clear and readable
- [ ] Tests are independent
- [ ] Tests are deterministic
- [ ] Tests use descriptive names
- [ ] Tests follow AAA pattern (Arrange, Act, Assert)

### Test Execution

- [ ] All tests pass locally
- [ ] All tests pass in CI/CD
- [ ] No flaky tests
- [ ] Tests run in reasonable time
- [ ] Race detector passes (for concurrent code)

### Test Types

**Unit Tests**:
- [ ] Test individual functions/methods
- [ ] Mock external dependencies
- [ ] Fast execution (<1s per test)
- [ ] High coverage of code paths

**Integration Tests**:
- [ ] Test component interactions
- [ ] Use real dependencies where possible
- [ ] Test realistic scenarios
- [ ] Reasonable execution time (<30s per test)

**Property-Based Tests** (if applicable):
- [ ] Properties clearly defined
- [ ] Generators produce valid inputs
- [ ] Sufficient iterations (100+)
- [ ] Shrinking works correctly

## Performance Checklist

### Efficiency

- [ ] No unnecessary allocations
- [ ] Efficient algorithms used
- [ ] No N+1 query problems
- [ ] Appropriate data structures
- [ ] No performance regressions

### Resource Usage

- [ ] Memory usage is reasonable
- [ ] No memory leaks
- [ ] File descriptors properly closed
- [ ] Goroutines properly cleaned up
- [ ] Database connections properly managed

### Scalability

- [ ] Code scales with data size
- [ ] No hardcoded limits
- [ ] Appropriate use of pagination
- [ ] Efficient for large datasets
- [ ] No blocking operations in hot paths

## Maintainability Checklist

### Code Organization

- [ ] Code is in appropriate package
- [ ] File organization is logical
- [ ] Related code is grouped together
- [ ] No circular dependencies
- [ ] Clear module boundaries

### Readability

- [ ] Code is easy to understand
- [ ] Logic flow is clear
- [ ] Variable names are descriptive
- [ ] Functions have single responsibility
- [ ] No magic numbers or strings

### Extensibility

- [ ] Code is easy to extend
- [ ] Interfaces used appropriately
- [ ] Dependency injection used
- [ ] Configuration is externalized
- [ ] No tight coupling

## Specific Component Checklists

### Filesystem Changes

- [ ] FUSE operations implemented correctly
- [ ] Inode management is correct
- [ ] File operations are atomic
- [ ] Directory operations are consistent
- [ ] Permissions handled correctly

### Graph API Changes

- [ ] Authentication handled correctly
- [ ] Token refresh implemented
- [ ] Error handling is comprehensive
- [ ] Rate limiting respected
- [ ] Retry logic implemented

### Cache Changes

- [ ] Cache invalidation is correct
- [ ] Cache consistency maintained
- [ ] Cache cleanup implemented
- [ ] Cache size limits enforced
- [ ] Cache statistics updated

### State Management Changes

- [ ] State transitions are valid
- [ ] State is persisted correctly
- [ ] State recovery works
- [ ] Concurrent state access is safe
- [ ] State machine is documented

### UI Changes

- [ ] UI is responsive
- [ ] Error messages are clear
- [ ] User feedback is provided
- [ ] Accessibility considered
- [ ] Desktop integration works

## Common Issues to Watch For

### Concurrency Issues

- [ ] Race conditions
- [ ] Deadlocks
- [ ] Goroutine leaks
- [ ] Improper mutex usage
- [ ] Missing synchronization

### Error Handling Issues

- [ ] Errors not checked
- [ ] Errors silently ignored
- [ ] Insufficient error context
- [ ] Improper error wrapping
- [ ] Missing error logging

### Resource Management Issues

- [ ] File descriptors not closed
- [ ] Memory leaks
- [ ] Database connections not closed
- [ ] Goroutines not cleaned up
- [ ] Temporary files not deleted

### Security Issues

- [ ] Hardcoded credentials
- [ ] Insufficient input validation
- [ ] SQL injection vulnerabilities
- [ ] XSS vulnerabilities
- [ ] Insecure file permissions

### Performance Issues

- [ ] Inefficient algorithms
- [ ] Unnecessary allocations
- [ ] N+1 query problems
- [ ] Blocking operations
- [ ] Missing indexes

## Review Feedback Guidelines

### Providing Feedback

**Be Constructive**:
- Focus on the code, not the person
- Explain why something should change
- Suggest specific improvements
- Acknowledge good work

**Be Clear**:
- Use clear, specific language
- Provide examples when helpful
- Link to relevant documentation
- Explain the impact of issues

**Be Respectful**:
- Assume good intentions
- Ask questions rather than making demands
- Offer to discuss complex issues
- Thank the author for their work

### Feedback Categories

**Critical** (Must Fix):
- Security vulnerabilities
- Data loss risks
- Breaking changes
- Major bugs

**Important** (Should Fix):
- Performance issues
- Maintainability concerns
- Missing tests
- Incomplete documentation

**Suggestion** (Nice to Have):
- Code style improvements
- Refactoring opportunities
- Additional tests
- Documentation enhancements

**Question** (Clarification):
- Unclear logic
- Missing context
- Design decisions
- Alternative approaches

### Example Feedback

**Good**:
> "This function could cause a race condition when accessed concurrently. Consider adding a mutex to protect the shared state. See the concurrency guidelines in docs/guides/developer/concurrency-guidelines.md for examples."

**Bad**:
> "This is wrong. Fix it."

## Approval Criteria

### Required for Approval

- [ ] All critical issues resolved
- [ ] All important issues resolved or acknowledged
- [ ] Tests pass
- [ ] Documentation updated
- [ ] No security concerns
- [ ] No performance regressions

### Optional for Approval

- [ ] All suggestions addressed
- [ ] All questions answered
- [ ] Code style perfect
- [ ] 100% test coverage

## Post-Review Actions

### After Approval

1. **Merge the PR**
   - Ensure CI/CD passes
   - Use appropriate merge strategy
   - Delete feature branch

2. **Monitor Deployment**
   - Watch for errors in production
   - Monitor performance metrics
   - Be available for questions

3. **Update Documentation**
   - Ensure all docs are current
   - Update changelog if needed
   - Notify relevant stakeholders

### After Rejection

1. **Provide Clear Feedback**
   - Explain why PR was rejected
   - List required changes
   - Offer to help with fixes

2. **Support the Author**
   - Answer questions
   - Provide guidance
   - Review again promptly

## Continuous Improvement

### Review Metrics

Track and improve:
- Average review time
- Number of review iterations
- Common issues found
- Documentation quality

### Learning Opportunities

- Share interesting PRs in team meetings
- Discuss complex review decisions
- Update checklist based on learnings
- Celebrate good code and reviews

## Conclusion

Thorough code reviews improve code quality, catch bugs early, and help the team learn. Use this checklist to ensure consistent, comprehensive reviews.

**Questions or Suggestions?**

Contact the development team or open an issue in the project repository.
