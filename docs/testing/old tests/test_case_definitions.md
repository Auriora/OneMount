# OneMount Test Case Definitions

This document contains detailed test case definitions for each test in the OneMount project. The test cases are organized by test type (unit test, integration test, etc.) and include a mapping of existing test names to a new name structure with test IDs.

## Test Files and Their Tests

The following table lists all _test.go files in the project and the tests defined in each file:

| Test File | Tests |
|-----------|-------|
| cmd/common/common_test.go | TestXDGVolumeInfo |
| cmd/common/config_test.go | TestLoadConfig, TestConfigMerge, TestLoadNonexistentConfig, TestWriteConfig |
| cmd/common/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/fs/cache_test.go | TestCacheOperations |
| internal/fs/dbus_test.go | TestDBusServerOperations, TestDBusServerFunctionality |
| internal/fs/delta_test.go | TestDeltaOperations, TestDeltaContentChangeRemote, TestDeltaContentChangeBoth, TestDeltaBadContentInCache, TestDeltaFolderDeletion, TestDeltaFolderDeletionNonEmpty, TestDeltaNoModTimeUpdate, TestDeltaMissingHash |
| internal/fs/fs_test.go | TestReaddir, TestLs, TestTouchOperations, TestFilePermissions, TestDirectoryOperations, TestDirectoryRemoval, TestFileOperations, TestWriteOffset, TestFileMovementOperations, TestPositionalFileOperations, TestBasicFileSystemOperations, TestCaseSensitivityHandling, TestFilenameCase, TestShellFileOperations, TestFileInfo, TestNoQuestionMarks, TestGIOTrash, TestListChildrenPaging, TestLibreOfficeSavePattern, TestDisallowedFilenames |
| internal/fs/graph/drive_item_methods_test.go | TestDriveItemIsDir, TestDriveItemModTimeUnix, TestDriveItemVerifyChecksum, TestDriveItemETagIsMatch |
| internal/fs/graph/drive_item_test.go | TestGetItem |
| internal/fs/graph/graph_test.go | TestResourcePath, TestRequestUnauthenticated |
| internal/fs/graph/hashes_test.go | TestSha1HashReader, TestQuickXORHashReader, TestHashSeekPosition |
| internal/fs/graph/hash_functions_test.go | TestSHA256Hash, TestSHA256HashStream, TestSHA1Hash, TestSHA1HashStream, TestQuickXORHash, TestQuickXORHashStream |
| internal/fs/graph/mock_graph_test.go | (Tests not examined) |
| internal/fs/graph/oauth2_gtk_test.go | TestURIGetHost |
| internal/fs/graph/oauth2_test.go | TestAuthCodeFormat, TestAuthFromfile, TestAuthRefresh, TestAuthConfigMerge, TestAuthFailureWithNetworkAvailable |
| internal/fs/graph/offline_test.go | TestOperationalOfflineState, TestIsOfflineWithOperationalState, TestIsOffline |
| internal/fs/graph/path_test.go | TestIDPath, TestChildrenPath, TestChildrenPathID |
| internal/fs/graph/quickxorhash/quickxorhash_test.go | TestQuickXorHash, TestQuickXorHashByBlock, TestSize, TestBlockSize, TestReset |
| internal/fs/graph/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/fs/inode_test.go | TestInodeCreation, TestInodeProperties, TestFilenameEscaping, TestFileCreationBehavior |
| internal/fs/logging_test.go | TestLogging, TestGetCurrentGoroutineID, TestLoggingInMethods |
| internal/fs/offline/offline_test.go | TestOfflineFileAccess, TestOfflineFileSystemOperations, TestOfflineChangesCached, TestOfflineSynchronization |
| internal/fs/offline/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/fs/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/fs/sync_test.go | TestSyncDirectoryTree |
| internal/fs/thumbnail_test.go | TestThumbnailCacheOperations, TestThumbnailCacheCleanup, TestThumbnailOperations |
| internal/fs/upload_manager_test.go | TestUploadDiskSerialization, TestRepeatedUploads |
| internal/fs/upload_session_test.go | TestUploadSessionOperations |
| internal/fs/xattr_operations_test.go | TestXattrOperations, TestFileStatusXattr, TestFilesystemXattrOperations |
| internal/ui/onemount_test.go | TestMountpointIsValid, TestHomeEscapeUnescape, TestGetAccountName |
| internal/ui/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/ui/systemd/setup_test.go | (Contains only setup routines, see [Test Setup Documentation](test_setup_documentation.md)) |
| internal/ui/systemd/systemd_test.go | TestTemplateUnit, TestUntemplateUnit, TestUnitEnabled, TestUnitActive |

Note: For files marked with "(Tests not examined)", the tests have not been individually identified and documented yet.

## Test Types

The OneMount project contains the following types of tests:

1. **Unit Tests**: These test individual functions or small components in isolation.
2. **Integration Tests**: These test the interaction between multiple components.
3. **System Tests**: These test the system as a whole, including the filesystem, API integration, and UI.

## Test ID Structure

The test ID structure follows this pattern:

```
<TYPE>-<COMPONENT>-<TESTNUMBER>-<SUBTESTNUMER>
```

Where:
- `<TYPE>` is the test type (2 letters):
  - UT - Unit Test
  - IT - Integration Test
  - ST - System Test
  - PT - Performance Test
  - LT - Load Test
  - SC - Scenario Test
  - UA - User Acceptance Test
  - etc.
- `<COMPONENT>` is the component being tested (2/3 letters):
  - FS - File System
  - GR - Graph
  - UI - User Interface
  - CMD - Command
  - etc.
- `<TESTNUMBER>` is a 2-digit number uniquely identifying the test
- `<SUBTESTNUMER>` is a 2-digit number uniquely identifying the sub-test or test variant

## Test Function Naming Convention

Test function names follow this pattern:

```
Test<TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>_<UNIT-OF-WORK>_<STATE-UNDER-TEST>_<EXPECTED-BEHAVIOR>
```

Where:
- `<TYPE>`, `<COMPONENT>`, `<TESTNUMBER>`, and `<SUBTESTNUMER>` are the same as in the test ID structure
- `<UNIT-OF-WORK>` represents a single method, a class, or multiple classes
- `<STATE-UNDER-TEST>` represents the inputs or conditions being tested
- `<EXPECTED-BEHAVIOR>` represents the output or result

## Test Case Mapping

### Unit Tests

#### cmd/common Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| UT-CMD-01-01 | TestXDGVolumeInfo | TestUT_CMD_01_01_XDGVolumeInfo_ValidInput_MatchesExpected | Tests reading and writing .xdg-volume-info files | 1. Create a temporary file<br>2. Write XDG volume info with a specific name<br>3. Read the name from the file | The read name matches the written name |
| UT-CMD-02-01 | TestLoadConfig | TestUT_CMD_02_01_Config_ValidConfigFile_LoadsCorrectValues | Tests loading configuration from a file | 1. Load configuration from a test config file<br>2. Get the user's home directory<br>3. Check if the loaded configuration matches expected values | The configuration values match the expected values from the config file |
| UT-CMD-03-01 | TestConfigMerge | TestUT_CMD_03_01_Config_MergedSettings_ContainsMergedValues | Tests merging configuration settings | 1. Load configuration from a test config file with merged settings<br>2. Check if the loaded configuration contains the merged values | The configuration contains the merged values |
| UT-CMD-04-01 | TestLoadNonexistentConfig | TestUT_CMD_04_01_Config_NonexistentFile_LoadsDefaultValues | Tests loading default configuration when the config file doesn't exist | 1. Load configuration from a nonexistent config file<br>2. Get the user's home directory<br>3. Check if the loaded configuration contains default values | The configuration contains the default values |
| UT-CMD-05-01 | TestWriteConfig | TestUT_CMD_05_01_Config_ValidSettings_WritesSuccessfully | Tests writing a configuration file | 1. Load configuration from a test config file<br>2. Write the configuration to a new file<br>3. Check if the write operation succeeds | The configuration is successfully written to the file |

#### internal/fs/graph Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| UT-GR-01-01 | TestResourcePath | TestUT_GR_01_01_ResourcePath_VariousInputs_ReturnsEscapedPath | Tests the ResourcePath function with various inputs | 1. Call ResourcePath with different path inputs<br>2. Compare the result with the expected output | The escaped path matches the expected format for Microsoft Graph API |
| UT-GR-02-01 | TestRequestUnauthenticated | TestUT_GR_02_01_Request_UnauthenticatedUser_ReturnsError | Tests the behavior of an unauthenticated request | 1. Create an Auth object with expired token<br>2. Attempt to make a GET request<br>3. Check if an error is returned | An error is returned for the unauthenticated request |
| UT-GR-03-01 | TestDriveItemIsDir | TestUT_GR_03_01_DriveItem_DifferentTypes_IsDirectoryReturnsCorrectValue | Tests the IsDir method of DriveItem | 1. Create DriveItem objects with different types (folder, file, empty)<br>2. Call IsDir on each object<br>3. Check if the result matches expectations | IsDir returns true for folders and false for files and empty items |
| UT-GR-04-01 | TestDriveItemModTimeUnix | TestUT_GR_04_01_DriveItem_ModificationTime_ReturnsCorrectUnixTimestamp | Tests the ModTimeUnix method of DriveItem | 1. Create a DriveItem with a specific modification time<br>2. Call ModTimeUnix on the item<br>3. Check if the result matches the expected Unix timestamp | ModTimeUnix returns the correct Unix timestamp |
| UT-GR-05-01 | TestDriveItemVerifyChecksum | TestUT_GR_05_01_DriveItem_VariousChecksums_VerificationReturnsCorrectResult | Tests the VerifyChecksum method of DriveItem | 1. Create DriveItem objects with different checksums<br>2. Call VerifyChecksum with matching and non-matching checksums<br>3. Check if the result matches expectations | VerifyChecksum returns true for matching checksums and false for non-matching checksums |
| UT-GR-06-01 | TestDriveItemETagIsMatch | TestUT_GR_06_01_DriveItem_VariousETags_MatchReturnsCorrectResult | Tests the ETagIsMatch method of DriveItem | 1. Create DriveItem objects with different ETags<br>2. Call ETagIsMatch with matching and non-matching ETags<br>3. Check if the result matches expectations | ETagIsMatch returns true for matching ETags and false for non-matching ETags |
| UT-GR-07-01 | TestGetItem | TestUT_GR_07_01_GraphAPI_VariousPaths_ReturnsCorrectItems | Tests retrieving items from the Microsoft Graph API | 1. Load authentication tokens<br>2. Call GetItemPath with different paths<br>3. Check if the result matches expectations | GetItemPath returns the correct item for valid paths and an error for invalid paths |
| UT-GR-08-01 | TestSha1HashReader | TestUT_GR_08_01_SHA1Hash_ReaderInput_MatchesDirectCalculation | Tests the SHA1HashStream function with a reader | 1. Create a byte array with test content<br>2. Calculate the SHA1 hash of the content<br>3. Create a reader from the content<br>4. Calculate the SHA1 hash using SHA1HashStream<br>5. Compare the two hashes | The hash calculated from the reader matches the hash calculated from the byte array |
| UT-GR-09-01 | TestQuickXORHashReader | TestUT_GR_09_01_QuickXORHash_ReaderInput_MatchesDirectCalculation | Tests the QuickXORHashStream function with a reader | 1. Create a byte array with test content<br>2. Calculate the QuickXOR hash of the content<br>3. Create a reader from the content<br>4. Calculate the QuickXOR hash using QuickXORHashStream<br>5. Compare the two hashes | The hash calculated from the reader matches the hash calculated from the byte array |
| UT-GR-10-01 | TestHashSeekPosition | TestUT_GR_10_01_HashFunctions_AfterReading_ResetSeekPosition | Tests that hash functions reset the seek position | 1. Create a temporary file with test content<br>2. Read a portion of the file to move the seek position<br>3. Calculate hashes using the hash stream functions<br>4. Verify that the seek position is reset to the beginning after each hash calculation | The seek position is reset to the beginning after each hash calculation |
| UT-GR-11-01 | TestSHA256Hash | TestUT_GR_11_01_SHA256Hash_VariousInputs_ReturnsCorrectHash | Tests the SHA256Hash function with different inputs | 1. Create byte arrays with different test content<br>2. Calculate the SHA256 hash of each content<br>3. Compare the results with expected values | SHA256Hash returns the correct hash for each input |
| UT-GR-12-01 | TestSHA256HashStream | TestUT_GR_12_01_SHA256HashStream_VariousInputs_ReturnsCorrectHash | Tests the SHA256HashStream function with different inputs | 1. Create readers with different test content<br>2. Calculate the SHA256 hash of each content using SHA256HashStream<br>3. Compare the results with expected values | SHA256HashStream returns the correct hash for each input |
| UT-GR-13-01 | TestSHA1Hash | TestUT_GR_13_01_SHA1Hash_VariousInputs_ReturnsCorrectHash | Tests the SHA1Hash function with different inputs | 1. Create byte arrays with different test content<br>2. Calculate the SHA1 hash of each content<br>3. Compare the results with expected values | SHA1Hash returns the correct hash for each input |
| UT-GR-14-01 | TestSHA1HashStream | TestUT_GR_14_01_SHA1HashStream_VariousInputs_ReturnsCorrectHash | Tests the SHA1HashStream function with different inputs | 1. Create readers with different test content<br>2. Calculate the SHA1 hash of each content using SHA1HashStream<br>3. Compare the results with expected values | SHA1HashStream returns the correct hash for each input |
| UT-GR-15-01 | TestQuickXORHash | TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash | Tests the QuickXORHash function with different inputs | 1. Create byte arrays with different test content<br>2. Calculate the QuickXOR hash of each content<br>3. Compare the results with expected values | QuickXORHash returns the correct hash for each input |
| UT-GR-16-01 | TestQuickXORHashStream | TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash | Tests the QuickXORHashStream function with different inputs | 1. Create readers with different test content<br>2. Calculate the QuickXOR hash of each content using QuickXORHashStream<br>3. Compare the results with expected values | QuickXORHashStream returns the correct hash for each input |
| UT-GR-17-01 | TestURIGetHost | TestUT_GR_17_01_URIGetHost_VariousURIs_ReturnsCorrectHost | Tests the uriGetHost function with various inputs | 1. Call uriGetHost with an invalid URI<br>2. Call uriGetHost with a valid HTTPS URI with a path<br>3. Call uriGetHost with a valid HTTP URI without a path<br>4. Check if the results match expectations | uriGetHost returns the correct host for valid URIs and an empty string for invalid URIs |
| UT-GR-18-01 | TestAuthCodeFormat | TestUT_GR_18_01_ParseAuthCode_VariousFormats_ExtractsCorrectCode | Tests the parseAuthCode function with various inputs | 1. Call parseAuthCode with different input formats<br>2. Check if the results match expectations | parseAuthCode correctly extracts the authorization code from different URL formats |
| UT-GR-19-01 | TestAuthFromfile | TestUT_GR_19_01_Auth_LoadFromFile_TokensLoadedSuccessfully | Tests loading authentication tokens from a file | 1. Verify that the auth tokens file exists<br>2. Load authentication tokens from the file<br>3. Check if the access token is not empty | Authentication tokens are successfully loaded from the file |
| UT-GR-20-01 | TestAuthRefresh | TestUT_GR_20_01_Auth_TokenRefresh_TokensRefreshedSuccessfully | Tests refreshing authentication tokens | 1. Load authentication tokens from a file<br>2. Force an auth refresh by setting ExpiresAt to 0<br>3. Refresh the authentication tokens<br>4. Check if the new expiration time is in the future | Authentication tokens are successfully refreshed |
| UT-GR-21-01 | TestAuthConfigMerge | TestUT_GR_21_01_AuthConfig_MergeWithDefaults_PreservesCustomValues | Tests merging authentication configuration with default values | 1. Create a test AuthConfig with a custom RedirectURL<br>2. Apply defaults to the AuthConfig<br>3. Check if the RedirectURL is preserved and default values are applied | Default values are correctly applied while preserving custom values |
| UT-GR-22-01 | TestAuthFailureWithNetworkAvailable | TestUT_GR_22_01_Auth_FailureWithNetwork_ReturnsErrorAndInvalidState | Tests the behavior when authentication fails but network is available | 1. Create an Auth with invalid credentials but valid configuration<br>2. Apply defaults to the AuthConfig<br>3. Attempt to refresh the tokens<br>4. Check if an error is returned and the auth state is still invalid | An error is returned and the auth state remains invalid |
| UT-GR-23-01 | TestOperationalOfflineState | TestUT_GR_23_01_OfflineState_SetAndGet_StateCorrectlyManaged | Tests setting and getting the operational offline state | 1. Reset the operational offline state<br>2. Check the default state<br>3. Set the state to true and check it<br>4. Set the state back to false and check it | The operational offline state is correctly set and retrieved |
| UT-GR-24-01 | TestIsOfflineWithOperationalState | TestUT_GR_24_01_IsOffline_OperationalStateSet_ReturnsTrue | Tests the IsOffline function when operational offline state is set | 1. Set the operational offline state to true<br>2. Call IsOffline with different errors<br>3. Reset the operational offline state<br>4. Call IsOffline with different errors again | IsOffline returns true when operational offline is set, regardless of the error |
| UT-GR-25-01 | TestIsOffline | TestUT_GR_25_01_IsOffline_VariousErrors_IdentifiesNetworkErrors | Tests the IsOffline function with various error types | 1. Reset the operational offline state<br>2. Call IsOffline with different types of errors<br>3. Check if the results match expectations | IsOffline correctly identifies network-related errors |
| UT-GR-26-01 | TestIDPath | TestUT_GR_26_01_IDPath_VariousItemIDs_FormatsCorrectly | Tests the IDPath function with various inputs | 1. Call IDPath with different item IDs<br>2. Check if the results match expectations | IDPath correctly formats item IDs for API requests |
| UT-GR-27-01 | TestChildrenPath | TestUT_GR_27_01_ChildrenPath_VariousPaths_FormatsCorrectly | Tests the childrenPath function with various inputs | 1. Call childrenPath with different paths<br>2. Check if the results match expectations | childrenPath correctly formats paths for retrieving children |
| UT-GR-28-01 | TestChildrenPathID | TestUT_GR_28_01_ChildrenPathID_VariousItemIDs_FormatsCorrectly | Tests the childrenPathID function with various inputs | 1. Call childrenPathID with different item IDs<br>2. Check if the results match expectations | childrenPathID correctly formats item IDs for retrieving children |
| UT-GR-29-01 | TestQuickXorHash | TestUT_GR_29_01_QuickXORHash_Sum_CalculatesCorrectHashes | Tests the Sum function with various inputs | 1. Decode test vectors from base64<br>2. Calculate QuickXOR hashes using Sum<br>3. Compare the results with expected values | Sum correctly calculates QuickXOR hashes |
| UT-GR-30-01 | TestQuickXorHashByBlock | TestUT_GR_30_01_QuickXORHash_WriteByBlocks_CalculatesCorrectHashes | Tests calculating QuickXOR hashes by writing data in blocks | 1. Decode test vectors from base64<br>2. Create a new QuickXOR hash<br>3. Write data in blocks of different sizes<br>4. Calculate the hash<br>5. Compare the result with the expected value | QuickXOR hashes are correctly calculated when writing data in blocks |
| UT-GR-31-01 | TestSize | TestUT_GR_31_01_QuickXORHash_Size_Returns20Bytes | Tests the Size method of the QuickXOR hash | 1. Create a new QuickXOR hash<br>2. Call the Size method<br>3. Check if the result is 20 | Size returns the correct hash size (20 bytes) |
| UT-GR-32-01 | TestBlockSize | TestUT_GR_32_01_QuickXORHash_BlockSize_Returns64Bytes | Tests the BlockSize method of the QuickXOR hash | 1. Create a new QuickXOR hash<br>2. Call the BlockSize method<br>3. Check if the result is 64 | BlockSize returns the correct block size (64 bytes) |
| UT-GR-33-01 | TestReset | TestUT_GR_33_01_QuickXORHash_Reset_RestoresInitialState | Tests the Reset method of the QuickXOR hash | 1. Create a new QuickXOR hash<br>2. Calculate the hash of an empty input<br>3. Write some data and calculate the hash<br>4. Reset the hash<br>5. Calculate the hash again<br>6. Compare with the original empty hash | Reset correctly resets the hash state |

#### internal/fs Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| UT-FS-01-01 | TestInodeCreation | TestUT_FS_01_01_Inode_Creation_HasCorrectProperties | Tests that inodes are created with the correct properties | 1. Create inodes with different modes (file, directory, executable)<br>2. Verify the properties of each inode | Inodes have the correct properties (ID, name, mode, directory status) |
| UT-FS-02-01 | TestInodeProperties | TestUT_FS_02_01_Inode_Properties_ModeAndDirectoryDetection | Tests various properties of inodes, including mode and directory detection | 1. Create test directories and files<br>2. Get the items from the server<br>3. Create inodes from the drive items<br>4. Test the mode and IsDir methods | Inodes have the correct mode and directory status |
| UT-FS-03-01 | TestFilenameEscaping | TestUT_FS_03_01_Filename_SpecialCharacters_ProperlyEscaped | Tests that filenames with special characters are properly escaped | 1. Create files with special characters in their names<br>2. Verify the files are created successfully<br>3. Verify the files are uploaded to the server<br>4. Verify the file content matches what was written | Files with special characters in their names are properly handled |
| UT-FS-04-01 | TestFileCreationBehavior | TestUT_FS_04_01_FileCreation_VariousScenarios_BehavesCorrectly | Tests various behaviors when creating files | 1. Create a file<br>2. Create the same file again<br>3. Verify the same inode is returned<br>4. Test with different modes and after writing content | File creation behavior is correct, including truncation and returning the same inode |
| UT-FS-05-01 | TestLogging | TestUT_FS_05_01_Logging_MethodCallsAndReturns_LogsCorrectly | Tests the LogMethodCall and LogMethodReturn functions | 1. Call LogMethodCall<br>2. Call LogMethodReturn with different types of return values<br>3. Verify the log output contains the expected information | Logging functions correctly log method calls and returns |
| UT-FS-06-01 | TestGetCurrentGoroutineID | TestUT_FS_06_01_GoroutineID_GetCurrent_ReturnsValidID | Tests the getCurrentGoroutineID function | 1. Call getCurrentGoroutineID<br>2. Verify the result is not empty and is a number | getCurrentGoroutineID returns a valid goroutine ID |
| UT-FS-07-01 | TestLoggingInMethods | TestUT_FS_07_01_Logging_InMethods_LogsCorrectly | Tests the logging in actual methods | 1. Create a test filesystem<br>2. Call methods with logging<br>3. Verify the log output contains the expected information | Methods correctly log their calls and returns |
| UT-FS-08-01 | TestThumbnailCacheOperations | TestUT_FS_08_01_ThumbnailCache_BasicOperations_WorkCorrectly | Tests various operations on the thumbnail cache | 1. Create a thumbnail cache<br>2. Insert thumbnails<br>3. Check if thumbnails exist<br>4. Retrieve thumbnails<br>5. Delete thumbnails | Thumbnail cache operations work correctly |
| UT-FS-09-01 | TestThumbnailCacheCleanup | TestUT_FS_09_01_ThumbnailCache_Cleanup_RemovesExpiredThumbnails | Tests the cleanup functionality of the thumbnail cache | 1. Create a thumbnail cache<br>2. Insert thumbnails<br>3. Set expiration times<br>4. Run cleanup<br>5. Verify expired thumbnails are removed | Thumbnail cache cleanup correctly removes expired thumbnails |
| UT-FS-10-01 | TestThumbnailOperations | TestUT_FS_10_01_Thumbnails_FileSystemOperations_WorkCorrectly | Tests various operations on thumbnails in the filesystem | 1. Create a filesystem<br>2. Find an image file<br>3. Get thumbnails of different sizes<br>4. Delete thumbnails<br>5. Verify thumbnails are cached and deleted correctly | Thumbnail operations in the filesystem work correctly |

#### internal/ui Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| UT-UI-01-01 | TestMountpointIsValid | TestUT_UI_01_01_Mountpoint_Validation_ReturnsCorrectResult | Tests that mountpoints are validated correctly | 1. Create test directories and files<br>2. Call MountpointIsValid with different paths<br>3. Check if the result matches expectations | Valid mountpoints return true, invalid mountpoints return false |
| UT-UI-02-01 | TestHomeEscapeUnescape | TestUT_UI_02_01_HomePath_EscapeAndUnescape_ConvertsPaths | Tests converting paths from ~/some_path to /home/username/some_path and back | 1. Call EscapeHome with different paths<br>2. Call UnescapeHome with the escaped paths<br>3. Check if the results match expectations | Paths are correctly escaped and unescaped |
| UT-UI-03-01 | TestGetAccountName | TestUT_UI_03_01_AccountName_FromAuthTokenFiles_ReturnsCorrectNames | Tests retrieving account names from auth token files | 1. Create various auth token files (valid, invalid, empty)<br>2. Call GetAccountName with different instances<br>3. Check if the results match expectations | Account names are correctly retrieved from valid files, errors are returned for invalid files |

#### internal/ui/systemd Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| UT-UI-04-01 | TestTemplateUnit | TestUT_UI_04_01_SystemdUnit_Template_AppliesInstanceName | Tests the TemplateUnit function | 1. Define test cases with different unit names and instance names<br>2. Call TemplateUnit with each test case<br>3. Check if the result matches the expected templated unit name | Unit names are correctly templated with instance names |
| UT-UI-05-01 | TestUntemplateUnit | TestUT_UI_05_01_SystemdUnit_Untemplate_ExtractsUnitAndInstanceName | Tests the UntemplateUnit function | 1. Define test cases with different templated unit names<br>2. Call UntemplateUnit with each test case<br>3. Check if the result matches the expected unit name and instance name | Templated unit names are correctly untemplated into unit name and instance name |

### Integration Tests

#### internal/fs Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| IT-FS-01-01 | TestCacheOperations | TestIT_FS_01_01_Cache_BasicOperations_WorkCorrectly | Tests various cache operations | 1. Create a filesystem cache<br>2. Perform operations on the cache (get path, get children, check pointers)<br>3. Verify the results of each operation | Cache operations work correctly |
| IT-FS-02-01 | TestDBusServerFunctionality | TestIT_FS_02_01_DBusServer_BasicFunctionality_WorksCorrectly | Tests D-Bus server functionality | 1. Set up a D-Bus server<br>2. Perform operations (get file status, emit signals, reconnect)<br>3. Verify the results of each operation | D-Bus server functionality works correctly |
| IT-FS-03-01 | TestDeltaOperations | TestIT_FS_03_01_Delta_SyncOperations_ChangesAreSynced | Tests delta operations for syncing changes | 1. Set up test files/directories<br>2. Perform operations on the server (create, delete, rename, move)<br>3. Verify that changes are synced to the client | Delta operations correctly sync changes |
| IT-FS-04-01 | TestDeltaContentChangeRemote | TestIT_FS_04_01_Delta_RemoteContentChange_ClientIsUpdated | Tests syncing content changes from server to client | 1. Create a file on the client<br>2. Change the content on the server<br>3. Verify that the client content is updated | Remote content changes are synced to the client |
| IT-FS-05-01 | TestDeltaContentChangeBoth | TestIT_FS_05_01_Delta_ConflictingChanges_LocalChangesPreserved | Tests handling of conflicting changes | 1. Create a file with initial content<br>2. Change the content both locally and remotely<br>3. Verify that local changes are preserved | Local changes are preserved when there are conflicts |
| IT-FS-06-01 | TestDeltaBadContentInCache | TestIT_FS_06_01_Delta_CorruptedCache_ContentIsRestored | Tests handling of corrupted cache content | 1. Create a file with correct content<br>2. Corrupt the cache content<br>3. Verify that the correct content is restored | Corrupted cache content is detected and fixed |
| IT-FS-07-01 | TestDeltaFolderDeletion | TestIT_FS_07_01_Delta_FolderDeletion_EmptyFoldersAreDeleted | Tests folder deletion during sync | 1. Create a nested directory structure<br>2. Delete the folder on the server<br>3. Verify that the folder is deleted on the client | Folders are deleted when empty |
| IT-FS-08-01 | TestDeltaFolderDeletionNonEmpty | TestIT_FS_08_01_Delta_NonEmptyFolderDeletion_FolderIsPreserved | Tests handling of non-empty folder deletion | 1. Create a folder with files<br>2. Attempt to delete the folder via delta sync<br>3. Verify that the folder is not deleted until empty | Non-empty folders are not deleted |
| IT-FS-09-01 | TestDeltaNoModTimeUpdate | TestIT_FS_09_01_Delta_UnchangedContent_ModTimeIsPreserved | Tests preservation of modification times | 1. Create a file with initial content<br>2. Wait for delta sync to run multiple times<br>3. Verify that the modification time is not updated | Modification times are preserved when content is unchanged |
| IT-FS-10-01 | TestDeltaMissingHash | TestIT_FS_10_01_Delta_MissingHash_HandledCorrectly | Tests handling of deltas with missing hash | 1. Create a file in the filesystem<br>2. Apply a delta with missing hash information<br>3. Verify that the delta is applied without errors | Deltas with missing hash information are handled correctly |
| IT-FS-11-01 | TestDBusServerOperations | TestIT_FS_11_01_DBusServer_StartStop_OperatesCorrectly | Tests D-Bus server start/stop operations | 1. Create a D-Bus server<br>2. Perform start/stop operations<br>3. Verify the server state after each operation | D-Bus server start/stop operations work correctly |
| IT-FS-12-01 | TestReaddir | TestIT_FS_12_01_Directory_ReadContents_EntriesCorrectlyReturned | Tests reading directory contents | 1. Create a test directory with files<br>2. Call Readdir on the directory<br>3. Check if the returned entries match the expected files | Directory entries are correctly returned |
| IT-FS-13-01 | TestLs | TestIT_FS_13_01_Directory_ListContents_OutputMatchesExpected | Tests listing directory contents using ls command | 1. Create a test directory with files<br>2. Run ls command on the directory<br>3. Check if the output matches the expected files | Directory contents are correctly listed |
| IT-FS-14-01 | TestTouchOperations | TestIT_FS_14_01_Touch_CreateAndUpdate_FilesCorrectlyModified | Tests creating and updating files using touch command | 1. Run touch command to create a new file<br>2. Run touch command to update an existing file<br>3. Check if the files are created and updated correctly | Files are correctly created and updated |
| IT-FS-15-01 | TestFilePermissions | TestIT_FS_15_01_File_ChangePermissions_PermissionsCorrectlyApplied | Tests file permission operations | 1. Create a file with specific permissions<br>2. Change the file permissions<br>3. Check if the permissions are correctly applied | File permissions are correctly applied |
| IT-FS-16-01 | TestDirectoryOperations | TestIT_FS_16_01_Directory_CreateAndModify_OperationsSucceed | Tests directory creation and modification | 1. Create a directory<br>2. Create subdirectories<br>3. Check if the directories are correctly created | Directories are correctly created and modified |
| IT-FS-17-01 | TestDirectoryRemoval | TestIT_FS_17_01_Directory_Remove_DirectoryIsDeleted | Tests directory removal | 1. Create a directory with files<br>2. Remove the directory<br>3. Check if the directory is correctly removed | Directories are correctly removed |
| IT-FS-18-01 | TestFileOperations | TestIT_FS_18_01_File_BasicOperations_DataCorrectlyManaged | Tests file creation, reading, and writing | 1. Create a file<br>2. Write data to the file<br>3. Read data from the file<br>4. Check if the data matches | Files are correctly created, read, and written |
| IT-FS-19-01 | TestWriteOffset | TestIT_FS_19_01_File_WriteAtOffset_DataCorrectlyPositioned | Tests writing to a file at a specific offset | 1. Create a file with initial content<br>2. Write data at a specific offset<br>3. Check if the data is correctly written at the offset | Data is correctly written at the specified offset |
| IT-FS-20-01 | TestFileMovementOperations | TestIT_FS_20_01_File_MoveAndRename_FileCorrectlyRelocated | Tests moving and renaming files | 1. Create a file<br>2. Move the file to a new location<br>3. Check if the file is correctly moved | Files are correctly moved and renamed |
| IT-FS-21-01 | TestPositionalFileOperations | TestIT_FS_21_01_File_PositionalOperations_WorkCorrectly | Tests reading and writing at specific positions in a file | 1. Create a file with initial content<br>2. Read and write at specific positions<br>3. Check if the operations are correctly performed | Positional read and write operations work correctly |
| IT-FS-22-01 | TestBasicFileSystemOperations | TestIT_FS_22_01_FileSystem_BasicOperations_WorkCorrectly | Tests basic filesystem operations | 1. Create files and directories<br>2. Perform basic operations (read, write, delete)<br>3. Check if the operations are correctly performed | Basic filesystem operations work correctly |
| IT-FS-23-01 | TestCaseSensitivityHandling | TestIT_FS_23_01_Filename_CaseSensitivity_HandledCorrectly | Tests handling of case sensitivity in filenames | 1. Create files with similar names but different case<br>2. Perform operations on these files<br>3. Check if the operations respect case sensitivity | Case sensitivity is correctly handled |
| IT-FS-24-01 | TestFilenameCase | TestIT_FS_24_01_Filename_Case_PreservedCorrectly | Tests preservation of filename case | 1. Create files with specific case in names<br>2. Check if the case is preserved<br>3. Perform operations that might affect case | Filename case is correctly preserved |
| IT-FS-25-01 | TestShellFileOperations | TestIT_FS_25_01_Shell_FileOperations_WorkCorrectly | Tests file operations performed through shell commands | 1. Run shell commands to create, modify, and delete files<br>2. Check if the operations are correctly performed | Shell file operations work correctly |
| IT-FS-26-01 | TestFileInfo | TestIT_FS_26_01_File_GetInfo_AttributesCorrectlyRetrieved | Tests retrieving file information | 1. Create files with specific attributes<br>2. Retrieve file information<br>3. Check if the information matches the expected attributes | File information is correctly retrieved |
| IT-FS-27-01 | TestNoQuestionMarks | TestIT_FS_27_01_Filename_QuestionMarks_HandledCorrectly | Tests handling of question marks in filenames | 1. Create files with question marks in names<br>2. Check if the files are correctly handled | Question marks in filenames are correctly handled |
| IT-FS-28-01 | TestGIOTrash | TestIT_FS_28_01_GIO_TrashIntegration_WorksCorrectly | Tests integration with GIO trash functionality | 1. Create files<br>2. Move files to trash using GIO<br>3. Check if the files are correctly moved to trash | GIO trash integration works correctly |
| IT-FS-29-01 | TestListChildrenPaging | TestIT_FS_29_01_Directory_ListWithPaging_AllFilesListed | Tests paging when listing directory contents | 1. Create a directory with many files<br>2. List the directory contents with paging<br>3. Check if all files are correctly listed | Directory listing with paging works correctly |
| IT-FS-30-01 | TestLibreOfficeSavePattern | TestIT_FS_30_01_LibreOffice_SavePattern_HandledCorrectly | Tests handling of LibreOffice save pattern | 1. Simulate LibreOffice save operations<br>2. Check if the operations are correctly handled | LibreOffice save pattern is correctly handled |
| IT-FS-31-01 | TestDisallowedFilenames | TestIT_FS_31_01_Filename_Disallowed_CorrectlyRejected | Tests handling of disallowed filenames | 1. Attempt to create files with disallowed names<br>2. Check if the operations are correctly rejected | Disallowed filenames are correctly rejected |
| IT-FS-32-01 | TestOfflineFileAccess | TestIT_FS_32_01_Offline_FileAccess_FilesAccessible | Tests that files and directories can be accessed in offline mode | 1. Read directory contents in offline mode<br>2. Find and access specific files<br>3. Verify file contents match expected values | Files and directories can be accessed in offline mode |
| IT-FS-33-01 | TestOfflineFileSystemOperations | TestIT_FS_33_01_Offline_FileSystemOperations_OperationsSucceed | Tests various file and directory operations in offline mode | 1. Create files and directories in offline mode<br>2. Modify files in offline mode<br>3. Delete files and directories in offline mode<br>4. Verify operations succeed | File and directory operations succeed in offline mode |
| IT-FS-34-01 | TestOfflineChangesCached | TestIT_FS_34_01_Offline_Changes_CorrectlyCached | Tests that changes made in offline mode are cached | 1. Create a file in offline mode<br>2. Verify the file exists and has the correct content<br>3. Verify the file is marked as changed in the filesystem | Changes made in offline mode are cached |
| IT-FS-35-01 | TestOfflineSynchronization | TestIT_FS_35_01_Offline_GoingOnline_FilesSynchronized | Tests that when going back online, files are synchronized | 1. Create a file in offline mode<br>2. Verify the file exists and has the correct content<br>3. Simulate going back online<br>4. Verify the file is synchronized with the server | Files are synchronized when going back online |
| IT-FS-36-01 | TestSyncDirectoryTree | TestIT_FS_36_01_DirectoryTree_Sync_TreeSuccessfullySynchronized | Tests the SyncDirectoryTree function | 1. Create a test directory structure<br>2. Clear the filesystem metadata cache<br>3. Call SyncDirectoryTree<br>4. Verify that the directories are cached in the filesystem metadata | Directory tree is successfully synchronized |
| IT-FS-37-01 | TestUploadDiskSerialization | TestIT_FS_37_01_Upload_DiskSerialization_UploadsCanBeResumed | Tests that uploads are serialized to disk for resuming later | 1. Create a test file<br>2. Wait for the upload session to be created and serialized to disk<br>3. Cancel the upload before it completes<br>4. Create a new UploadManager from scratch<br>5. Verify the file is uploaded | Uploads are properly serialized to disk and can be resumed |
| IT-FS-38-01 | TestRepeatedUploads | TestIT_FS_38_01_Upload_RepeatedModifications_AllChangesUploaded | Tests uploading the same file multiple times | 1. Create a test file with initial content<br>2. Wait for the file to be uploaded<br>3. Modify the file multiple times<br>4. Verify each modification is successfully uploaded | Multiple uploads of the same file work correctly |
| IT-FS-39-01 | TestUploadSessionOperations | TestIT_FS_39_01_UploadSession_VariousOperations_WorkCorrectly | Tests various upload session operations | 1. Test direct uploads using internal functions<br>2. Test small file uploads using the filesystem interface<br>3. Test large file uploads using the filesystem interface<br>4. Verify uploads are successful and content is correct | Upload sessions work correctly for different file sizes and methods |
| IT-FS-40-01 | TestXattrOperations | TestIT_FS_40_01_XAttr_BasicOperations_WorkCorrectly | Tests extended attribute operations | 1. Define test cases for different xattr operations<br>2. Create test files and directories<br>3. Perform operations such as setting, getting, and listing xattrs<br>4. Verify the operations work correctly | Extended attribute operations work correctly |
| IT-FS-41-01 | TestFileStatusXattr | TestIT_FS_41_01_XAttr_FileStatus_StatusCorrectlyReported | Tests file status extended attributes | 1. Create test files with different statuses<br>2. Get the file status xattr for each file<br>3. Verify the status matches the expected value | File status extended attributes work correctly |
| IT-FS-42-01 | TestFilesystemXattrOperations | TestIT_FS_42_01_XAttr_FilesystemOperations_WorkCorrectly | Tests filesystem-level extended attribute operations | 1. Create a test filesystem<br>2. Perform xattr operations through the filesystem interface<br>3. Verify the operations work correctly | Filesystem-level extended attribute operations work correctly |

#### internal/ui/systemd Package

| Test ID | Original Test Name | New Test Function Name | Description | Test Steps | Expected Result |
|---------|-------------------|------------------------|-------------|------------|-----------------|
| IT-UI-01-01 | TestUnitEnabled | TestIT_UI_01_01_SystemdUnit_Enabled_StateCorrectlyDetermined | Tests checking if a systemd unit is enabled | 1. Define test cases with different unit states<br>2. Mock D-Bus responses for each test case<br>3. Call UnitEnabled with each test case<br>4. Verify the result matches the expected state | UnitEnabled correctly determines if a unit is enabled |
| IT-UI-02-01 | TestUnitActive | TestIT_UI_02_01_SystemdUnit_Active_StateCorrectlyDetermined | Tests checking if a systemd unit is active | 1. Define test cases with different unit states<br>2. Mock D-Bus responses for each test case<br>3. Call UnitActive with each test case<br>4. Verify the result matches the expected state | UnitActive correctly determines if a unit is active |

## Detailed Test Case Definitions

### Unit Tests

#### OM-UNIT-CMD-001: TestXDGVolumeInfo

**Description**: Tests reading and writing .xdg-volume-info files.

**Test Steps**:
1. Create a temporary file in the test mount point.
2. Generate XDG volume info content with a specific volume name using TemplateXDGVolumeInfo.
3. Write the content to the temporary file.
4. Read the volume name from the file using GetXDGVolumeInfoName.
5. Compare the read name with the original name.

**Expected Result**: The read name should match the original name.

#### OM-UNIT-CMD-002: TestLoadConfig

**Description**: Tests loading configuration from a file.

**Test Steps**:
1. Load configuration from a test config file (config-test.yml).
2. Get the user's home directory.
3. Check if the CacheDir in the loaded configuration matches the expected path (home/somewhere/else).
4. Check if the LogLevel in the loaded configuration matches the expected value ("warn").

**Expected Result**: The configuration values should match the expected values from the config file.

#### OM-UNIT-CMD-003: TestConfigMerge

**Description**: Tests merging configuration settings.

**Test Steps**:
1. Load configuration from a test config file with merged settings (config-test-merge.yml).
2. Check if the LogLevel in the loaded configuration matches the expected value ("debug").
3. Check if the CacheDir in the loaded configuration matches the expected value ("/some/directory").

**Expected Result**: The configuration should contain the merged values.

#### OM-UNIT-CMD-004: TestLoadNonexistentConfig

**Description**: Tests loading default configuration when the config file doesn't exist.

**Test Steps**:
1. Load configuration from a nonexistent config file (does-not-exist.yml).
2. Get the user's home directory.
3. Check if the CacheDir in the loaded configuration matches the default path (home/.cache/onemount).
4. Check if the LogLevel in the loaded configuration matches the default value ("debug").

**Expected Result**: The configuration should contain the default values.

#### OM-UNIT-CMD-005: TestWriteConfig

**Description**: Tests writing a configuration file.

**Test Steps**:
1. Load configuration from a test config file (config-test.yml).
2. Write the configuration to a new file in a nested directory.
3. Check if the write operation succeeds without errors.
4. Set up cleanup to remove the test directory after the test completes.

**Expected Result**: The configuration should be successfully written to the file without errors.

#### OM-UNIT-GRAPH-001: TestResourcePath

**Description**: Tests the ResourcePath function with various inputs.

**Test Steps**:
1. Define test cases with different path inputs and expected outputs.
2. For each test case:
   a. Call ResourcePath with the input path.
   b. Compare the result with the expected output.

**Expected Result**: The escaped path should match the expected format for Microsoft Graph API.

#### OM-UNIT-GRAPH-002: TestRequestUnauthenticated

**Description**: Tests the behavior of an unauthenticated request.

**Test Steps**:
1. Create an Auth object with an expiration time set far in the future.
2. Attempt to make a GET request to "/me/drive/root" with the invalid Auth object.
3. Check if an error is returned.

**Expected Result**: An error should be returned for the unauthenticated request.

#### OM-UNIT-FS-041: TestInodeCreation

**Description**: Tests that inodes are created with the correct properties.

**Test Steps**:
1. Define test cases with different modes (regular file, directory, executable file).
2. For each test case:
   a. Create an inode with the specified name, mode, and parent.
   b. Verify that the inode has the correct properties:
      - The ID begins with "local-"
      - The name matches the expected value
      - The mode matches the expected value
      - IsDir returns the correct value based on the mode

**Expected Result**: Inodes are created with the correct properties (ID, name, mode, directory status).

#### OM-UNIT-FS-042: TestInodeProperties

**Description**: Tests various properties of inodes, including mode and directory detection.

**Test Steps**:
1. Define test cases for different types of items (directory, file).
2. For each test case:
   a. Set up the test by creating the necessary directories or files.
   b. Get the item from the server.
   c. Create an inode from the drive item.
   d. Test the mode and IsDir methods.
   e. Clean up the test resources.

**Expected Result**: Inodes have the correct mode and directory status.

#### OM-UNIT-FS-043: TestFilenameEscaping

**Description**: Tests that filenames with special characters are properly escaped and can be successfully uploaded to the server.

**Test Steps**:
1. Define test cases with different special characters in filenames (hash, question mark, asterisk, etc.).
2. For each test case:
   a. Create a file with the special character in its name.
   b. Write content to the file.
   c. Verify the file is created successfully.
   d. Verify the file is uploaded to the server.
   e. Verify the file content matches what was written.

**Expected Result**: Files with special characters in their names are properly handled.

#### OM-UNIT-FS-044: TestFileCreationBehavior

**Description**: Tests various behaviors when creating files, including creating a file that already exists.

**Test Steps**:
1. Define test cases for different file creation scenarios.
2. For each test case:
   a. Get the parent directory.
   b. Create a file for the first time.
   c. Get the child after first creation.
   d. Store the original ID.
   e. For some test cases, write content to the file.
   f. Create the file for the second time.
   g. Get the child after second creation.
   h. Verify the ID is the same.
   i. For some test cases, verify the file was truncated.

**Expected Result**: File creation behavior is correct, including truncation and returning the same inode.

#### OM-UNIT-FS-045: TestLogging

**Description**: Tests the LogMethodCall and LogMethodReturn functions.

**Test Steps**:
1. Capture log output.
2. Call LogMethodCall and store the method name and start time.
3. Simulate some work by sleeping.
4. Call LogMethodReturn with a boolean return value.
5. Verify the log output contains "Method called", "Method completed", "return1", and "goroutine".
6. Reset the buffer.
7. Call LogMethodCall again.
8. Call LogMethodReturn with multiple return values (string, int, nil).
9. Verify the log output contains "Method called", "Method completed", "return1", "return2", and "return3".
10. Reset the buffer.
11. Call LogMethodCall again.
12. Call LogMethodReturn with a struct return value.
13. Verify the log output contains "Method called", "Method completed", and "return1".

**Expected Result**: Logging functions correctly log method calls and returns.

#### OM-UNIT-FS-046: TestGetCurrentGoroutineID

**Description**: Tests the getCurrentGoroutineID function.

**Test Steps**:
1. Call getCurrentGoroutineID.
2. Verify the result is not empty.
3. Verify the result is a valid number by converting it to an integer.

**Expected Result**: getCurrentGoroutineID returns a valid goroutine ID.

#### OM-UNIT-FS-047: TestLoggingInMethods

**Description**: Tests the logging in actual methods.

**Test Steps**:
1. Capture log output.
2. Create a test filesystem.
3. Call the IsOffline method.
4. Verify the log output contains "Method called", "method=IsOffline", "phase=entry", "goroutine", "Method completed", "phase=exit", and "return1=false".
5. Reset the buffer.
6. Call the GetNodeID method with a non-existent ID.
7. Verify the log output contains "Method called", "method=GetNodeID", "phase=entry", "Method completed", "phase=exit", and "return1=null".

**Expected Result**: Methods correctly log their calls and returns.

#### OM-UNIT-FS-048: TestThumbnailCacheOperations

**Description**: Tests various operations on the thumbnail cache.

**Test Steps**:
1. Define test cases for different thumbnail cache operations.
2. For each test case:
   a. Create a temporary directory for the thumbnail cache.
   b. Create a thumbnail cache.
   c. Set up test data (ID, sizes, contents).
   d. Run the test function, which may include:
      - Inserting thumbnails
      - Checking if thumbnails exist
      - Retrieving thumbnails
      - Deleting thumbnails
   e. Clean up the temporary directory.

**Expected Result**: Thumbnail cache operations work correctly.

#### OM-UNIT-FS-049: TestThumbnailCacheCleanup

**Description**: Tests the cleanup functionality of the thumbnail cache.

**Test Steps**:
1. Define test cases for different cleanup scenarios.
2. For each test case:
   a. Create a temporary directory for the thumbnail cache.
   b. Create a thumbnail cache.
   c. Set up test data, which may include:
      - Inserting thumbnails
      - Setting the last cleanup time to a long time ago
      - Setting the modification time of thumbnails
   d. Verify thumbnails exist before cleanup.
   e. Run cleanup with the specified expiration time.
   f. Verify the number of removed thumbnails matches the expected count.
   g. Verify thumbnails exist or not based on the expected count.
   h. Clean up the temporary directory.

**Expected Result**: Thumbnail cache cleanup correctly removes expired thumbnails.

#### OM-UNIT-FS-050: TestThumbnailOperations

**Description**: Tests various operations on thumbnails in the filesystem.

**Test Steps**:
1. Skip the test if no valid auth token is available.
2. Define test cases for different thumbnail operations.
3. For each test case:
   a. Create a temporary directory for the filesystem.
   b. Create a filesystem.
   c. Find an image file to test with.
   d. Run the test function, which may include:
      - Getting thumbnails of different sizes
      - Verifying thumbnails are cached
      - Deleting thumbnails
      - Verifying thumbnails are removed from the cache
   e. Clean up the temporary directory.

**Expected Result**: Thumbnail operations in the filesystem work correctly.

#### OM-UNIT-UI-001: TestMountpointIsValid

**Description**: Tests that mountpoints are validated correctly.

**Test Steps**:
1. Create a test directory and file.
2. Define test cases with different paths and expected validation results.
3. For each test case:
   a. Call MountpointIsValid with the path.
   b. Compare the result with the expected validation result.

**Expected Result**: Valid mountpoints should return true, invalid mountpoints should return false.

#### OM-UNIT-UI-002: TestHomeEscapeUnescape

**Description**: Tests converting paths from ~/some_path to /home/username/some_path and back.

**Test Steps**:
1. Get the user's home directory.
2. Define test cases with different paths and expected escaped/unescaped results.
3. For each test case:
   a. Call EscapeHome with the unescaped path.
   b. Compare the result with the expected escaped path.
   c. Call UnescapeHome with the escaped path.
   d. Compare the result with the expected unescaped path.

**Expected Result**: Paths should be correctly escaped and unescaped.

#### OM-UNIT-UI-003: TestGetAccountName

**Description**: Tests retrieving account names from auth token files.

**Test Steps**:
1. Create a test instance directory.
2. Create various auth token files (valid, invalid, empty).
3. Define test cases with different instances and expected results.
4. For each test case:
   a. Call GetAccountName with the instance.
   b. Compare the result with the expected account name or error.

**Expected Result**: Account names should be correctly retrieved from valid files, errors should be returned for invalid files.

#### OM-UNIT-UI-004: TestTemplateUnit

**Description**: Tests the TemplateUnit function that converts a unit name and instance name into a templated unit name.

**Test Steps**:
1. Define test cases with different unit names and instance names:
   - Regular unit name with simple instance name
   - Unit name with @ character with simple instance name
   - Unit name with @ character with instance name containing special characters
2. For each test case:
   a. Call TemplateUnit with the unit name and instance name
   b. Compare the result with the expected templated unit name
3. Verify that the templated unit name follows the systemd convention (unit@instance.service)

**Expected Result**: Unit names are correctly templated with instance names, following the systemd convention.

#### OM-UNIT-UI-005: TestUntemplateUnit

**Description**: Tests the UntemplateUnit function that extracts the unit name and instance name from a templated unit name.

**Test Steps**:
1. Define test cases with different templated unit names:
   - Regular templated unit name (unit@instance.service)
   - Templated unit name with special characters in the instance name
   - Non-templated unit name
2. For each test case:
   a. Call UntemplateUnit with the templated unit name
   b. Compare the results (unit name and instance name) with the expected values
3. Verify that the function correctly handles both templated and non-templated unit names

**Expected Result**: Templated unit names are correctly untemplated into unit name and instance name, and non-templated unit names are handled appropriately.

#### OM-UNIT-GRAPH-003: TestDriveItemIsDir

**Description**: Tests the IsDir method of DriveItem.

**Test Steps**:
1. Define test cases with different DriveItem types:
   - A folder item with a Folder field
   - A file item with a File field
   - An empty item with neither field
2. For each test case:
   a. Call IsDir on the DriveItem
   b. Compare the result with the expected value
3. Verify that IsDir returns true for folders and false for files and empty items.

**Expected Result**: IsDir should return true for items with a Folder field and false for items with a File field or empty items.

#### OM-UNIT-GRAPH-004: TestDriveItemModTimeUnix

**Description**: Tests the ModTimeUnix method of DriveItem.

**Test Steps**:
1. Create a fixed time for testing.
2. Create a DriveItem with the ModTime field set to the fixed time.
3. Call ModTimeUnix on the DriveItem.
4. Compare the result with the expected Unix timestamp.

**Expected Result**: ModTimeUnix should return the correct Unix timestamp for the ModTime field.

#### OM-UNIT-GRAPH-005: TestDriveItemVerifyChecksum

**Description**: Tests the VerifyChecksum method of DriveItem.

**Test Steps**:
1. Define test cases with different DriveItem and checksum combinations:
   - A DriveItem with a matching checksum
   - A DriveItem with a non-matching checksum
   - A DriveItem with a case-insensitive matching checksum
   - A DriveItem with an empty checksum
   - A DriveItem with a nil File field
2. For each test case:
   a. Call VerifyChecksum on the DriveItem with the test checksum
   b. Compare the result with the expected value

**Expected Result**: VerifyChecksum should return true for matching checksums (case-insensitive) and false for non-matching checksums, empty checksums, or when the File field is nil.

#### OM-UNIT-GRAPH-006: TestDriveItemETagIsMatch

**Description**: Tests the ETagIsMatch method of DriveItem.

**Test Steps**:
1. Define test cases with different DriveItem and ETag combinations:
   - A DriveItem with a matching ETag
   - A DriveItem with a non-matching ETag
   - A DriveItem with an empty ETag
   - A DriveItem with an empty ETag parameter
2. For each test case:
   a. Call ETagIsMatch on the DriveItem with the test ETag
   b. Compare the result with the expected value

**Expected Result**: ETagIsMatch should return true for matching ETags and false for non-matching ETags, empty ETags, or when the ETag parameter is empty.

#### OM-UNIT-GRAPH-007: TestGetItem

**Description**: Tests retrieving items from the Microsoft Graph API.

**Test Steps**:
1. Define test cases with different paths:
   - The root path ("/")
   - A nonexistent path
   - A valid path ("/Documents")
2. For each test case:
   a. Load authentication tokens
   b. Call GetItemPath with the test path
   c. Check if the result matches expectations

**Expected Result**: GetItemPath should return the correct item for valid paths and an error for invalid paths.

#### OM-UNIT-GRAPH-008: TestSha1HashReader

**Description**: Tests the SHA1HashStream function with a reader.

**Test Steps**:
1. Create a byte array with test content.
2. Calculate the SHA1 hash of the content using SHA1Hash.
3. Create a reader from the content.
4. Calculate the SHA1 hash using SHA1HashStream.
5. Compare the two hashes.

**Expected Result**: The hash calculated from the reader should match the hash calculated from the byte array.

#### OM-UNIT-GRAPH-009: TestQuickXORHashReader

**Description**: Tests the QuickXORHashStream function with a reader.

**Test Steps**:
1. Create a byte array with test content.
2. Calculate the QuickXOR hash of the content using QuickXORHash.
3. Create a reader from the content.
4. Calculate the QuickXOR hash using QuickXORHashStream.
5. Compare the two hashes.

**Expected Result**: The hash calculated from the reader should match the hash calculated from the byte array.

#### OM-UNIT-GRAPH-010: TestHashSeekPosition

**Description**: Tests that hash functions reset the seek position.

**Test Steps**:
1. Create a temporary file with test content.
2. Read a portion of the file to move the seek position.
3. Verify that the seek position is not at the beginning.
4. Calculate the QuickXOR hash using QuickXORHashStream.
5. Verify that the seek position is reset to the beginning.
6. Repeat steps 2-5 for SHA1HashStream and SHA256HashStream.

**Expected Result**: The seek position should be reset to the beginning after each hash calculation.

#### OM-UNIT-GRAPH-011: TestSHA256Hash

**Description**: Tests the SHA256Hash function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Calculate the SHA256 hash of the content
   b. Compare the result with the expected hash

**Expected Result**: SHA256Hash should return the correct hash for each input.

#### OM-UNIT-GRAPH-012: TestSHA256HashStream

**Description**: Tests the SHA256HashStream function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Create a reader from the content
   b. Calculate the SHA256 hash using SHA256HashStream
   c. Compare the result with the expected hash

**Expected Result**: SHA256HashStream should return the correct hash for each input.

#### OM-UNIT-GRAPH-013: TestSHA1Hash

**Description**: Tests the SHA1Hash function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Calculate the SHA1 hash of the content
   b. Compare the result with the expected hash

**Expected Result**: SHA1Hash should return the correct hash for each input.

#### OM-UNIT-GRAPH-014: TestSHA1HashStream

**Description**: Tests the SHA1HashStream function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Create a reader from the content
   b. Calculate the SHA1 hash using SHA1HashStream
   c. Compare the result with the expected hash

**Expected Result**: SHA1HashStream should return the correct hash for each input.

#### OM-UNIT-GRAPH-015: TestQuickXORHash

**Description**: Tests the QuickXORHash function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Calculate the QuickXOR hash of the content
   b. Compare the result with the expected hash

**Expected Result**: QuickXORHash should return the correct hash for each input.

#### OM-UNIT-GRAPH-016: TestQuickXORHashStream

**Description**: Tests the QuickXORHashStream function with different inputs.

**Test Steps**:
1. Define test cases with different content:
   - Empty data
   - A simple string
   - A longer text
2. For each test case:
   a. Create a reader from the content
   b. Calculate the QuickXOR hash using QuickXORHashStream
   c. Compare the result with the expected hash

**Expected Result**: QuickXORHashStream should return the correct hash for each input.

#### OM-UNIT-GRAPH-017: TestURIGetHost

**Description**: Tests the uriGetHost function with various inputs.

**Test Steps**:
1. Call uriGetHost with an invalid URI ("this won't work").
2. Verify that it returns an empty string.
3. Call uriGetHost with a valid HTTPS URI with a path ("https://account.live.com/test/index.html").
4. Verify that it returns "account.live.com".
5. Call uriGetHost with a valid HTTP URI without a path ("http://account.live.com").
6. Verify that it returns "account.live.com".

**Expected Result**: uriGetHost should return the correct host for valid URIs and an empty string for invalid URIs.

#### OM-UNIT-GRAPH-018: TestAuthCodeFormat

**Description**: Tests the parseAuthCode function with various inputs.

**Test Steps**:
1. Define test cases with different input formats:
   - An arbitrary format with a code
   - A personal auth code URL
   - A business auth code URL
   - An invalid format without a code
2. For each test case:
   a. Call parseAuthCode with the input
   b. Check if the result matches the expected code or error

**Expected Result**: parseAuthCode should correctly extract the authorization code from different URL formats and return an error for invalid formats.

#### OM-UNIT-GRAPH-019: TestAuthFromfile

**Description**: Tests loading authentication tokens from a file.

**Test Steps**:
1. Verify that the auth tokens file exists.
2. Create a new Auth object.
3. Call FromFile with the path to the auth tokens file.
4. Check if the access token is not empty.

**Expected Result**: Authentication tokens should be successfully loaded from the file.

#### OM-UNIT-GRAPH-020: TestAuthRefresh

**Description**: Tests refreshing authentication tokens.

**Test Steps**:
1. Verify that the auth tokens file exists.
2. Create a new Auth object.
3. Call FromFile with the path to the auth tokens file.
4. Force an auth refresh by setting ExpiresAt to 0.
5. Call Refresh with a nil context.
6. Check if the new expiration time is in the future.

**Expected Result**: Authentication tokens should be successfully refreshed.

#### OM-UNIT-GRAPH-021: TestAuthConfigMerge

**Description**: Tests merging authentication configuration with default values.

**Test Steps**:
1. Create a test AuthConfig with a custom RedirectURL.
2. Call applyDefaults on the AuthConfig.
3. Check if the RedirectURL is preserved.
4. Check if the ClientID is set to the default value.

**Expected Result**: Default values should be correctly applied while preserving custom values.

#### OM-UNIT-GRAPH-022: TestAuthFailureWithNetworkAvailable

**Description**: Tests the behavior when authentication fails but network is available.

**Test Steps**:
1. Create an Auth with invalid credentials but valid configuration.
2. Apply defaults to the AuthConfig.
3. Attempt to refresh the tokens.
4. Check if an error is returned.
5. Verify that the auth state is still invalid (tokens not updated).

**Expected Result**: An error should be returned and the auth state should remain invalid.

#### OM-UNIT-GRAPH-023: TestOperationalOfflineState

**Description**: Tests setting and getting the operational offline state.

**Test Steps**:
1. Reset the operational offline state to false.
2. Check that the default state is false.
3. Set the state to true and check that it's true.
4. Set the state back to false and check that it's false.
5. Set up cleanup to reset the state after the test.

**Expected Result**: The operational offline state should be correctly set and retrieved.

#### OM-UNIT-GRAPH-024: TestIsOfflineWithOperationalState

**Description**: Tests the IsOffline function when operational offline state is set.

**Test Steps**:
1. Reset the operational offline state to false.
2. Set the operational offline state to true.
3. Call IsOffline with nil error and verify it returns true.
4. Call IsOffline with an HTTP error and verify it returns true.
5. Reset the operational offline state to false.
6. Call IsOffline with nil error and verify it returns false.
7. Call IsOffline with an HTTP error and verify it returns false.
8. Call IsOffline with a network error and verify it returns true.
9. Set up cleanup to reset the state after the test.

**Expected Result**: IsOffline should return true when operational offline is set, regardless of the error. When operational offline is not set, it should return true only for network-related errors.

#### OM-UNIT-GRAPH-025: TestIsOffline

**Description**: Tests the IsOffline function with various error types.

**Test Steps**:
1. Reset the operational offline state to false.
2. Define test cases with different error types:
   - nil error
   - HTTP error
   - HTTP error with different format
   - network error
   - timeout error
   - connection refused error
3. For each test case:
   a. Call IsOffline with the error
   b. Check if the result matches the expected value
4. Set up cleanup to reset the state after the test.

**Expected Result**: IsOffline should correctly identify network-related errors and return true for them, and false for other types of errors.

#### OM-UNIT-GRAPH-026: TestIDPath

**Description**: Tests the IDPath function with various inputs.

**Test Steps**:
1. Define test cases with different item IDs:
   - root ID
   - regular ID
   - ID with special characters
2. For each test case:
   a. Call IDPath with the ID
   b. Check if the result matches the expected path

**Expected Result**: IDPath should correctly format item IDs for API requests.

#### OM-UNIT-GRAPH-027: TestChildrenPath

**Description**: Tests the childrenPath function with various inputs.

**Test Steps**:
1. Define test cases with different paths:
   - root path
   - simple path
   - nested path
   - path with spaces
2. For each test case:
   a. Call childrenPath with the path
   b. Check if the result matches the expected path

**Expected Result**: childrenPath should correctly format paths for retrieving children.

#### OM-UNIT-GRAPH-028: TestChildrenPathID

**Description**: Tests the childrenPathID function with various inputs.

**Test Steps**:
1. Define test cases with different item IDs:
   - root ID
   - regular ID
   - ID with special characters
2. For each test case:
   a. Call childrenPathID with the ID
   b. Check if the result matches the expected path

**Expected Result**: childrenPathID should correctly format item IDs for retrieving children.

#### OM-UNIT-GRAPH-029: TestQuickXorHash

**Description**: Tests the Sum function with various inputs.

**Test Steps**:
1. For each test vector:
   a. Decode the input from base64
   b. Calculate the QuickXOR hash using Sum
   c. Decode the expected output from base64
   d. Compare the calculated hash with the expected hash

**Expected Result**: Sum should correctly calculate QuickXOR hashes for all test vectors.

#### OM-UNIT-GRAPH-030: TestQuickXorHashByBlock

**Description**: Tests calculating QuickXOR hashes by writing data in blocks.

**Test Steps**:
1. For each block size (1, 2, 4, 7, 8, 16, 32, 64, 128, 256, 512):
   a. For each test vector:
      i. Decode the input from base64
      ii. Create a new QuickXOR hash
      iii. Write the data in blocks of the current size
      iv. Calculate the hash
      v. Decode the expected output from base64
      vi. Compare the calculated hash with the expected hash

**Expected Result**: QuickXOR hashes should be correctly calculated when writing data in blocks of different sizes.

#### OM-UNIT-GRAPH-031: TestSize

**Description**: Tests the Size method of the QuickXOR hash.

**Test Steps**:
1. Create a new QuickXOR hash.
2. Call the Size method.
3. Check if the result is 20.

**Expected Result**: Size should return the correct hash size (20 bytes).

#### OM-UNIT-GRAPH-032: TestBlockSize

**Description**: Tests the BlockSize method of the QuickXOR hash.

**Test Steps**:
1. Create a new QuickXOR hash.
2. Call the BlockSize method.
3. Check if the result is 64.

**Expected Result**: BlockSize should return the correct block size (64 bytes).

#### OM-UNIT-GRAPH-033: TestReset

**Description**: Tests the Reset method of the QuickXOR hash.

**Test Steps**:
1. Create a new QuickXOR hash.
2. Calculate the hash of an empty input.
3. Write some data (byte 1) to the hash.
4. Verify that the hash is different from the empty hash.
5. Reset the hash.
6. Calculate the hash again.
7. Compare with the original empty hash.

**Expected Result**: Reset should correctly reset the hash state to its initial state.

### Integration Tests

#### OM-INT-FS-001: TestCacheOperations

**Description**: Tests various cache operations using a table-driven approach.

**Test Steps**:
1. Define test cases for different cache operations (get path, get children, check pointers).
2. For each test case:
   a. Create a unique database location for the test.
   b. Create a new filesystem cache.
   c. Run the verification function for the specific operation.
3. Verify that each operation produces the expected results.

**Expected Result**: All cache operations should work correctly, including getting paths, getting children, and checking that the same item returns the same pointer.

#### OM-INT-FS-002: TestDBusServerFunctionality

**Description**: Tests various functionality of the D-Bus server.

**Test Steps**:
1. Define test cases for different D-Bus server functionality (get file status, emit signals, reconnect).
2. For each test case:
   a. Set up test resources (create files, initialize D-Bus server).
   b. Perform the specific operation being tested.
   c. Verify the results of the operation.
3. Clean up test resources.

**Expected Result**: The D-Bus server should correctly handle file status requests, emit signals when file status changes, and be able to reconnect after being stopped.

#### OM-INT-FS-003: TestDeltaOperations

**Description**: Tests various delta operations for syncing changes from the server to the client.

**Test Steps**:
1. Define test cases for different delta operations (create directory, delete directory, rename file, move file).
2. For each test case:
   a. Set up the test environment (create files/directories).
   b. Perform the operation on the server.
   c. Verify that the changes are synced to the client.
3. Clean up test resources.

**Expected Result**: Changes made on the server should be correctly synced to the client, including creating directories, deleting directories, renaming files, and moving files.

#### OM-INT-FS-004: TestDeltaContentChangeRemote

**Description**: Tests that content changes on the server are propagated to the client.

**Test Steps**:
1. Create a file on the client with initial content.
2. Change the content on the server using the API.
3. Verify that the content is uploaded to the server.
4. Wait for the DeltaLoop to detect the change and update the local file.
5. Verify that the local file content matches the new content.

**Expected Result**: Content changes made on the server should be correctly synced to the client.

#### OM-INT-FS-005: TestDeltaContentChangeBoth

**Description**: Tests handling of conflicting changes when content is changed both on the server and the client.

**Test Steps**:
1. Create a file with initial content.
2. Write new content to the file locally without closing it (simulating an in-use file).
3. Apply a fake delta to simulate a change on the server.
4. Verify that the local changes are preserved while the file is open.
5. Simulate closing the file and apply the delta again.
6. Verify that the server changes are applied after the file is closed.

**Expected Result**: Local changes should be preserved when there are conflicts with a file that is currently open. Server changes should be applied after the file is closed.

#### OM-INT-FS-006: TestDeltaBadContentInCache

**Description**: Tests handling of corrupted content in the cache.

**Test Steps**:
1. Create a file on the client with correct content.
2. Upload the file to the server and get its ID.
3. Corrupt the cache content by inserting incorrect data.
4. Attempt to read the file.
5. Verify that the correct content is downloaded from the server and the corrupted cache is fixed.

**Expected Result**: Corrupted cache content should be detected and fixed by downloading the correct content from the server.

#### OM-INT-FS-007: TestDeltaFolderDeletion

**Description**: Tests folder deletion during delta synchronization.

**Test Steps**:
1. Create a nested directory structure on the client.
2. Delete the folder on the server using the API.
3. Wait for the DeltaLoop to detect the change.
4. Verify that the folder is deleted on the client.

**Expected Result**: Folders should be deleted on the client when they are deleted on the server.

#### OM-INT-FS-008: TestDeltaFolderDeletionNonEmpty

**Description**: Tests handling of non-empty folder deletion during delta synchronization.

**Test Steps**:
1. Create a folder with a file in it.
2. Create a delta that indicates the folder has been deleted.
3. Apply the delta and verify that it fails because the folder is not empty.
4. Delete the file from the folder.
5. Apply the delta again and verify that it succeeds.
6. Verify that the folder is deleted.

**Expected Result**: Non-empty folders should not be deleted during delta synchronization. The deletion should succeed after the folder is emptied.

#### OM-INT-FS-009: TestDeltaNoModTimeUpdate

**Description**: Tests that modification times are not updated if content is unchanged.

**Test Steps**:
1. Create a file with initial content.
2. Record the initial modification time.
3. Wait for the DeltaLoop to run multiple times.
4. Verify that the modification time has not changed.

**Expected Result**: Modification times should not be updated if the content is unchanged, even after multiple delta synchronization cycles.

#### OM-INT-FS-010: TestDeltaMissingHash

**Description**: Tests handling of deltas with missing hash information.

**Test Steps**:
1. Create a file in the filesystem.
2. Create a delta with missing hash information.
3. Apply the delta to the file.
4. Verify that the delta is applied without errors.

**Expected Result**: Deltas with missing hash information should be handled correctly without causing errors.

#### OM-INT-FS-011: TestDBusServerOperations

**Description**: Tests the basic operations of the D-Bus server (start, stop, etc.).

**Test Steps**:
1. Create a temporary filesystem for testing.
2. Define test cases for different server operations (check initial state, stop server, start server, stop again).
3. For each test case:
   a. Perform the operation.
   b. Verify that the server state matches the expected state.
4. Clean up test resources.

**Expected Result**: The D-Bus server should correctly handle start and stop operations, and the server state should match the expected state after each operation.

#### OM-INT-FS-021: TestReaddir

**Description**: Tests reading directory contents.

**Test Steps**:
1. Create a test directory with files.
2. Call Readdir on the directory.
3. Check if the returned entries match the expected files.

**Expected Result**: Directory entries should be correctly returned.

#### OM-INT-FS-002: TestLs

**Description**: Tests listing directory contents using ls command.

**Test Steps**:
1. Create a test directory with files.
2. Run ls command on the directory.
3. Check if the output matches the expected files.

**Expected Result**: Directory contents should be correctly listed.

#### OM-INT-FS-003: TestTouchOperations

**Description**: Tests creating and updating files using touch command.

**Test Steps**:
1. Run touch command to create a new file.
2. Run touch command to update an existing file.
3. Check if the files are created and updated correctly.

**Expected Result**: Files should be correctly created and updated.

#### OM-INT-FS-004: TestFilePermissions

**Description**: Tests file permission operations.

**Test Steps**:
1. Create a file with specific permissions.
2. Change the file permissions.
3. Check if the permissions are correctly applied.

**Expected Result**: File permissions should be correctly applied.

#### OM-INT-FS-005: TestDirectoryOperations

**Description**: Tests directory creation and modification.

**Test Steps**:
1. Create a directory.
2. Create subdirectories.
3. Check if the directories are correctly created.

**Expected Result**: Directories should be correctly created and modified.

#### OM-INT-FS-006: TestDirectoryRemoval

**Description**: Tests directory removal.

**Test Steps**:
1. Create a directory with files.
2. Remove the directory.
3. Check if the directory is correctly removed.

**Expected Result**: Directories should be correctly removed.

#### OM-INT-FS-007: TestFileOperations

**Description**: Tests file creation, reading, and writing.

**Test Steps**:
1. Create a file.
2. Write data to the file.
3. Read data from the file.
4. Check if the data matches.

**Expected Result**: Files should be correctly created, read, and written.

#### OM-INT-FS-008: TestWriteOffset

**Description**: Tests writing to a file at a specific offset.

**Test Steps**:
1. Create a file with initial content.
2. Write data at a specific offset.
3. Check if the data is correctly written at the offset.

**Expected Result**: Data should be correctly written at the specified offset.

#### OM-INT-FS-009: TestFileMovementOperations

**Description**: Tests moving and renaming files.

**Test Steps**:
1. Create a file.
2. Move the file to a new location.
3. Check if the file is correctly moved.

**Expected Result**: Files should be correctly moved and renamed.

#### OM-INT-FS-010: TestPositionalFileOperations

**Description**: Tests reading and writing at specific positions in a file.

**Test Steps**:
1. Create a file with initial content.
2. Read and write at specific positions.
3. Check if the operations are correctly performed.

**Expected Result**: Positional read and write operations should work correctly.

#### OM-INT-FS-011: TestBasicFileSystemOperations

**Description**: Tests basic filesystem operations.

**Test Steps**:
1. Create files and directories.
2. Perform basic operations (read, write, delete).
3. Check if the operations are correctly performed.

**Expected Result**: Basic filesystem operations should work correctly.

#### OM-INT-FS-012: TestCaseSensitivityHandling

**Description**: Tests handling of case sensitivity in filenames.

**Test Steps**:
1. Create files with similar names but different case.
2. Perform operations on these files.
3. Check if the operations respect case sensitivity.

**Expected Result**: Case sensitivity should be correctly handled.

#### OM-INT-FS-013: TestFilenameCase

**Description**: Tests preservation of filename case.

**Test Steps**:
1. Create files with specific case in names.
2. Check if the case is preserved.
3. Perform operations that might affect case.

**Expected Result**: Filename case should be correctly preserved.

#### OM-INT-FS-014: TestShellFileOperations

**Description**: Tests file operations performed through shell commands.

**Test Steps**:
1. Run shell commands to create, modify, and delete files.
2. Check if the operations are correctly performed.

**Expected Result**: Shell file operations should work correctly.

#### OM-INT-FS-015: TestFileInfo

**Description**: Tests retrieving file information.

**Test Steps**:
1. Create files with specific attributes.
2. Retrieve file information.
3. Check if the information matches the expected attributes.

**Expected Result**: File information should be correctly retrieved.

#### OM-INT-FS-016: TestNoQuestionMarks

**Description**: Tests handling of question marks in filenames.

**Test Steps**:
1. Create files with question marks in names.
2. Check if the files are correctly handled.

**Expected Result**: Question marks in filenames should be correctly handled.

#### OM-INT-FS-017: TestGIOTrash

**Description**: Tests integration with GIO trash functionality.

**Test Steps**:
1. Create files.
2. Move files to trash using GIO.
3. Check if the files are correctly moved to trash.

**Expected Result**: GIO trash integration should work correctly.

#### OM-INT-FS-018: TestListChildrenPaging

**Description**: Tests paging when listing directory contents.

**Test Steps**:
1. Create a directory with many files.
2. List the directory contents with paging.
3. Check if all files are correctly listed.

**Expected Result**: Directory listing with paging should work correctly.

#### OM-INT-FS-019: TestLibreOfficeSavePattern

**Description**: Tests handling of LibreOffice save pattern.

**Test Steps**:
1. Simulate LibreOffice save operations.
2. Check if the operations are correctly handled.

**Expected Result**: LibreOffice save pattern should be correctly handled.

#### OM-INT-FS-020: TestDisallowedFilenames

**Description**: Tests handling of disallowed filenames.

**Test Steps**:
1. Attempt to create files with disallowed names.
2. Check if the operations are correctly rejected.

**Expected Result**: Disallowed filenames should be correctly rejected.

#### OM-INT-FS-041: TestOfflineFileAccess

**Description**: Tests that files and directories can be accessed in offline mode.

**Test Steps**:
1. Define test cases for different offline file access scenarios.
2. For each test case:
   a. Read directory contents in offline mode.
   b. For some test cases, find and access specific files.
   c. For some test cases, verify file contents match expected values.

**Expected Result**: Files and directories can be accessed in offline mode.

#### OM-INT-FS-042: TestOfflineFileSystemOperations

**Description**: Tests various file and directory operations in offline mode.

**Test Steps**:
1. Define test cases for different file and directory operations in offline mode.
2. For each test case:
   a. Set up test resources.
   b. Perform operations such as:
      - Creating files and directories
      - Modifying files
      - Deleting files and directories
   c. Verify the operations succeed.
   d. Clean up test resources.

**Expected Result**: File and directory operations succeed in offline mode.

#### OM-INT-FS-043: TestOfflineChangesCached

**Description**: Tests that changes made in offline mode are cached.

**Test Steps**:
1. Create a test file in offline mode.
2. Verify the file exists and has the correct content.
3. Verify the file is marked as changed in the filesystem.

**Expected Result**: Changes made in offline mode are cached.

#### OM-INT-FS-044: TestOfflineSynchronization

**Description**: Tests that when going back online, files are synchronized.

**Test Steps**:
1. Create a test file in offline mode.
2. Verify the file exists and has the correct content.
3. Simulate going back online.
4. Verify the file is synchronized with the server.

**Expected Result**: Files are synchronized when going back online.

#### OM-INT-FS-045: TestSyncDirectoryTree

**Description**: Tests the SyncDirectoryTree function.

**Test Steps**:
1. Skip the test if using mock authentication.
2. Create a test directory structure if it doesn't exist.
3. Get the root directory ID.
4. Clear the filesystem metadata cache.
5. Call SyncDirectoryTree.
6. Verify that the test directories are cached in the filesystem metadata.
7. Verify that other known directories are also cached.

**Expected Result**: Directory tree is successfully synchronized.

#### OM-INT-FS-046: TestUploadDiskSerialization

**Description**: Tests that uploads are serialized to disk to support resuming them later if the user shuts down their computer.

**Test Steps**:
1. Create a test file by copying a known file.
2. Wait for the file to be recognized by the filesystem.
3. Wait for the upload session to be created and serialized to disk.
4. Cancel the upload before it completes.
5. Confirm that the file didn't get uploaded yet.
6. Create a new UploadManager from scratch with the file injected into its database.
7. Wait for the file to be uploaded.
8. Verify that the file was successfully uploaded.

**Expected Result**: Uploads are properly serialized to disk and can be resumed after creating a new UploadManager.

#### OM-INT-FS-047: TestRepeatedUploads

**Description**: Tests that uploading the same file multiple times works correctly.

**Test Steps**:
1. Create a test file with initial content.
2. Wait for the file to be recognized and uploaded.
3. Verify the file has a non-local ID after upload.
4. Test multiple uploads of the same file:
   a. Create new content for each iteration.
   b. Write the content to the file.
   c. Wait for the file to be uploaded.
   d. Verify the content on the server matches what was uploaded.

**Expected Result**: Multiple uploads of the same file work correctly, with each new content being properly uploaded.

#### OM-INT-FS-048: TestUploadSessionOperations

**Description**: Tests various upload session operations using different methods and file sizes.

**Test Steps**:
1. Test direct uploads using internal functions:
   a. Create an inode with test data.
   b. Create and upload a session.
   c. Verify the upload was successful.
   d. Test overwriting with new data.
   e. Verify the content was updated correctly.
2. Test small file uploads using the filesystem interface:
   a. Create a test file with initial data.
   b. Wait for the file to be uploaded.
   c. Verify the content was uploaded correctly.
   d. Test uploading again with new data.
   e. Verify the content was updated correctly.
3. Test large file uploads using the filesystem interface:
   a. Create a test file with large data.
   b. Verify the file was written correctly.
   c. Wait for the file to be uploaded.
   d. Verify the content was uploaded correctly.
   e. Test multipart downloads.

**Expected Result**: Upload sessions work correctly for different file sizes and methods, with content being properly uploaded and updated.

#### OM-INT-FS-049: TestXattrOperations

**Description**: Tests extended attribute operations on files and directories.

**Test Steps**:
1. Define test cases for different xattr operations:
   - Setting xattrs on files and directories
   - Getting xattrs from files and directories
   - Listing xattrs on files and directories
   - Removing xattrs from files and directories
2. For each test case:
   a. Create test files or directories.
   b. Perform the xattr operation.
   c. Verify the operation works correctly.
   d. Clean up test resources.

**Expected Result**: Extended attribute operations work correctly on both files and directories.

#### OM-INT-FS-050: TestFileStatusXattr

**Description**: Tests the file status extended attribute functionality.

**Test Steps**:
1. Create test files with different statuses:
   - Synced file
   - Locally modified file
   - Uploading file
   - Error state file
2. For each test file:
   a. Get the file status xattr.
   b. Verify the status matches the expected value.
   c. Test setting the status to different values.
   d. Verify the status is updated correctly.

**Expected Result**: File status extended attributes correctly reflect the file's status and can be updated.

#### OM-INT-FS-051: TestFilesystemXattrOperations

**Description**: Tests filesystem-level extended attribute operations.

**Test Steps**:
1. Create a test filesystem.
2. Perform xattr operations through the filesystem interface:
   - Setting xattrs
   - Getting xattrs
   - Listing xattrs
   - Removing xattrs
3. Verify the operations work correctly.
4. Test error handling for invalid operations.

**Expected Result**: Filesystem-level extended attribute operations work correctly and handle errors appropriately.

#### OM-INT-UI-001: TestUnitEnabled

**Description**: Tests checking if a systemd unit is enabled.

**Test Steps**:
1. Define test cases with different unit states:
   - Enabled unit
   - Disabled unit
   - Static unit
   - Masked unit
   - Non-existent unit
2. For each test case:
   a. Mock D-Bus responses to simulate the unit state.
   b. Call UnitEnabled with the unit name.
   c. Verify the result matches the expected state.
3. Test error handling for D-Bus connection failures.

**Expected Result**: UnitEnabled correctly determines if a unit is enabled and handles errors appropriately.

#### OM-INT-UI-002: TestUnitActive

**Description**: Tests checking if a systemd unit is active.

**Test Steps**:
1. Define test cases with different unit states:
   - Active unit
   - Inactive unit
   - Failed unit
   - Non-existent unit
2. For each test case:
   a. Mock D-Bus responses to simulate the unit state.
   b. Call UnitActive with the unit name.
   c. Verify the result matches the expected state.
3. Test error handling for D-Bus connection failures.

**Expected Result**: UnitActive correctly determines if a unit is active and handles errors appropriately.
