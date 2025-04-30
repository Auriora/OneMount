# GoLand Run Configurations for OneMount

This directory contains run configurations for JetBrains GoLand that replicate the functionality of the `make test` command.

## Available Run Configurations

1. **UI Tests** - Runs tests in the UI package, excluding offline tests
   - Command equivalent: `gotest -v -parallel=8 -count=1 $(shell go list ./ui/... | grep -v offline)`

2. **Command Tests** - Runs tests in the cmd package
   - Command equivalent: `gotest -v -parallel=8 -count=1 ./cmd/...`

3. **Graph Tests with Race Detection** - Runs tests in the fs/graph package with race detection
   - Command equivalent: `gotest -race -v -parallel=8 -count=1 ./fs/graph/...`

4. **FS Tests with Race Detection** - Runs tests in the fs package with race detection
   - Command equivalent: `gotest -race -v -parallel=8 -count=1 ./fs`

5. **Offline Tests** - Builds the offline test binary and provides instructions for running it
   - Note: This requires sudo privileges and cannot be run directly from GoLand
   - Command equivalent: 
     ```
     go test -c ./fs/offline
     sudo unshare -n sudo -u $(whoami) ./offline.test -test.v -test.parallel=8 -test.count=1
     ```

6. **All Tests Except Offline** - Runs all the above tests except for Offline Tests

## Usage

1. Open the project in GoLand
2. Go to the Run/Debug Configurations dropdown in the toolbar
3. Select the desired configuration and click the Run button

For Offline Tests, you'll need to run the command manually in a terminal after building the test binary.