# Test Quality Improvement Junie Prompts

This document contains Junie prompts for improving the quality of the OneMount test suite.

## 1. Test Overlap and Conflict Analysis

```
I need to analyze the OneMount test suite for overlaps and conflicts, particularly focusing on test IDs. Please help me:

1. Identify any duplicate or conflicting test IDs across the test suite
2. Find test cases that cover the same functionality but with different approaches
3. Detect redundant test cases that don't add additional coverage
4. Identify tests with ambiguous or unclear boundaries between them

For this analysis:
1. First, scan all test files to extract test IDs, names, and descriptions
2. Create a mapping of test IDs to their locations and purposes
3. Flag any duplicate test IDs or naming conflicts
4. Analyze test descriptions and implementations to identify functional overlaps
5. Suggest consolidation opportunities where tests can be combined
6. Recommend clear boundaries between related tests

The output should include:
- A list of any conflicting test IDs with their locations
- Groups of tests with significant functional overlap
- Recommendations for resolving conflicts and reducing redundancy
- Suggestions for better organizing related tests

Please focus on the test files in the following directories:
- cmd/common
- internal/fs
- internal/fs/graph
- internal/fs/offline
- internal/ui
```

## 2. Test Case Matrix Update

```
I need to update the test case traceability matrix for the OneMount project to ensure it accurately reflects all current test cases and their relationships to requirements. Please help me:

1. Compare the existing test-cases-traceability-matrix.md with the actual test implementations
2. Update the matrix to include any new test cases that have been implemented
3. Verify that the requirements coverage information is accurate
4. Ensure all test IDs in the matrix match the actual test implementations

For this task:
1. First, scan all test files to extract the current test IDs and descriptions
2. Compare this list with the test cases in the existing traceability matrix
3. Identify any tests that are in the code but missing from the matrix
4. Identify any tests in the matrix that don't exist in the code
5. For each test, verify that the requirements coverage information is accurate
6. Update the matrix with any new or changed information

The output should be an updated version of the test-cases-traceability-matrix.md file that:
- Includes all implemented test cases
- Has accurate requirements coverage information
- Maintains the same format and structure as the original
- Includes any new test cases with their appropriate requirements mappings

Please use the test-case-stubs-checklist.md file as a reference for all implemented test cases, and the requirements documents in docs/requirements/srs/ for understanding the requirements.
```

## 3. Gap Analysis Between Tests and Requirements

```
I need to perform a comprehensive gap analysis between the OneMount test cases and the project's requirements, architecture, and design documentation. Please help me identify:

1. Requirements that are not adequately covered by tests
2. Architectural elements that lack sufficient test coverage
3. Design components that need additional testing
4. Areas where test coverage could be improved

For this analysis:
1. First, review the requirements in docs/requirements/srs/3-specific-requirements.md
2. Review the architecture documentation in docs/design/software-architecture-specification.md
3. Review the design documentation and traceability matrices
4. Compare these documents with the existing test cases in test-cases-traceability-matrix.md
5. Identify requirements without corresponding test cases
6. Identify architectural elements without adequate test coverage
7. Identify design components that lack sufficient testing

The output should include:
- A list of requirements not covered by existing tests, organized by priority
- Architectural elements that need additional test coverage
- Design components that require more testing
- Recommendations for new test cases to fill the identified gaps
- Suggestions for improving existing tests to better cover requirements

For each gap identified, please provide:
- The specific requirement, architectural element, or design component
- The current test coverage (if any)
- The recommended approach to address the gap
- A suggested priority level for addressing the gap

This analysis will help ensure that our test suite comprehensively validates that the system meets all its requirements and conforms to its architecture and design.
```

## 4. Related Testing Tasks

```
I need suggestions for additional testing-related tasks that would improve the quality and effectiveness of the OneMount test suite. Please provide recommendations for:

1. Improving test infrastructure and automation
2. Enhancing test coverage metrics and reporting
3. Implementing additional types of testing
4. Improving test documentation and maintenance

For each suggested task, please provide:
- A clear description of the task
- The benefits of implementing it
- The estimated effort required (low, medium, high)
- Any dependencies or prerequisites
- Implementation steps or approach

Consider the following areas for improvement:
- Test automation and CI/CD integration
- Performance testing enhancements
- Security testing improvements
- Usability and accessibility testing
- Test data management
- Test environment management
- Test result analysis and reporting
- Test maintenance and refactoring
- Test documentation improvements
- Developer testing practices

The output should be a prioritized list of tasks that would provide the most value for improving the overall quality and effectiveness of the OneMount test suite. For each task, include enough detail that it could be assigned to a developer for implementation.
```