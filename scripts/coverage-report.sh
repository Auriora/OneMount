#!/bin/bash

# Coverage Report Generator for OneMount
# This script generates comprehensive coverage reports with detailed analysis

set -e

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="$COVERAGE_DIR/coverage.out"
HTML_REPORT="$COVERAGE_DIR/coverage.html"
DETAILED_REPORT="$COVERAGE_DIR/coverage-detailed.html"
JSON_REPORT="$COVERAGE_DIR/coverage.json"
HISTORY_FILE="$COVERAGE_DIR/coverage_history.json"
THRESHOLD_LINE=80
THRESHOLD_FUNC=90
THRESHOLD_BRANCH=70
CI_MODE=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --ci)
            CI_MODE=true
            shift
            ;;
        --threshold-line)
            THRESHOLD_LINE="$2"
            shift 2
            ;;
        --threshold-func)
            THRESHOLD_FUNC="$2"
            shift 2
            ;;
        --threshold-branch)
            THRESHOLD_BRANCH="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [--ci] [--threshold-line N] [--threshold-func N] [--threshold-branch N]"
            echo "  --ci                 Enable CI mode (machine-readable output)"
            echo "  --threshold-line N   Set line coverage threshold (default: 80)"
            echo "  --threshold-func N   Set function coverage threshold (default: 90)"
            echo "  --threshold-branch N Set branch coverage threshold (default: 70)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Colors for output (disabled in CI mode)
if [ "$CI_MODE" = false ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if coverage file exists
if [ ! -f "$COVERAGE_FILE" ]; then
    log_error "Coverage file not found: $COVERAGE_FILE"
    log_info "Run 'make coverage' first to generate coverage data"
    exit 1
fi

# Create coverage directory if it doesn't exist
mkdir -p "$COVERAGE_DIR"

log_info "Generating coverage reports..."

# Generate standard HTML report
log_info "Generating HTML coverage report..."
go tool cover -html="$COVERAGE_FILE" -o "$HTML_REPORT"
log_success "HTML report generated: $HTML_REPORT"

# Generate function coverage report
log_info "Generating function coverage analysis..."
go tool cover -func="$COVERAGE_FILE" > "$COVERAGE_DIR/coverage-func.txt"

# Extract total coverage
TOTAL_COVERAGE=$(go tool cover -func="$COVERAGE_FILE" | grep total | awk '{print $3}' | sed 's/%//')
log_info "Total coverage: ${TOTAL_COVERAGE}%"

# Generate detailed package-by-package analysis
log_info "Generating package analysis..."
cat > "$COVERAGE_DIR/package-analysis.txt" << EOF
OneMount Coverage Analysis Report
Generated: $(date)
Total Coverage: ${TOTAL_COVERAGE}%

Package Coverage Details:
========================
EOF

# Process each package
go tool cover -func="$COVERAGE_FILE" | grep -v total | while read -r line; do
    if [[ $line == *".go:"* ]]; then
        package=$(echo "$line" | awk '{print $1}' | sed 's/\/[^\/]*\.go:.*$//')
        func=$(echo "$line" | awk '{print $2}')
        coverage=$(echo "$line" | awk '{print $3}' | sed 's/%//')
        
        echo "$package $func $coverage" >> "$COVERAGE_DIR/temp-package-data.txt"
    fi
done

# Aggregate by package
if [ -f "$COVERAGE_DIR/temp-package-data.txt" ]; then
    awk '{
        package = $1
        coverage = $3
        count[package]++
        total[package] += coverage
    }
    END {
        for (pkg in count) {
            avg = total[pkg] / count[pkg]
            printf "%-50s %6.1f%%\n", pkg, avg
        }
    }' "$COVERAGE_DIR/temp-package-data.txt" | sort >> "$COVERAGE_DIR/package-analysis.txt"
    
    rm -f "$COVERAGE_DIR/temp-package-data.txt"
fi

# Generate JSON report for programmatic access
log_info "Generating JSON coverage report..."
cat > "$JSON_REPORT" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "total_coverage": $TOTAL_COVERAGE,
  "thresholds": {
    "line": $THRESHOLD_LINE,
    "function": $THRESHOLD_FUNC,
    "branch": $THRESHOLD_BRANCH
  },
  "packages": [
EOF

# Add package data to JSON (simplified for now)
echo "  ]," >> "$JSON_REPORT"
echo "  \"files\": []" >> "$JSON_REPORT"
echo "}" >> "$JSON_REPORT"

# Update coverage history
log_info "Updating coverage history..."
TIMESTAMP=$(date +%s)

if [ ! -f "$HISTORY_FILE" ]; then
    echo "[]" > "$HISTORY_FILE"
fi

# Add current coverage to history
python3 -c "
import json
import sys

try:
    with open('$HISTORY_FILE', 'r') as f:
        history = json.load(f)
except:
    history = []

# Add new entry
new_entry = {
    'timestamp': $TIMESTAMP,
    'total_coverage': $TOTAL_COVERAGE,
    'date': '$(date -Iseconds)'
}

history.append(new_entry)

# Keep only last 100 entries
history = history[-100:]

with open('$HISTORY_FILE', 'w') as f:
    json.dump(history, f, indent=2)

print('Coverage history updated')
" 2>/dev/null || log_warning "Could not update coverage history (Python not available)"

# Check thresholds
log_info "Checking coverage thresholds..."
THRESHOLD_PASSED=true

if (( $(echo "$TOTAL_COVERAGE < $THRESHOLD_LINE" | bc -l) )); then
    log_error "Line coverage ${TOTAL_COVERAGE}% is below threshold of ${THRESHOLD_LINE}%"
    THRESHOLD_PASSED=false
else
    log_success "Line coverage ${TOTAL_COVERAGE}% meets threshold of ${THRESHOLD_LINE}%"
fi

# Generate coverage gaps report
log_info "Analyzing coverage gaps..."
cat > "$COVERAGE_DIR/coverage-gaps.txt" << EOF
Coverage Gaps Analysis
=====================
Generated: $(date)

Files with coverage below ${THRESHOLD_LINE}%:
EOF

go tool cover -func="$COVERAGE_FILE" | awk -v threshold="$THRESHOLD_LINE" '
$3 != "total:" && $3 != "" {
    coverage = $3
    gsub(/%/, "", coverage)
    if (coverage < threshold) {
        printf "%-60s %s\n", $1, $3
    }
}' >> "$COVERAGE_DIR/coverage-gaps.txt"

# Generate summary report
log_info "Generating summary report..."
cat > "$COVERAGE_DIR/summary.txt" << EOF
OneMount Coverage Summary
========================
Generated: $(date)
Total Coverage: ${TOTAL_COVERAGE}%

Thresholds:
- Line Coverage: ${THRESHOLD_LINE}% $([ "$(echo "$TOTAL_COVERAGE >= $THRESHOLD_LINE" | bc -l)" = "1" ] && echo "✅ PASS" || echo "❌ FAIL")
- Function Coverage: ${THRESHOLD_FUNC}% (target)
- Branch Coverage: ${THRESHOLD_BRANCH}% (target)

Reports Generated:
- HTML Report: $HTML_REPORT
- Function Analysis: $COVERAGE_DIR/coverage-func.txt
- Package Analysis: $COVERAGE_DIR/package-analysis.txt
- Coverage Gaps: $COVERAGE_DIR/coverage-gaps.txt
- JSON Report: $JSON_REPORT
- Coverage History: $HISTORY_FILE

EOF

# Display summary
if [ "$CI_MODE" = false ]; then
    echo
    log_info "Coverage Report Summary:"
    cat "$COVERAGE_DIR/summary.txt"
    echo
fi

# Generate Cobertura XML for CI systems
if command -v gocov >/dev/null 2>&1 && command -v gocov-xml >/dev/null 2>&1; then
    log_info "Generating Cobertura XML report..."
    gocov convert "$COVERAGE_FILE" | gocov-xml > "$COVERAGE_DIR/cobertura.xml"
    log_success "Cobertura XML report generated: $COVERAGE_DIR/cobertura.xml"
fi

log_success "Coverage analysis complete!"

# Exit with error if thresholds not met
if [ "$THRESHOLD_PASSED" = false ]; then
    exit 1
fi
