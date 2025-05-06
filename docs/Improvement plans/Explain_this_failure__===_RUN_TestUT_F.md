
# Analysis of Test Failure: TestUT_FS_05_02_RepeatedUploads_OnlineMode_SuccessfulUpload

The test is failing with the error message:
```
upload session not found: Failed to wait for upload
```

This error occurs at line 152 in `upload_manager_test.go`, which corresponds to the `WaitForUpload` call.

## Root Cause Analysis

After examining the code and logs, I've identified the likely cause of the failure:

1. The test creates a file, queues it for upload with high priority, mocks the upload response, and then calls `WaitForUpload`.

2. The `WaitForUpload` method checks if the upload session exists in the `sessions` map, and if not, returns an error with the message "upload session not found".

3. Looking at the logs, we can see that the file is successfully queued for upload:
   ```
   [2025-05-06 18:40:18.000] INFO: File queued for upload 
   {
     "id": "file-id",
     "level": "info",
     "message": "File queued for upload",
     "name": "repeated_upload.txt",
     "priority": "high",
     "time": "2025-05-06T18:40:18+02:00"
   }
   ```

4. However, there's a race condition in the test:
   - The file is queued for upload via `QueueUploadWithPriority`
   - This adds the session to the appropriate queue (highPriorityQueue)
   - But the session is not immediately added to the `sessions` map
   - The session is only added to the `sessions` map when it's processed in the `uploadLoop` method
   - The test immediately calls `WaitForUpload` without waiting for the session to be processed by the `uploadLoop`

5. The `uploadLoop` runs on a ticker with a specified duration, so there's a delay between when a session is queued and when it's processed and added to the `sessions` map.

## Solution

To fix this issue, the test should wait for the session to be processed by the `uploadLoop` before calling `WaitForUpload`. There are a few ways to do this:

1. **Add a small delay**: Insert a short sleep (e.g., 100ms) between queuing the upload and calling `WaitForUpload` to give the `uploadLoop` time to process the session.

2. **Modify the UploadManager for testing**: Add a method to the `UploadManager` that waits for a session to be added to the `sessions` map before returning.

3. **Use a synchronization mechanism**: Modify the `QueueUploadWithPriority` method to return only after the session has been added to the `sessions` map.

The simplest fix would be to add a small delay after queuing the upload:

```go
// Queue the upload
_, err = fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
assert.NoError(err, "Failed to queue upload")

// Mock the upload response
mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

// Mock the content upload response
fileItemJSON, err := json.Marshal(fileItem)
assert.NoError(err, "Failed to marshal file item")
mockClient.AddMockResponse("/me/drive/items/"+fileID+"/content", fileItemJSON, 200, nil)

// Wait for the upload to be processed by the uploadLoop
time.Sleep(100 * time.Millisecond)

// Wait for the upload to complete
err = fs.uploads.WaitForUpload(fileID)
assert.NoError(err, "Failed to wait for upload")
```

This should give the `uploadLoop` enough time to process the session and add it to the `sessions` map before `WaitForUpload` is called.

For a more robust solution, consider adding a method to the `UploadManager` that waits for a session to be added to the `sessions` map with a timeout.