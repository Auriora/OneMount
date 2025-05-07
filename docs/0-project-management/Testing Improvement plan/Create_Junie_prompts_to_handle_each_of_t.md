
# Junie Prompts for Mock Graph Testing

Based on the previous issue about determining if the mock graph object is being properly used, here are Junie prompts to handle each of the codifiable recommendations:

## 1. Enable Operational Offline Mode

```
Create a test that enables operational offline mode to prevent real network requests during testing. The test should:
1. Set operational offline mode at the beginning
2. Verify that network requests fail with the expected error
3. Reset operational offline mode at the end
```

## 2. Verify Mock Client Usage

```
Implement a test that verifies the mock client is recording method calls correctly. The test should:
1. Create a mock graph client
2. Perform several operations (GetItem, GetItemChildren, etc.)
3. Retrieve the recorder and verify the expected methods were called
4. Check the number of calls for each method matches expectations
```

## 3. Add Proper Mock Responses

```
Create a test that demonstrates how to add mock responses for different API calls. The test should:
1. Create a mock graph client
2. Add mock responses for item retrieval, content download, and children listing
3. Perform operations that use these mock responses
4. Verify the operations return the expected results
```

## 4. Use Test Helper Functions

```
Refactor an existing test to use the FSTestFixture helper. The refactored test should:
1. Use helpers.SetupFSTestFixture instead of manual setup
2. Configure any additional mock responses needed for the specific test
3. Use the fixture.Use pattern to run the test
4. Verify the test runs correctly with the helper
```

## 5. Implement a Comprehensive Test

```
Create a comprehensive test that combines all best practices for mock usage. The test should:
1. Use the FSTestFixture helper for setup
2. Enable operational offline mode
3. Add necessary mock responses
4. Perform filesystem operations
5. Verify mock client calls
6. Ensure no real network requests are made
```

## 6. Test Network Error Simulation

```
Create a test that simulates network errors using the mock client. The test should:
1. Configure the mock client with error conditions (using SetConfig)
2. Set error rates, throttling, and latency
3. Perform operations and verify they handle errors correctly
4. Test both random errors and API throttling scenarios
```

## 7. Test Pagination Support

```
Create a test that verifies pagination works correctly with the mock client. The test should:
1. Create a large collection of items (>25)
2. Add them to the mock client with pagination enabled
3. Retrieve the items and verify all pages are processed correctly
4. Check that the nextLink property is handled properly
```

## 8. Test Offline Mode File Operations

```
Create a test that verifies file operations work in offline mode. The test should:
1. Set up a filesystem with cached files
2. Enable operational offline mode
3. Perform read operations on cached files
4. Attempt write operations and verify they're queued
5. Verify error handling for uncached files
```

Each of these prompts addresses a specific aspect of properly using the mock graph object in tests, helping to prevent real network requests and Microsoft login prompts from appearing during testing.