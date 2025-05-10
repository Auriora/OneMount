// Package testutil provides testing utilities for the OneMount project.
package framework

import (
	"github.com/auriora/onemount/pkg/testutil"
	"testing"
)
import (
	"context"
	"github.com/stretchr/testify/assert"
)

// TestSecurityTestFramework tests the basic functionality of the SecurityTestFramework.
func TestSecurityTestFramework(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("security-test")

	// Create a security test config
	securityConfig := SecurityTestConfig{
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.GetDefaultArtifactsDir(),
		CustomOptions:  map[string]interface{}{"option1": "value1"},
	}

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(securityConfig, logger)

	// Verify the framework was created correctly
	assert.NotNil(t, securityFramework)
	assert.Equal(t, securityConfig, securityFramework.Config)
	assert.NotNil(t, securityFramework.scanners)
	assert.NotNil(t, securityFramework.attackSimulators)
	assert.NotNil(t, securityFramework.controlVerifiers)
	assert.NotNil(t, securityFramework.authenticationTesters)
	assert.NotNil(t, securityFramework.authorizationTesters)
	assert.NotNil(t, securityFramework.dataProtectionTesters)
	assert.Equal(t, logger, securityFramework.logger)
}

// TestSecurityTestScenarios tests the creation of security test scenarios.
func TestSecurityTestScenarios(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test config
	securityConfig := SecurityTestConfig{
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   testutil.GetDefaultArtifactsDir(),
		CustomOptions:  map[string]interface{}{"option1": "value1"},
	}

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(securityConfig, logger)

	// Create security test scenarios
	securityScenarios := NewSecurityTestScenarios(securityFramework)

	// Verify the scenarios were created correctly
	assert.NotNil(t, securityScenarios)
	assert.Equal(t, securityFramework, securityScenarios.SecurityFramework)
	assert.Equal(t, logger, securityScenarios.logger)

	// Verify each scenario is created correctly
	vulnerabilityScanScenario := securityScenarios.VulnerabilityScanScenario()
	assert.Equal(t, "Vulnerability Scan", vulnerabilityScanScenario.Name)
	assert.Contains(t, vulnerabilityScanScenario.Tags, "security")
	assert.Contains(t, vulnerabilityScanScenario.Tags, "vulnerability")
	assert.NotEmpty(t, vulnerabilityScanScenario.Steps)
	assert.NotEmpty(t, vulnerabilityScanScenario.Assertions)
	assert.NotEmpty(t, vulnerabilityScanScenario.Cleanup)

	securityAttackScenario := securityScenarios.SecurityAttackSimulationScenario()
	assert.Equal(t, "Security Attack Simulation", securityAttackScenario.Name)
	assert.Contains(t, securityAttackScenario.Tags, "security")
	assert.Contains(t, securityAttackScenario.Tags, "attack-simulation")
	assert.NotEmpty(t, securityAttackScenario.Steps)
	assert.NotEmpty(t, securityAttackScenario.Assertions)
	assert.NotEmpty(t, securityAttackScenario.Cleanup)

	securityControlScenario := securityScenarios.SecurityControlVerificationScenario()
	assert.Equal(t, "Security Control Verification", securityControlScenario.Name)
	assert.Contains(t, securityControlScenario.Tags, "security")
	assert.Contains(t, securityControlScenario.Tags, "controls")
	assert.NotEmpty(t, securityControlScenario.Steps)
	assert.NotEmpty(t, securityControlScenario.Assertions)
	assert.NotEmpty(t, securityControlScenario.Cleanup)

	authenticationScenario := securityScenarios.AuthenticationTestScenario()
	assert.Equal(t, "Authentication Testing", authenticationScenario.Name)
	assert.Contains(t, authenticationScenario.Tags, "security")
	assert.Contains(t, authenticationScenario.Tags, "authentication")
	assert.NotEmpty(t, authenticationScenario.Steps)
	assert.NotEmpty(t, authenticationScenario.Assertions)
	assert.NotEmpty(t, authenticationScenario.Cleanup)

	authorizationScenario := securityScenarios.AuthorizationTestScenario()
	assert.Equal(t, "Authorization Testing", authorizationScenario.Name)
	assert.Contains(t, authorizationScenario.Tags, "security")
	assert.Contains(t, authorizationScenario.Tags, "authorization")
	assert.NotEmpty(t, authorizationScenario.Steps)
	assert.NotEmpty(t, authorizationScenario.Assertions)
	assert.NotEmpty(t, authorizationScenario.Cleanup)

	dataProtectionScenario := securityScenarios.DataProtectionTestScenario()
	assert.Equal(t, "Data Protection Testing", dataProtectionScenario.Name)
	assert.Contains(t, dataProtectionScenario.Tags, "security")
	assert.Contains(t, dataProtectionScenario.Tags, "data-protection")
	assert.NotEmpty(t, dataProtectionScenario.Steps)
	assert.NotEmpty(t, dataProtectionScenario.Assertions)
	assert.NotEmpty(t, dataProtectionScenario.Cleanup)

	// Verify GetAllSecurityTestScenarios returns all scenarios
	allScenarios := securityScenarios.GetAllSecurityTestScenarios()
	assert.Equal(t, 6, len(allScenarios))
}

// TestSecurityScannerRegistration tests the registration and retrieval of security scanners.
func TestSecurityScannerRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create a security scanner
	scanner := NewBasicSecurityScanner("test-scanner", nil, logger)

	// Register the scanner
	securityFramework.RegisterScanner("test-scanner", scanner)

	// Verify the scanner was registered correctly
	registeredScanner, exists := securityFramework.GetScanner("test-scanner")
	assert.True(t, exists)
	assert.Equal(t, scanner, registeredScanner)

	// Verify a non-existent scanner returns false
	_, exists = securityFramework.GetScanner("non-existent-scanner")
	assert.False(t, exists)
}

// TestSecurityAttackSimulatorRegistration tests the registration and retrieval of security attack simulators.
func TestSecurityAttackSimulatorRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create a security attack simulator
	simulator := NewBasicAttackSimulator("test-simulator", nil, logger)

	// Register the simulator
	securityFramework.RegisterAttackSimulator("test-simulator", simulator)

	// Verify the simulator was registered correctly
	registeredSimulator, exists := securityFramework.GetAttackSimulator("test-simulator")
	assert.True(t, exists)
	assert.Equal(t, simulator, registeredSimulator)

	// Verify a non-existent simulator returns false
	_, exists = securityFramework.GetAttackSimulator("non-existent-simulator")
	assert.False(t, exists)
}

// TestSecurityControlVerifierRegistration tests the registration and retrieval of security control verifiers.
func TestSecurityControlVerifierRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create a security control verifier
	verifier := NewBasicControlVerifier("test-verifier", nil, logger)

	// Register the verifier
	securityFramework.RegisterControlVerifier("test-verifier", verifier)

	// Verify the verifier was registered correctly
	registeredVerifier, exists := securityFramework.GetControlVerifier("test-verifier")
	assert.True(t, exists)
	assert.Equal(t, verifier, registeredVerifier)

	// Verify a non-existent verifier returns false
	_, exists = securityFramework.GetControlVerifier("non-existent-verifier")
	assert.False(t, exists)
}

// TestAuthenticationTesterRegistration tests the registration and retrieval of authentication testers.
func TestAuthenticationTesterRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create an authentication tester
	tester := NewBasicAuthenticationTester("test-tester", nil, logger)

	// Register the tester
	securityFramework.RegisterAuthenticationTester("test-tester", tester)

	// Verify the tester was registered correctly
	registeredTester, exists := securityFramework.GetAuthenticationTester("test-tester")
	assert.True(t, exists)
	assert.Equal(t, tester, registeredTester)

	// Verify a non-existent tester returns false
	_, exists = securityFramework.GetAuthenticationTester("non-existent-tester")
	assert.False(t, exists)
}

// TestAuthorizationTesterRegistration tests the registration and retrieval of authorization testers.
func TestAuthorizationTesterRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create an authorization tester
	tester := NewBasicAuthorizationTester("test-tester", nil, logger)

	// Register the tester
	securityFramework.RegisterAuthorizationTester("test-tester", tester)

	// Verify the tester was registered correctly
	registeredTester, exists := securityFramework.GetAuthorizationTester("test-tester")
	assert.True(t, exists)
	assert.Equal(t, tester, registeredTester)

	// Verify a non-existent tester returns false
	_, exists = securityFramework.GetAuthorizationTester("non-existent-tester")
	assert.False(t, exists)
}

// TestDataProtectionTesterRegistration tests the registration and retrieval of data protection testers.
func TestDataProtectionTesterRegistration(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create a data protection tester
	tester := NewBasicDataProtectionTester("test-tester", nil, logger)

	// Register the tester
	securityFramework.RegisterDataProtectionTester("test-tester", tester)

	// Verify the tester was registered correctly
	registeredTester, exists := securityFramework.GetDataProtectionTester("test-tester")
	assert.True(t, exists)
	assert.Equal(t, tester, registeredTester)

	// Verify a non-existent tester returns false
	_, exists = securityFramework.GetDataProtectionTester("non-existent-tester")
	assert.False(t, exists)
}

// TestRunSecurityScan tests the RunSecurityScan method.
func TestRunSecurityScan(t *testing.T) {
	// Create a logger
	logger := testutil.NewZerologLogger("")

	// Create a security test framework
	securityFramework := NewSecurityTestFramework(SecurityTestConfig{}, logger)

	// Create a security scanner
	scanner := NewBasicSecurityScanner("test-scanner", nil, logger)

	// Register the scanner
	securityFramework.RegisterScanner("test-scanner", scanner)

	// Run a security scan
	result, err := securityFramework.RunSecurityScan(context.Background(), "test-scanner", "test-target", nil)

	// Verify the scan was successful
	assert.NoError(t, err)
	assert.Equal(t, "test-target", result.Target)
	assert.NotEmpty(t, result.Timestamp)
	assert.NotEmpty(t, result.Duration)
	assert.NotEmpty(t, result.Vulnerabilities)
	assert.NotEmpty(t, result.RawOutput)

	// Verify a scan with a non-existent scanner returns an error
	_, err = securityFramework.RunSecurityScan(context.Background(), "non-existent-scanner", "test-target", nil)
	assert.Error(t, err)
}
