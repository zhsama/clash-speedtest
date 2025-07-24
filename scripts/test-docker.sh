#!/bin/bash
# Docker Test Automation Script
# This script runs a comprehensive test suite for the Docker setup

set -euo pipefail

# Color codes for output
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly RED='\033[0;31m'
readonly NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Test results array
declare -A TEST_RESULTS

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test execution function
run_test() {
    local test_name="$1"
    local test_command="$2"
    local test_category="${3:-General}"
    
    ((TOTAL_TESTS++))
    echo -n "[$test_category] Testing: $test_name... "
    
    if eval "$test_command" > /tmp/test_output_$$.log 2>&1; then
        echo -e "${GREEN}✅ PASSED${NC}"
        ((PASSED_TESTS++))
        TEST_RESULTS["$test_category::$test_name"]="PASSED"
    else
        echo -e "${RED}❌ FAILED${NC}"
        ((FAILED_TESTS++))
        TEST_RESULTS["$test_category::$test_name"]="FAILED"
        if [[ -f /tmp/test_output_$$.log ]]; then
            echo "Error output:"
            tail -n 5 /tmp/test_output_$$.log
        fi
    fi
    
    rm -f /tmp/test_output_$$.log
}

# Skip test function (for optional tests)
skip_test() {
    local test_name="$1"
    local reason="$2"
    local test_category="${3:-General}"
    
    ((TOTAL_TESTS++))
    ((SKIPPED_TESTS++))
    echo -e "[$test_category] Testing: $test_name... ${YELLOW}⚠️  SKIPPED${NC} ($reason)"
    TEST_RESULTS["$test_category::$test_name"]="SKIPPED: $reason"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    docker rm -f $(docker ps -aq --filter "label=clash-speedtest-test") 2>/dev/null || true
    docker rmi $(docker images -q --filter "label=clash-speedtest-test") 2>/dev/null || true
    rm -f test-config.yaml test-hot-reload.sh run-all-tests.sh
    rm -rf test-configs output
}

# Trap cleanup on exit
trap cleanup EXIT

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local prereqs_met=true
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        prereqs_met=false
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_warn "Docker Compose is not installed (some tests will be skipped)"
    fi
    
    # Check if we're in the project root
    if [[ ! -d "backend" ]] || [[ ! -d "frontend" ]]; then
        log_error "Please run this script from the project root directory"
        prereqs_met=false
    fi
    
    if [[ "$prereqs_met" == "false" ]]; then
        exit 1
    fi
}

# Test Categories
test_build_validation() {
    log_info "Running Build Validation Tests..."
    
    # Backend build tests
    run_test "Backend clean build" \
        "docker build --label clash-speedtest-test -t clash-speedtest-backend-test:latest ./backend" \
        "Build"
    
    run_test "Backend build cache effectiveness" \
        "docker build --label clash-speedtest-test -t clash-speedtest-backend-test:cache ./backend && \
         [[ $(docker images -q clash-speedtest-backend-test:cache) ]]" \
        "Build"
    
    # Frontend build tests
    run_test "Frontend clean build" \
        "docker build --label clash-speedtest-test -t clash-speedtest-frontend-test:latest ./frontend" \
        "Build"
    
    run_test "Frontend production build" \
        "docker build --label clash-speedtest-test --target production -t clash-speedtest-frontend-test:prod ./frontend" \
        "Build"
    
    run_test "Frontend development build" \
        "docker build --label clash-speedtest-test --target development -t clash-speedtest-frontend-test:dev ./frontend" \
        "Build"
}

test_runtime_functionality() {
    log_info "Running Runtime Functionality Tests..."
    
    # Backend runtime tests
    run_test "Backend help command" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest --help" \
        "Runtime"
    
    run_test "Backend version command" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest --version" \
        "Runtime"
    
    # Test config file mounting
    echo "proxies: []" > test-config.yaml
    run_test "Backend config file mounting" \
        "docker run --rm --label clash-speedtest-test -v $(pwd)/test-config.yaml:/config/config.yaml \
         clash-speedtest-backend-test:latest -c /config/config.yaml" \
        "Runtime"
    
    # Frontend runtime tests
    run_test "Frontend nginx configuration" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-frontend-test:prod nginx -t" \
        "Runtime"
}

test_security() {
    log_info "Running Security Tests..."
    
    # User permission tests
    run_test "Backend runs as non-root user" \
        "[[ $(docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest whoami) == 'appuser' ]]" \
        "Security"
    
    run_test "Backend cannot write to /app" \
        "! docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest touch /app/test" \
        "Security"
    
    run_test "No sensitive environment variables" \
        "! docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest \
         sh -c \"env | grep -E '(PASSWORD|SECRET|KEY|TOKEN)'\"" \
        "Security"
    
    # Capability tests
    run_test "Backend works without capabilities" \
        "docker run --rm --label clash-speedtest-test --cap-drop=ALL \
         clash-speedtest-backend-test:latest --version" \
        "Security"
    
    # Read-only filesystem test
    run_test "Backend works with read-only filesystem" \
        "docker run --rm --label clash-speedtest-test --read-only \
         clash-speedtest-backend-test:latest --version" \
        "Security"
}

test_performance() {
    log_info "Running Performance Tests..."
    
    # Image size tests
    run_test "Backend image size < 100MB" \
        "[[ $(docker image inspect clash-speedtest-backend-test:latest --format='{{.Size}}') -lt 104857600 ]]" \
        "Performance"
    
    run_test "Frontend production image size < 150MB" \
        "[[ $(docker image inspect clash-speedtest-frontend-test:prod --format='{{.Size}}') -lt 157286400 ]]" \
        "Performance"
    
    # Startup time tests
    run_test "Backend startup time < 2 seconds" \
        "timeout 2 docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest --version" \
        "Performance"
}

test_networking() {
    log_info "Running Networking Tests..."
    
    # DNS resolution test
    run_test "DNS resolution works" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest \
         sh -c 'nslookup google.com'" \
        "Network"
    
    # External connectivity test
    run_test "HTTPS connectivity works" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest \
         sh -c 'wget -O- --timeout=5 https://api.github.com/rate_limit | grep -q rate'" \
        "Network"
    
    # Docker Compose tests (if available)
    if command -v docker-compose &> /dev/null; then
        run_test "Docker Compose configuration valid" \
            "docker-compose config" \
            "Network"
    else
        skip_test "Docker Compose configuration" "docker-compose not installed" "Network"
    fi
}

test_development_workflow() {
    log_info "Running Development Workflow Tests..."
    
    # Volume mounting tests
    mkdir -p output
    run_test "Output directory mounting" \
        "docker run --rm --label clash-speedtest-test -v $(pwd)/output:/output \
         clash-speedtest-backend-test:latest sh -c 'touch /output/test && ls /output/test'" \
        "Development"
    
    # Development tools availability
    run_test "Debug tools available in backend" \
        "docker run --rm --label clash-speedtest-test clash-speedtest-backend-test:latest \
         sh -c 'which wget && which curl'" \
        "Development"
}

# Optional tests (require additional tools)
test_optional_security() {
    log_info "Running Optional Security Tests..."
    
    # Trivy vulnerability scanning
    if command -v trivy &> /dev/null; then
        run_test "Backend vulnerability scan" \
            "trivy image --severity HIGH,CRITICAL --exit-code 0 clash-speedtest-backend-test:latest" \
            "Security-Scan"
            
        run_test "Frontend vulnerability scan" \
            "trivy image --severity HIGH,CRITICAL --exit-code 0 clash-speedtest-frontend-test:prod" \
            "Security-Scan"
    else
        skip_test "Backend vulnerability scan" "trivy not installed" "Security-Scan"
        skip_test "Frontend vulnerability scan" "trivy not installed" "Security-Scan"
    fi
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="docker-test-report-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "Docker Test Report"
        echo "=================="
        echo "Date: $(date)"
        echo ""
        echo "Summary:"
        echo "--------"
        echo "Total Tests: $TOTAL_TESTS"
        echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
        echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
        echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"
        echo ""
        echo "Success Rate: $(( PASSED_TESTS * 100 / (TOTAL_TESTS - SKIPPED_TESTS) ))%"
        echo ""
        echo "Detailed Results:"
        echo "-----------------"
        
        for test in "${!TEST_RESULTS[@]}"; do
            IFS='::' read -r category name <<< "$test"
            local result="${TEST_RESULTS[$test]}"
            local status_color=""
            
            case "$result" in
                "PASSED") status_color="${GREEN}" ;;
                "FAILED") status_color="${RED}" ;;
                SKIPPED*) status_color="${YELLOW}" ;;
            esac
            
            printf "%-15s %-40s %s\n" "[$category]" "$name" "${status_color}${result}${NC}"
        done | sort
    } | tee "$report_file"
    
    log_info "Report saved to: $report_file"
}

# Main execution
main() {
    echo "🧪 Docker Setup Validation Test Suite"
    echo "===================================="
    echo ""
    
    check_prerequisites
    
    # Run test categories
    test_build_validation
    test_runtime_functionality
    test_security
    test_performance
    test_networking
    test_development_workflow
    test_optional_security
    
    # Generate report
    echo ""
    generate_report
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -gt 0 ]]; then
        log_error "Some tests failed. Please review the report."
        exit 1
    else
        log_info "All tests passed successfully!"
        exit 0
    fi
}

# Run main function
main "$@"