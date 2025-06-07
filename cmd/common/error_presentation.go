// Package common provides shared functionality for OneMount command-line applications.
package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/logging"
)

// ErrorCategory represents a category of errors for user-friendly presentation
type ErrorCategory int

const (
	// ErrorCategoryGeneral represents general errors
	ErrorCategoryGeneral ErrorCategory = iota
	// ErrorCategoryNetwork represents network-related errors
	ErrorCategoryNetwork
	// ErrorCategoryAuth represents authentication-related errors
	ErrorCategoryAuth
	// ErrorCategoryRateLimit represents rate limit errors
	ErrorCategoryRateLimit
	// ErrorCategoryNotFound represents not found errors
	ErrorCategoryNotFound
	// ErrorCategoryPermission represents permission-related errors
	ErrorCategoryPermission
	// ErrorCategoryValidation represents validation errors
	ErrorCategoryValidation
	// ErrorCategoryOperation represents operation errors
	ErrorCategoryOperation
)

// UserFriendlyError represents an error with user-friendly presentation
type UserFriendlyError struct {
	// Original is the original error
	Original error
	// Category is the category of the error
	Category ErrorCategory
	// Title is a short, user-friendly title for the error
	Title string
	// Message is a user-friendly message explaining the error
	Message string
	// Suggestion is a suggestion for how to resolve the error
	Suggestion string
}

// CategorizeError categorizes an error for user-friendly presentation
func CategorizeError(err error) UserFriendlyError {
	if err == nil {
		return UserFriendlyError{
			Original:   nil,
			Category:   ErrorCategoryGeneral,
			Title:      "Unknown Error",
			Message:    "An unknown error occurred.",
			Suggestion: "Please try again later or contact support if the problem persists.",
		}
	}

	// Default values
	result := UserFriendlyError{
		Original:   err,
		Category:   ErrorCategoryGeneral,
		Title:      "Error",
		Message:    err.Error(),
		Suggestion: "Please try again later or contact support if the problem persists.",
	}

	// Check for specific error types
	switch {
	case errors.IsNetworkError(err):
		result.Category = ErrorCategoryNetwork
		result.Title = "Network Error"
		result.Message = "A network error occurred while communicating with OneDrive."
		result.Suggestion = "Please check your internet connection and try again. If the problem persists, it might be a temporary issue with the OneDrive service."

	case errors.IsAuthError(err):
		result.Category = ErrorCategoryAuth
		result.Title = "Authentication Error"
		result.Message = "There was a problem with your OneDrive authentication."
		result.Suggestion = "Please try re-authenticating with the '--auth-only' flag. If the problem persists, you may need to check your OneDrive account settings."

	case errors.IsResourceBusyError(err):
		result.Category = ErrorCategoryRateLimit
		result.Title = "Rate Limit Exceeded"
		result.Message = "OneDrive has temporarily limited access due to too many requests."
		result.Suggestion = "The system will automatically retry your request. For heavy usage, consider spacing out your operations or using the filesystem during off-peak hours."

	case errors.IsNotFoundError(err):
		result.Category = ErrorCategoryNotFound
		result.Title = "Resource Not Found"
		result.Message = "The requested file or folder could not be found in your OneDrive."
		result.Suggestion = "Please check if the file or folder exists in your OneDrive account. It may have been moved, renamed, or deleted."

	case errors.IsValidationError(err):
		result.Category = ErrorCategoryValidation
		result.Title = "Validation Error"
		result.Message = "The operation could not be completed due to invalid input."
		result.Suggestion = "Please check your input and try again. Make sure file names don't contain invalid characters and paths are correct."

	case errors.IsOperationError(err):
		result.Category = ErrorCategoryOperation
		result.Title = "Operation Failed"
		result.Message = "The requested operation could not be completed."
		result.Suggestion = "This might be a temporary issue. Please try again later. If the problem persists, check the logs for more details."
	}

	// Extract more specific information from the error message if available
	errStr := err.Error()
	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "TLS") || strings.Contains(errStr, "SSL") {
		result.Category = ErrorCategoryNetwork
		result.Title = "SSL/TLS Error"
		result.Message = "There was a problem with the secure connection to OneDrive."
		result.Suggestion = "This might be due to network security settings or a proxy. Check your network configuration and security software."
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "timed out") {
		result.Category = ErrorCategoryNetwork
		result.Title = "Connection Timeout"
		result.Message = "The connection to OneDrive timed out."
		result.Suggestion = "This might be due to slow internet or OneDrive being temporarily unavailable. Please try again later."
	}

	if strings.Contains(errStr, "quota") || strings.Contains(errStr, "storage") {
		result.Category = ErrorCategoryOperation
		result.Title = "Storage Quota Exceeded"
		result.Message = "You have reached your OneDrive storage limit."
		result.Suggestion = "Please free up space in your OneDrive account or upgrade your storage plan."
	}

	return result
}

// PrintUserFriendlyError prints a user-friendly error message to stderr
func PrintUserFriendlyError(err error) {
	if err == nil {
		return
	}

	// Record the error for monitoring
	errors.MonitorError(err)

	// Log the original error with full details
	logging.Error().Err(err).Msg("Error occurred")

	// Categorize the error for user-friendly presentation
	friendly := CategorizeError(err)

	// Print a user-friendly error message to stderr
	fmt.Fprintf(os.Stderr, "\n%s: %s\n\n", friendly.Title, friendly.Message)
	fmt.Fprintf(os.Stderr, "Suggestion: %s\n\n", friendly.Suggestion)

	// For debugging, print the original error message
	if os.Getenv("ONEMOUNT_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "Technical details: %s\n\n", err.Error())
	}
}

// HandleErrorAndExit prints a user-friendly error message and exits with the given exit code
func HandleErrorAndExit(err error, exitCode int) {
	PrintUserFriendlyError(err)
	os.Exit(exitCode)
}
