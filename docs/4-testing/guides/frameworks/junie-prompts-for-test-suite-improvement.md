# Junie Prompts for Test Suite Improvement

This document contains Junie prompts for implementing the remaining tasks in the [Test Suite Improvement Plan](../../test-suit-improvement-plan.md).

## Phase 2: Improve Test Documentation

### Item 1: Add Comments to Clarify Test Purposes

```
I need to add comments to clarify the purposes of tests in the OneMount project. The tests follow a structured naming convention (TestXX_YY_ZZ_NN_Description_ExpectedBehavior), but many lack detailed comments explaining what they're testing and why.

Please help me create a script that:

1. Scans the codebase for test functions
2. For each test function, checks if it has a comment block explaining its purpose
3. For tests without comments, generates a template comment based on the test name
4. Outputs a report of tests that need comments

The comment template should include:
- A brief description of what the test is testing (derived from the test name)
- Placeholders for test inputs/preconditions
- Placeholders for expected outcomes
- A placeholder for explaining why this test is important

For example, for a test named "TestUT_FS_01_01_FileOperations_BasicReadWrite_SuccessfullyPreservesContent", the generated comment template might be:

```go
/*
TestUT_FS_01_01_FileOperations_BasicReadWrite_SuccessfullyPreservesContent tests that basic file read/write operations
successfully preserve content.

Inputs/Preconditions:
- [TODO: Describe the test inputs and preconditions]

Expected Outcomes:
- [TODO: Describe the expected outcomes in detail]

Importance:
- [TODO: Explain why this test is important and what it verifies about the system]
*/
```

After generating the report, I'll manually review and complete the comment templates for each test.
```

### Item 4: Document Test ID Ranges

```
I need to document test ID ranges for the OneMount project to prevent overlaps and make the test suite more organized. The test ID structure follows the pattern <TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>, and we need to assign specific ID ranges to different modules.

Please help me:

1. Analyze the existing test IDs in the registry (data/test_id_registry.json)
2. Identify patterns in how test IDs are currently assigned
3. Create a document that:
   - Assigns specific feature number ranges to different functional areas
   - Documents the purpose of each range
   - Provides guidelines for assigning new test IDs

The document should include:

1. An overview of the test ID structure
2. A table of assigned feature number ranges for each module
3. Guidelines for assigning new test IDs
4. Examples of proper test ID assignment

For example:
- FS (File System) module:
  - 01-10: Core file operations (read, write, create, delete)
  - 11-20: Directory operations (list, create, delete)
  - 21-30: Metadata operations (permissions, attributes)
  - etc.

This documentation will help prevent future test ID overlaps and make the test suite more organized.
```

## Phase 3: Consolidate Similar Tests

```
I need to consolidate similar tests in the OneMount project to reduce code duplication and improve maintainability. The test_suite_tool.py has identified 108 groups of tests with significant functional overlap.

Please help me:

1. Create a strategy for consolidating similar tests into table-driven tests
2. Develop a template for table-driven tests that preserves all test cases
3. Create a step-by-step guide for refactoring overlapping tests

The strategy should include:

1. Criteria for identifying tests that can be consolidated
2. A template for table-driven tests that includes:
   - A clear description of what's being tested
   - A table of test cases with inputs and expected outputs
   - A single test function that iterates through the test cases

For example, instead of having multiple similar tests like:

```go
func TestUT_FS_01_01_FileOperations_SmallFile_SuccessfullyWrites(t *testing.T) {
    // Test code for small file
}

func TestUT_FS_01_02_FileOperations_MediumFile_SuccessfullyWrites(t *testing.T) {
    // Similar test code for medium file
}

func TestUT_FS_01_03_FileOperations_LargeFile_SuccessfullyWrites(t *testing.T) {
    // Similar test code for large file
}
```

We would consolidate them into a single table-driven test:

```go
func TestUT_FS_01_01_FileOperations_DifferentSizes_SuccessfullyWrites(t *testing.T) {
    testCases := []struct {
        name     string
        fileSize int
        expected string
    }{
        {"SmallFile", 100, "expected result for small file"},
        {"MediumFile", 1000, "expected result for medium file"},
        {"LargeFile", 10000, "expected result for large file"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test code that works for all file sizes
        })
    }
}
```

The guide should also include best practices for:
- Preserving test IDs during consolidation
- Ensuring all test cases are still covered
- Maintaining clear test boundaries
```

## Phase 4: Clarify Test Boundaries

```
I need to clarify test boundaries in the OneMount project to make the test suite more maintainable and easier to understand. The test_suite_tool.py has identified 22 pairs of tests with ambiguous or unclear boundaries.

Please help me:

1. Create guidelines for clarifying test boundaries
2. Develop a template for documenting test boundaries
3. Create a step-by-step guide for refactoring tests with ambiguous boundaries

The guidelines should include:

1. Criteria for identifying tests with ambiguous boundaries
2. Principles for defining clear test boundaries
3. Examples of good and bad test boundaries

The template for documenting test boundaries should include:

1. A clear description of what the test is testing
2. The scope and limitations of the test
3. Relationships with other tests
4. Explicit boundary conditions

For example, for a pair of tests with ambiguous boundaries:

```go
// Before: Ambiguous boundaries
func TestUT_FS_01_01_FileOperations_ReadFile_SuccessfullyReads(t *testing.T) {
    // Test code for reading a file
}

func TestUT_FS_01_02_FileOperations_ReadFileContents_ContentsMatch(t *testing.T) {
    // Similar test code for reading a file and checking contents
}

// After: Clear boundaries
/*
TestUT_FS_01_01_FileOperations_ReadFile_SuccessfullyReads tests that a file can be opened and read.
This test focuses only on the ability to open and read a file, not on the correctness of the contents.
It verifies that the read operation completes without errors and returns some data.

Related tests:
- TestUT_FS_01_02_FileOperations_ReadFileContents_ContentsMatch: Tests that the contents read from a file match the expected contents.
*/
func TestUT_FS_01_01_FileOperations_ReadFile_SuccessfullyReads(t *testing.T) {
    // Test code for reading a file
}

/*
TestUT_FS_01_02_FileOperations_ReadFileContents_ContentsMatch tests that the contents read from a file match the expected contents.
This test assumes that the file can be opened and read (tested by TestUT_FS_01_01) and focuses specifically on the correctness of the contents.

Related tests:
- TestUT_FS_01_01_FileOperations_ReadFile_SuccessfullyReads: Tests that a file can be opened and read.
*/
func TestUT_FS_01_02_FileOperations_ReadFileContents_ContentsMatch(t *testing.T) {
    // Test code for reading a file and checking contents
}
```

The guide should also include best practices for:
- Refactoring tests to have clearer boundaries
- Documenting relationships between tests
- Ensuring all functionality is still tested after refactoring
```

## Phase 5: Establish Ongoing Test Quality Processes

```
I need to establish ongoing test quality processes for the OneMount project to maintain the quality of the test suite over time. This includes regular reviews, integration with CI/CD, guidelines for new tests, and monitoring test coverage.

Please help me:

1. Create a test quality checklist for code reviews
2. Develop a CI/CD integration plan for test analysis
3. Create guidelines for writing new tests
4. Develop a test coverage monitoring strategy

The test quality checklist should include:

1. Verification that new tests have unique IDs
2. Verification that new tests have clear boundaries
3. Verification that new tests have adequate comments
4. Verification that new tests follow the naming convention
5. Consideration of whether new tests can be consolidated with existing tests

The CI/CD integration plan should include:

1. Running the test_suite_tool.py as part of the CI/CD pipeline
2. Failing the build if duplicate test IDs are found
3. Generating reports on test quality metrics
4. Tracking test quality trends over time

The guidelines for writing new tests should include:

1. How to choose a test ID
2. How to name a test
3. How to document a test
4. How to define clear test boundaries
5. When to use table-driven tests

The test coverage monitoring strategy should include:

1. Tools for measuring test coverage
2. Targets for test coverage
3. Processes for addressing coverage gaps
4. Regular reporting on test coverage

These processes will help maintain the quality of the test suite over time and prevent the recurrence of the issues identified in the test suite analysis.
```

### Implementation Plan

To implement these improvements, I recommend the following approach:

1. Start with Phase 2 (Improve Test Documentation) to establish a solid foundation
2. Move on to Phase 4 (Clarify Test Boundaries) to make the test suite more understandable
3. Then implement Phase 3 (Consolidate Similar Tests) to reduce code duplication
4. Finally, implement Phase 5 (Establish Ongoing Test Quality Processes) to maintain quality over time

For each phase:
1. Use the Junie prompt to generate an implementation plan
2. Review and refine the plan
3. Implement the plan
4. Verify the results
5. Document the improvements

This approach will ensure that the test suite improvements are implemented in a logical order and that each phase builds on the previous one.