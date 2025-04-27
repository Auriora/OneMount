# Onedriver Code Complexity Analyzer

This tool performs static code analysis on the Onedriver codebase to calculate the cyclomatic complexity of all functions and methods.

## What is Cyclomatic Complexity?

Cyclomatic complexity is a software metric used to measure the complexity of a program. It directly measures the number of linearly independent paths through a program's source code. Higher complexity indicates code that may be more difficult to understand, test, and maintain.

The complexity is calculated by counting the number of decision points in a function plus one. Decision points include:
- if statements
- for and while loops
- case statements in switch blocks
- logical operators (&&, ||)

## Usage

To run the analyzer:

```bash
cd utils
go run complexity_analyzer.go
```

This will:
1. Traverse all Go files in the Onedriver project
2. Calculate the cyclomatic complexity for each function and method
3. Output the results to a CSV file named `complexity_analysis.csv`

## Output Format

The output CSV file contains the following columns:

- **Name**: The name of the function or method (methods are prefixed with their receiver type)
- **Type**: Either "function" or "method"
- **FilePath**: The path to the file containing the function or method
- **Complexity**: The cyclomatic complexity value

## Interpreting Results

Cyclomatic complexity values can be interpreted as follows:

- 1-10: Simple code, low risk
- 11-20: Moderately complex, moderate risk
- 21-50: Complex, high risk
- 50+: Untestable, very high risk

Functions or methods with high complexity values are candidates for refactoring to improve maintainability and testability.

## Example

To find the most complex functions in the codebase, you can sort the CSV file by complexity:

```bash
# Using the sort command (requires the CSV file to have been generated)
sort -t, -k4 -nr complexity_analysis.csv | head -n 10
```

This will show the 10 most complex functions or methods in the codebase.