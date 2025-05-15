#!/bin/bash

# Script to update CI files in ci-base and propagate to other branches

# Configuration
CI_INSTALL_URL="https://raw.githubusercontent.com/earcamone/gwy/assets/scripts/install-gwy.sh"
BRANCHES=("master" "develop" "feature/no-ref/simple-bogus-success-app" "feature/no-ref/gwy-alerts-trigger")
DRY_RUN=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Function to run commands (respects dry-run)
run_cmd() {
    echo "Running: $@"
    if [ "$DRY_RUN" = false ]; then
        "$@"
        local status=$?
        if [ $status -ne 0 ]; then
            echo "Error: Command failed with status $status: $@"
            exit $status
        fi
    fi
}

# Function to check if working directory is clean
check_clean() {
    if [ "$DRY_RUN" = false ]; then
        if ! git status --porcelain | grep -q .; then
            return 0
        else
            echo "Error: Working directory is not clean. Please commit or stash changes."
            exit 1
        fi
    fi
}

# Function to check for rebase conflicts
check_rebase_conflicts() {
    if [ "$DRY_RUN" = false ]; then
        if git status | grep -q "rebase in progress"; then
            echo "Error: Rebase conflicts detected. Please resolve manually."
            git rebase --abort
            exit 1
        fi
    fi
}

# Ensure repo is clean
# check_clean

# Fetch latest remote state
run_cmd git fetch origin

# Update ci-base
run_cmd git checkout ci-base
run_cmd git pull origin ci-base

# Run CI installation
if [ "$DRY_RUN" = false ]; then
    echo "Running CI installation: curl $CI_INSTALL_URL"
    curl -fsSL "$CI_INSTALL_URL" | bash || { echo "Error: CI installation failed"; exit 1; }
fi

# Check if CI files changed
if [ "$DRY_RUN" = false ]; then
    if git diff --quiet; then
        echo "No changes to CI files. Skipping commit."
	exit 0
    else
        run_cmd git add .
        run_cmd git commit -m "Updated CI files with latest release candidate files"
    fi
fi

# Push ci-base
run_cmd git push origin ci-base

# Backup branches
if [ "$DRY_RUN" = false ]; then
    timestamp=$(date +%Y%m%d_%H%M%S)
    for branch in "${BRANCHES[@]}"; do
        git checkout "$branch" && git branch "backup-$branch-$timestamp" || echo "Warning: Could not back up $branch"
    done
fi

# Update master
run_cmd git checkout master
run_cmd git pull origin master
run_cmd git rebase ci-base
check_rebase_conflicts
run_cmd git push origin master --force

# Update develop
run_cmd git checkout develop
run_cmd git pull origin develop
run_cmd git rebase master
check_rebase_conflicts
run_cmd git push origin develop --force

# Update feature branches
for feature in "${BRANCHES[@]:2}"; do
    run_cmd git checkout "$feature"
    run_cmd git pull origin "$feature"
    run_cmd git rebase develop
    check_rebase_conflicts
    run_cmd git push origin "$feature" --force
done

echo "Success: CI files updated and propagated to all branches!"
