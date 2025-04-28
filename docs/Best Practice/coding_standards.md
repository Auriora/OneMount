# onedriver Coding Standards

This document serves as the main entry point for all coding standards and best practices in the onedriver project. It provides an overview of the standards and links to more detailed documents for specific aspects of development.

## Overview

The onedriver project follows industry best practices for software development, with a focus on code quality, maintainability, and performance. These standards ensure consistency across the codebase and make it easier for new developers to contribute to the project.

## Language-Specific Standards

### Go

The majority of the onedriver codebase is written in Go. For Go code, refer to the following documents:

- [Go Coding Standards](go_coding_standards.md) - Comprehensive guide for Go code organization, naming conventions, error handling, and more
- [Go Logging Best Practices](go_logging_best_practices.md) - Guidelines for structured logging with zerolog
- [Test Best Practices](test_best_practices.md) - Best practices for writing effective tests

### Shell Scripts

For shell scripts (bash), follow these guidelines:

- Use shellcheck to validate scripts
- Include a shebang line (`#!/bin/bash`)
- Use double quotes around variables to prevent word splitting
- Use meaningful variable names
- Add comments for complex operations
- Check for error conditions and handle them appropriately

### Python

For Python code (used in testing and utilities), follow these guidelines:

- Follow PEP 8 style guide
- Use Python 3 features
- Document functions and classes with docstrings
- Use type hints where appropriate
- Handle exceptions properly

## Documentation Standards

- Use Markdown for documentation
- Keep documentation up-to-date with code changes
- Include examples where appropriate
- Use proper headings and formatting for readability
- Link related documents together

## Version Control Standards

- Write clear, descriptive commit messages
- Keep commits focused on a single change
- Use feature branches for new development
- Run tests before committing
- Review code before merging

## Code Review Standards

- Review all code changes before merging
- Check for adherence to coding standards
- Verify that tests are included and pass
- Look for potential security issues
- Ensure documentation is updated

## Security Standards

- Never store sensitive information (tokens, passwords) in plaintext
- Validate all user input
- Use secure defaults
- Keep dependencies up to date
- Follow the principle of least privilege
- Be cautious with file permissions
- Sanitize file paths to prevent path traversal attacks

## Performance Standards

- Profile before optimizing
- Avoid premature optimization
- Consider memory usage and allocations
- Use appropriate data structures for the task
- Be mindful of API rate limits and network latency

## Accessibility Standards

- Ensure UI elements are accessible
- Provide keyboard shortcuts for common actions
- Use high-contrast colors
- Support screen readers
- Test with accessibility tools

## Conclusion

Following these coding standards will help maintain a consistent, high-quality codebase for the onedriver project. These standards should be applied to all new code and, when feasible, to existing code during refactoring.

For questions or suggestions regarding these standards, please open an issue or pull request in the project repository.