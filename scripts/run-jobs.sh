#!/bin/bash

#
# Simple script to run manually at once all GWY workflows in its development
# repository with different options to ensure everything works as it should
#

# GWY files location
REPO="earcamone/gwy-playground"

# workflow options
OPTION_GO_VERSION="1.24.1"
ERRORS_BRANCH="feature/no-ref/gwy-alerts-trigger"
SUCCESS_BRANCH="feature/no-ref/simple-bogus-success-app"

# GWY workflow files
# List of workflow files to run
WORKFLOW_FILES=(
  "gwy-ci.yml"
  "gwy-coverage.yml"
  "gwy-dependencies.yml"
  "gwy-gofmt.yml"
  "gwy-lint.yml"
  "gwy-secrets.yml"
  "gwy-vulnerabilities.yml"
)

run_workflow() {
  local workflow_file=$1
  shift
  local params="$@"

  echo "Triggering workflow: $workflow_file with params: $params"
  echo " - gh workflow run \"$workflow_file\" --repo \"$REPO\" $params"

  gh workflow run "$workflow_file" --repo "$REPO" --ref master $params
  [ $? -eq 0 ] && echo "Successfully triggered $workflow_file" || { echo "Failed to trigger $workflow_file"; exit 1; }
}


# Run all workflows with default options and common custom ones
for workflow in "${WORKFLOW_FILES[@]}"; do
  run_workflow "$workflow" -f branch=invalid   # Invalid branch

  run_workflow "$workflow" -f branch="$ERRORS_BRANCH"  # errors branch
  run_workflow "$workflow" -f branch="$ERRORS_BRANCH" \
    -f go-version="$OPTION_GO_VERSION"  # errors branch with custom Go version

  run_workflow "$workflow" -f branch="$SUCCESS_BRANCH"   # No issues branch
  run_workflow "$workflow" -f branch="$SUCCESS_BRANCH" \
    -f timeout="1m" # No issues with TO
done

# Run badges generation workflow
run_workflow "gwy-badges.yml" -f branch=invalid
run_workflow "gwy-badges.yml" -f branch="$SUCCESS_BRANCH" -f badges-branch="<NONE>"
run_workflow "gwy-badges.yml" -f branch="$ERRORS_BRANCH" -f badges-branch="<GH-PAGES>"

run_workflow "gwy-badges.yml" -f branch="$ERRORS_BRANCH" -f badges-branch="badges" \
  -f badges-directory="images/badges" -f badges-url="github.com"

run_workflow "gwy-badges.yml" -f branch="$ERRORS_BRANCH" -f badges-branch="badges" \
  -f badges-directory="images/badges" -f badges-url="githubusercontent.com"

# Run release workflow with custom branch
run_workflow "gwy-aws-release.yml" -f branch="$ERRORS_BRANCH"
