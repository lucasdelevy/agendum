#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${GREEN}[setup]${NC} $*"; }
warn() { echo -e "${YELLOW}[setup]${NC} $*"; }
info() { echo -e "${CYAN}[setup]${NC} $*"; }
err()  { echo -e "${RED}[setup]${NC} $*" >&2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE="$SCRIPT_DIR/infrastructure/github-oidc.yml"
STACK_NAME="github-oidc-agendum"

GITHUB_ORG="lucasdelevy"
GITHUB_REPO="agendum"
DEPLOY_BRANCH="main"
EXISTING_OIDC_ARN=""
AWS_REGION="${AWS_REGION:-us-east-1}"

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Deploy the GitHub Actions OIDC provider and IAM role to your AWS account."
    echo ""
    echo "Options:"
    echo "  --region REGION           AWS region (default: us-east-1)"
    echo "  --org ORG                 GitHub org/username (default: lucasdelevy)"
    echo "  --repo REPO               GitHub repo name (default: agendum)"
    echo "  --branch BRANCH           Branch allowed to deploy (default: main)"
    echo "  --oidc-arn ARN            Existing OIDC provider ARN (if already created)"
    echo "  --delete                  Delete the stack"
    echo "  --help                    Show this help message"
}

check_prerequisites() {
    if ! command -v aws &> /dev/null; then
        err "AWS CLI is not installed. See: https://aws.amazon.com/cli/"
        exit 1
    fi

    if ! aws sts get-caller-identity &> /dev/null; then
        err "AWS credentials not configured. Run 'aws configure' first."
        exit 1
    fi

    local account_id
    account_id=$(aws sts get-caller-identity --query Account --output text)
    local identity
    identity=$(aws sts get-caller-identity --query Arn --output text)
    log "Authenticated as: $identity"
    log "AWS Account: $account_id"
    log "Region: $AWS_REGION"
}

check_existing_oidc() {
    local existing
    existing=$(aws iam list-open-id-connect-providers \
        --query "OpenIDConnectProviderList[?ends_with(Arn, '/token.actions.githubusercontent.com')].Arn" \
        --output text 2>/dev/null || echo "")

    if [ -n "$existing" ] && [ "$existing" != "None" ]; then
        warn "Found existing GitHub OIDC provider: $existing"
        warn "Reusing it (only one per account is needed)."
        EXISTING_OIDC_ARN="$existing"
    fi
}

deploy_stack() {
    log "Deploying CloudFormation stack: $STACK_NAME"

    local params=(
        "ParameterKey=GitHubOrg,ParameterValue=$GITHUB_ORG"
        "ParameterKey=GitHubRepo,ParameterValue=$GITHUB_REPO"
        "ParameterKey=DeployBranch,ParameterValue=$DEPLOY_BRANCH"
        "ParameterKey=OIDCProviderArn,ParameterValue=$EXISTING_OIDC_ARN"
    )

    aws cloudformation deploy \
        --template-file "$TEMPLATE" \
        --stack-name "$STACK_NAME" \
        --parameter-overrides "${params[@]}" \
        --capabilities CAPABILITY_NAMED_IAM \
        --region "$AWS_REGION" \
        --no-fail-on-empty-changeset

    log "Stack deployed successfully."
    echo ""

    local role_arn
    role_arn=$(aws cloudformation describe-stacks \
        --stack-name "$STACK_NAME" \
        --query "Stacks[0].Outputs[?OutputKey=='DeployRoleArn'].OutputValue" \
        --output text \
        --region "$AWS_REGION")

    echo ""
    echo "============================================================"
    info "OIDC setup complete!"
    echo ""
    info "Deploy Role ARN: ${CYAN}$role_arn${NC}"
    info "Region:          ${CYAN}$AWS_REGION${NC}"
    echo ""
    info "Next steps -- add these as GitHub repository secrets:"
    echo ""
    echo "  1. Go to: https://github.com/$GITHUB_ORG/$GITHUB_REPO/settings/secrets/actions"
    echo "  2. Add secret: AWS_ROLE_ARN = $role_arn"
    echo "  3. Add secret: AWS_REGION   = $AWS_REGION"
    echo ""
    echo "============================================================"
}

delete_stack() {
    warn "Deleting CloudFormation stack: $STACK_NAME"
    aws cloudformation delete-stack \
        --stack-name "$STACK_NAME" \
        --region "$AWS_REGION"

    log "Waiting for stack deletion..."
    aws cloudformation wait stack-delete-complete \
        --stack-name "$STACK_NAME" \
        --region "$AWS_REGION"

    log "Stack deleted."
}

main() {
    local action="deploy"

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --region)    AWS_REGION="$2"; shift 2 ;;
            --org)       GITHUB_ORG="$2"; shift 2 ;;
            --repo)      GITHUB_REPO="$2"; shift 2 ;;
            --branch)    DEPLOY_BRANCH="$2"; shift 2 ;;
            --oidc-arn)  EXISTING_OIDC_ARN="$2"; shift 2 ;;
            --delete)    action="delete"; shift ;;
            --help|-h)   usage; exit 0 ;;
            *)           err "Unknown option: $1"; usage; exit 1 ;;
        esac
    done

    check_prerequisites

    if [ "$action" = "delete" ]; then
        delete_stack
    else
        if [ -z "$EXISTING_OIDC_ARN" ]; then
            check_existing_oidc
        fi
        deploy_stack
    fi
}

main "$@"
