# Security Policy

## Supported Versions

The following versions of onedriver are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

The onedriver team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Vulnerability

If you believe you've found a security vulnerability in onedriver, please follow these steps:

1. **Do not disclose the vulnerability publicly** until it has been addressed by the maintainers.
2. Email your findings to [INSERT SECURITY EMAIL]. If you don't receive a response within 48 hours, please follow up.
3. Provide detailed information about the vulnerability, including:
   - A description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Any suggestions for mitigation or remediation

### What to Expect

After you report a vulnerability:

1. The maintainers will acknowledge receipt of your report within 48 hours.
2. We will investigate the issue and determine its severity and impact.
3. We will work on a fix and keep you informed of our progress.
4. Once the issue is resolved, we will publicly acknowledge your responsible disclosure (unless you prefer to remain anonymous).

## Security Best Practices for Users

To ensure the security of your onedriver installation:

1. **Keep onedriver updated**: Always use the latest version to benefit from security patches.
2. **Protect your Microsoft account**: Use strong passwords and enable two-factor authentication for your Microsoft account.
3. **Be cautious with permissions**: Only grant onedriver access to the OneDrive folders you need.
4. **Review application logs**: Regularly check logs for any unusual activity.
5. **Report suspicious behavior**: If you notice anything unusual, report it to the maintainers.

## Security Design Principles

onedriver follows these security principles:

1. **Minimal permissions**: onedriver only requests the permissions it needs to function.
2. **Secure authentication**: We use Microsoft's OAuth 2.0 implementation for secure authentication.
3. **Local encryption**: Cached files are stored with appropriate filesystem permissions.
4. **No telemetry**: onedriver does not collect or transmit user data beyond what's needed for OneDrive operations.

## Third-Party Dependencies

onedriver relies on several third-party libraries. We regularly review and update these dependencies to address known vulnerabilities.

Key dependencies include:
- Go standard library
- FUSE (go-fuse/v2)
- GTK3 (gotk3)
- bbolt database
- zerolog

## Acknowledgments

We would like to thank the following individuals who have helped improve the security of onedriver through responsible disclosure:

- [List will be updated as contributions are received]