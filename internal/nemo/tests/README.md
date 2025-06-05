# OneMount Nemo Extension Test Suite

This directory contains comprehensive tests for the OneMount Nemo file manager extension.

## Overview

The test suite provides comprehensive coverage for:

- **Unit Tests**: Individual component testing with mocked dependencies
- **Integration Tests**: D-Bus communication between Go service and Python extension
- **Mock Tests**: Offline scenarios and edge cases
- **End-to-End Tests**: Complete workflow testing

## Test Structure

```
tests/
├── __init__.py                 # Python package marker
├── conftest.py                 # Pytest configuration and fixtures
├── test_nemo_extension.py      # Unit tests for the main extension
├── test_dbus_integration.py    # D-Bus integration tests
├── test_mocks.py              # Mock and offline scenario tests
└── README.md                  # This file
```

## Test Categories

### Unit Tests (`@pytest.mark.unit`)

Test individual components in isolation:

- Extension initialization
- Mount point detection
- File status retrieval
- Emblem assignment logic
- Error handling
- Signal processing

### Integration Tests (`@pytest.mark.integration`)

Test component interactions:

- D-Bus service connection
- Method calls and responses
- Signal emission and reception
- Service availability handling

### D-Bus Tests (`@pytest.mark.dbus`)

Specifically test D-Bus functionality:

- Connection establishment
- Method invocation
- Signal handling
- Service discovery
- Error recovery

### Mock Tests (`@pytest.mark.mock`)

Test with mocked dependencies:

- Offline scenarios
- Service unavailability
- Filesystem limitations
- Error conditions

## Running Tests

### Using the OneMount Development CLI

The easiest way to run tests is using the integrated OneMount development CLI:

```bash
# Run all tests
scripts/dev.py test nemo all

# Run specific test categories
scripts/dev.py test nemo unit
scripts/dev.py test nemo integration
scripts/dev.py test nemo dbus
scripts/dev.py test nemo mock

# Run with coverage
scripts/dev.py test nemo all --coverage
scripts/dev.py test nemo coverage

# Run specific test file
scripts/dev.py test nemo unit --test-file test_nemo_extension.py

# Run specific test function
scripts/dev.py test nemo unit --test-file test_nemo_extension.py --test-function test_extension_initialization_success

# Verbose output
scripts/dev.py test nemo all --verbose-pytest

# Check status and setup
scripts/dev.py test nemo status
scripts/dev.py test nemo setup

# Run complete test suite (Python + Go)
scripts/dev.py test nemo full
```

### Using Pytest Directly

You can also run tests directly with pytest (from the `internal/nemo` directory):

```bash
# Run all tests
python -m pytest tests/

# Run with markers
python -m pytest tests/ -m unit
python -m pytest tests/ -m integration
python -m pytest tests/ -m "unit and not slow"

# Run with coverage
python -m pytest tests/ --cov=../src --cov-report=html

# Run specific test
python -m pytest tests/test_nemo_extension.py::TestOneMountExtensionInitialization::test_extension_initialization_success
```

## Dependencies

Required Python packages:

- `pytest` - Testing framework
- `PyGObject` - GObject introspection bindings
- `dbus-python` - D-Bus Python bindings

Install with:

```bash
pip install -r ../requirements.txt
```

## Test Configuration

Test configuration is managed through:

- `pytest.ini` - Pytest configuration
- `conftest.py` - Fixtures and test setup
- Environment variables for test mode

## Fixtures

Common fixtures provided by `conftest.py`:

- `mock_dbus` - Mock D-Bus interface
- `temp_mount_point` - Temporary mount directory
- `mock_proc_mounts` - Mock /proc/mounts content
- `mock_file_info` - Mock Nemo FileInfo object
- `mock_file_object` - Mock Nemo File object
- `sample_file_statuses` - Status to emblem mappings
- `mock_xattr` - Mock extended attributes

## Writing New Tests

### Test Naming Convention

- Test files: `test_*.py`
- Test classes: `Test*`
- Test functions: `test_*`

### Test Markers

Use appropriate markers for your tests:

```python
@pytest.mark.unit
def test_unit_functionality():
    """Test individual component."""
    pass

@pytest.mark.integration
@pytest.mark.dbus
def test_dbus_integration():
    """Test D-Bus integration."""
    pass

@pytest.mark.mock
def test_offline_scenario():
    """Test offline behavior."""
    pass

@pytest.mark.slow
def test_long_running():
    """Test that takes time."""
    pass
```

### Using Fixtures

```python
def test_with_fixtures(mock_dbus, temp_mount_point, mock_file_info):
    """Test using common fixtures."""
    # Test implementation
    pass
```

### Mocking Guidelines

- Mock external dependencies (D-Bus, filesystem, Nemo)
- Use `patch` for temporary mocking within tests
- Use fixtures for common mock setups
- Test both success and failure scenarios

## Coverage Goals

Target coverage levels:

- **Unit Tests**: 95%+ line coverage
- **Integration Tests**: All D-Bus interactions
- **Error Handling**: All exception paths
- **Edge Cases**: Boundary conditions

## Continuous Integration

Tests are designed to run in CI environments:

- No external dependencies required
- All components mocked appropriately
- Fast execution (< 30 seconds total)
- Deterministic results

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure `src/` directory is in Python path
2. **D-Bus Errors**: Tests use mocked D-Bus, real D-Bus not required
3. **Permission Errors**: Tests use temporary directories
4. **Missing Dependencies**: Run `./run_tests.py --check-deps`

### Debug Mode

Run tests with verbose output and no capture:

```bash
python -m pytest tests/ -v -s --tb=long
```

### Test Isolation

Each test runs in isolation with:

- Fresh extension instances
- Clean mock states
- Temporary directories
- Independent D-Bus mocks
