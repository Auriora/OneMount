# Logging Refactoring Implementation Plan

This document provides a detailed plan for implementing the recommendations in the [Logging Refactoring Report](logging-refactoring-report.md).

## Phase 1: Refactor Core Structure

### 1.1 Reorganize Files

1. Create new files with the consolidated structure:
   - `logger.go`: Core logger implementation and level management
   - `context.go`: Context-aware logging functionality
   - `method.go`: Method entry/exit logging (both with and without context)
   - `error.go`: Error logging functionality
   - `performance.go`: Performance optimization utilities

2. Move functionality from existing files to the new structure:
   - Move core logger implementation from `logger.go` to the new `logger.go`
   - Move level-related functionality from `level.go` to the new `logger.go`
   - Move context-related functionality from `context.go` to the new `context.go`
   - Move method logging from `method_logging.go` and `method_logging_context.go` to the new `method.go`
   - Move error logging from `error_logging.go` and `log_errors.go` to the new `error.go`
   - Move performance optimization from `log_performance.go` to the new `performance.go`

3. Ensure all constants are defined in a single location (either `logger.go` or a separate `constants.go` file)

### 1.2 Ensure Tests Pass

1. Run all existing tests to ensure they still pass after reorganization
2. Fix any failing tests
3. Add new tests for any functionality that isn't already covered

### 1.3 Update Documentation

1. Update package documentation to reflect the new structure
2. Update inline documentation in each file to explain its purpose and functionality

## Phase 2: Simplify API

### 2.1 Consolidate Error Logging Functions

1. Identify redundant error logging functions:
   - `LogError` vs `LogErrorWithFields`
   - `LogErrorAndReturn` vs `LogErrorWithContextAndReturn`
   - `WrapAndLog` vs `WrapfAndLog`

2. Consolidate into a smaller set of functions:
   - `LogError(err, msg, fields...)`: Basic error logging
   - `LogErrorWithContext(err, ctx, msg, fields...)`: Context-aware error logging
   - `WrapAndLogError(err, msg, fields...)`: Wrap, log, and return error
   - `WrapAndLogErrorWithContext(err, ctx, msg, fields...)`: Context-aware version

3. Implement the consolidated functions in `error.go`

### 2.2 Standardize Method Logging

1. Rename method logging functions for clarity:
   - `LogMethodCall` -> `LogMethodEntry`
   - `LogMethodReturn` -> `LogMethodExit`
   - `LogMethodCallWithContext` -> `LogMethodEntryWithContext`
   - `LogMethodReturnWithContext` -> `LogMethodExitWithContext`

2. Add helper functions for common patterns:
   - `WithMethodLogging(methodName, fn, params...)`: Execute function with logging
   - `WithMethodLoggingAndContext(methodName, ctx, fn, params...)`: Execute function with context-aware logging

3. Implement the standardized functions in `method.go`

### 2.3 Standardize Naming Conventions

1. Use capitalized names for all exported functions:
   - `isDebugEnabled` -> `IsDebugEnabled`
   - `isTraceEnabled` -> `IsTraceEnabled`

2. Use consistent prefixes for related functions:
   - All error logging functions start with `LogError` or `WrapAndLog`
   - All method entry logging functions start with `LogMethodEntry`
   - All method exit logging functions start with `LogMethodExit`

3. Use consistent parameter ordering across similar functions:
   - Error logging: `(err, msg, fields...)` or `(err, ctx, msg, fields...)`
   - Method logging: `(methodName, ...)` or `(methodName, ctx, ...)`

### 2.4 Update Tests

1. Update all tests to use the new API
2. Add tests for the new helper functions
3. Ensure all tests pass

### 2.5 Update Documentation

1. Update API documentation to reflect the simplified API
2. Update examples to show the recommended patterns
3. Add more inline documentation to explain function behavior

## Phase 3: Optimize Performance

### 3.1 Reduce Reflection Usage

1. Identify areas where reflection is used:
   - Method parameter logging
   - Return value logging

2. Implement alternatives to reduce reflection:
   - Add type-specific logging helpers for common types
   - Consider code generation for type-specific method logging

3. Enhance the type caching mechanism:
   - Expand the cache to cover more reflection operations
   - Add more aggressive caching strategies

### 3.2 Add Level Checks

1. Add level checks before expensive operations:
   - Add `IsLevelEnabled(level)` function to check if a specific level is enabled
   - Add level checks in all logging functions that perform expensive operations

2. Add helper functions for common patterns:
   - `LogIfEnabled(level, logFn)`: Execute logging function only if level is enabled
   - `LogComplexObjectIfEnabled(level, fieldName, obj, msg)`: Log complex object only if level is enabled

### 3.3 Benchmark Performance

1. Create benchmarks for common logging operations:
   - Basic logging
   - Structured logging
   - Method logging
   - Error logging

2. Run benchmarks before and after optimization to measure improvements
3. Document performance improvements

## Phase 4: Update Usage

### 4.1 Update Codebase

1. Update all usage of the logging package throughout the codebase:
   - Replace deprecated functions with their new equivalents
   - Update method logging to use the standardized pattern
   - Add level checks before expensive logging operations

2. Ensure consistent usage patterns:
   - Use context-aware logging for operations that span multiple functions
   - Use method logging for all public methods
   - Use appropriate log levels for different types of information

3. Add more context to logs where beneficial:
   - Add request IDs for operations that span multiple functions
   - Add user IDs for user-initiated operations
   - Add operation names for high-level operations

### 4.2 Final Testing

1. Run all tests to ensure they pass
2. Perform manual testing to verify logging behavior
3. Review logs to ensure they are useful and consistent

## Timeline

- Phase 1: 1-2 weeks
- Phase 2: 2-3 weeks
- Phase 3: 1-2 weeks
- Phase 4: 2-3 weeks

Total: 6-10 weeks

## Dependencies

- Access to all code that uses the logging package
- Ability to run tests for all affected code
- Coordination with other teams to ensure smooth transition

## Risks and Mitigations

### Risk: Breaking Changes

**Risk**: The API changes may break existing code that uses the logging package.

**Mitigation**:
- Maintain backward compatibility where possible
- Provide clear migration guides
- Update all usage in the codebase as part of the refactoring

### Risk: Performance Regression

**Risk**: Changes to the logging implementation may impact performance.

**Mitigation**:
- Benchmark before and after changes
- Focus on optimizing hot paths
- Add more aggressive level checks

### Risk: Inconsistent Usage

**Risk**: Developers may continue to use the old patterns even after refactoring.

**Mitigation**:
- Provide clear documentation and examples
- Update all existing code to use the new patterns
- Add linting rules to enforce consistent usage

## Success Criteria

The refactoring will be considered successful if:

1. All tests pass after refactoring
2. The API is simpler and more intuitive
3. Performance is maintained or improved
4. Logs are consistent and useful
5. Developers find the new API easier to use

## Conclusion

This implementation plan provides a detailed roadmap for refactoring the logging package according to the recommendations in the Logging Refactoring Report. By following this plan, we can create a more maintainable, user-friendly, and performant logging system that maintains all the current capabilities while being easier to use correctly.