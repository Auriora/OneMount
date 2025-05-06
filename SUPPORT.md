# Support for OneMount

This document outlines how to get help with OneMount. Please read through the available support options below.

## Table of Contents

1. [Documentation](#documentation)
2. [Troubleshooting](#troubleshooting)
3. [Community Support](#community-support)
4. [Reporting Bugs](#reporting-bugs)
5. [Feature Requests](#feature-requests)
6. [Security Issues](#security-issues)

## Documentation

Before seeking support, please check if your question is answered in our documentation:

- [README](README.md) - Overview of the project and basic usage
- [Quickstart Guide](docs/quickstart-guide.md) - Step-by-step guide to get started quickly
- [Installation Guide](docs/installation-guide.md) - Detailed installation and configuration instructions
- [Development Guidelines](docs/DEVELOPMENT.md) - Information for developers

## Troubleshooting

If you're experiencing issues with OneMount, try these troubleshooting steps:

1. **Check the logs**: View the OneMount logs to identify the issue:
   ```bash
   journalctl --user -u onemount@.service --since today
   ```

2. **Restart the service**: Sometimes restarting OneMount can resolve issues:
   ```bash
   systemctl --user restart onemount@.service
   ```

3. **Unmount and remount**: If OneMount becomes unresponsive:
   ```bash
   fusermount3 -uz $MOUNTPOINT
   onemount $MOUNTPOINT
   ```

4. **Enable debug mode**: For more detailed logs:
   ```bash
   ONEMOUNT_DEBUG=1 onemount $MOUNTPOINT
   ```

5. **Check your OneDrive account**: Verify that your OneDrive account is working correctly by accessing it through the web interface.

## Common Issues

### Authentication Problems

If you're having trouble authenticating:
1. Try re-authenticating by removing the credentials and starting over
2. Check if your Microsoft account has two-factor authentication enabled
3. Ensure your system time is correct

### Performance Issues

If OneMount is running slowly:
1. Check your internet connection
2. Consider using the `--cache-size` option to increase the cache size
3. Verify that you're not trying to access very large files (multi-GB files may cause performance issues)

## Community Support

Get help from the community and project maintainers:

- **GitHub Issues**: Browse [existing issues](https://github.com/auriora/OneMount/issues) to see if your problem has already been reported or discussed.
- **GitHub Discussions**: For general questions and discussions about OneMount.

## Reporting Bugs

If you've found a bug in OneMount:

1. Check if the bug has already been reported in the [GitHub Issues](https://github.com/auriora/OneMount/issues).
2. If not, [create a new issue](https://github.com/auriora/OneMount/issues/new) with:
   - A clear and descriptive title
   - Steps to reproduce the issue
   - Expected behavior
   - Actual behavior
   - Log output
   - Your Linux distribution and version

## Feature Requests

We welcome suggestions for new features:

1. Check if the feature has already been requested in the [GitHub Issues](https://github.com/auriora/OneMount/issues).
2. If not, [create a new issue](https://github.com/auriora/OneMount/issues/new) with:
   - A clear and descriptive title
   - A detailed description of the proposed feature
   - Any relevant use cases
   - If possible, a suggestion for how to implement the feature

## Security Issues

For security-related issues, please refer to our [Security Policy](SECURITY.md) and follow the vulnerability reporting process outlined there.

## Support Timeline

OneMount is an open-source project maintained by volunteers. While we strive to address all issues in a timely manner, response times may vary based on:

- The severity of the issue
- The complexity of the problem
- The availability of maintainers

We prioritize security issues and critical bugs that affect a large number of users.

Thank you for using OneMount!
