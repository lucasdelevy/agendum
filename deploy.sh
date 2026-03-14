#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()   { echo -e "${GREEN}[deploy]${NC} $*"; }
warn()  { echo -e "${YELLOW}[deploy]${NC} $*"; }
error() { echo -e "${RED}[deploy]${NC} $*" >&2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

LAMBDA_DIRS=(
    cmd/lambda-user
    cmd/lambda-auth
    cmd/lambda-task
    cmd/lambda-team
    cmd/lambda-list-teams
)
INFRA_DIR=infrastructure

check_prerequisites() {
    local missing=0
    for cmd in go aws cdk; do
        if ! command -v "$cmd" &> /dev/null; then
            error "'$cmd' is not installed or not in PATH."
            missing=1
        fi
    done

    if [ $missing -ne 0 ]; then
        error "Please install missing prerequisites and try again."
        exit 1
    fi

    if ! aws sts get-caller-identity &> /dev/null; then
        error "AWS credentials are not configured. Run 'aws configure' first."
        exit 1
    fi

    log "All prerequisites met."
}

build_lambdas() {
    log "Building Lambda functions..."
    for dir in "${LAMBDA_DIRS[@]}"; do
        log "  Building $dir..."
        (cd "$dir" && go mod tidy && GOOS=linux GOARCH=amd64 go build -o bootstrap main.go)
    done
    log "All Lambda functions built."
}

deploy_infrastructure() {
    log "Preparing infrastructure..."
    (cd "$INFRA_DIR" && go mod tidy)

    if [ "${1:-}" = "--bootstrap" ]; then
        warn "Bootstrapping CDK (first-time setup)..."
        (cd "$INFRA_DIR" && cdk bootstrap)
    fi

    log "Deploying with CDK..."
    (cd "$INFRA_DIR" && cdk deploy --require-approval never)
    log "Deployment complete!"
}

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Deploy Agendum to AWS using CDK."
    echo ""
    echo "Options:"
    echo "  --bootstrap    Run 'cdk bootstrap' before deploying (first-time setup)"
    echo "  --build-only   Build Lambda functions without deploying"
    echo "  --help         Show this help message"
}

main() {
    case "${1:-}" in
        --help|-h)
            usage
            exit 0
            ;;
        --build-only)
            check_prerequisites
            build_lambdas
            log "Build-only mode: skipping deployment."
            exit 0
            ;;
        *)
            check_prerequisites
            build_lambdas
            deploy_infrastructure "${1:-}"
            ;;
    esac
}

main "$@"
