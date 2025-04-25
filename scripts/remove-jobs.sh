#!/bin/bash

# Check if all three arguments are provided
if [ "$#" -ne 3 ]; then
  echo "This script allows you to remove bulk GitHub workflow runs older than the specified time frame"
  echo "Usage: $0 <username> <repo> <time_frame>"
  echo "Example: $0 earcamone gwy-playground 1h (for 1 hour), 10m (for 10 minutes), 2d (for 2 days)"
  exit 1
fi

USER="$1"          # GitHub username
REPO="$2"          # Repository name
TIME_THRESHOLD="$3" # Time threshold (e.g., 1h, 10m, 2d)

# List of job IDs to skip (space-separated, add your IDs here)
SKIP_IDS="15073819303 15073827359 15073827669"

# Function to convert time threshold to seconds
convert_to_seconds() {
  local input="$1"
  local value="${input%[a-z]}"  # Extract number
  local unit="${input##*[0-9]}" # Extract unit (h, m, d)

  case "$unit" in
    h) echo $((value * 3600)) ;;  # Hours to seconds
    m) echo $((value * 60)) ;;    # Minutes to seconds
    d) echo $((value * 86400)) ;; # Days to seconds
    *) echo "Invalid unit: $unit. Use h, m, or d." >&2; exit 1 ;;
  esac
}

# Get current time in seconds since epoch
CURRENT_TIME=$(date +%s)

# Calculate cutoff time in seconds
SECONDS_THRESHOLD=$(convert_to_seconds "$TIME_THRESHOLD")
CUTOFF_TIME=$((CURRENT_TIME - SECONDS_THRESHOLD))

# Fetch and filter workflow runs, then process those older than the cutoff
gh api "repos/$USER/$REPO/actions/runs" --paginate -q '.workflow_runs[] | {id: .id, created_at: .created_at}' | \
jq -r --argjson cutoff "$CUTOFF_TIME" '
  # Convert created_at to epoch time and compare
  select((.created_at | sub("Z"; "") | sub("\\.[0-9]+"; "") | strptime("%Y-%m-%dT%H:%M:%S") | mktime) < $cutoff) | .id
' | while read -r run_id; do

  # Check if run_id is in SKIP_IDS (more robust check)
  skip=false
  for skip_id in $SKIP_IDS; do
    if [ "$run_id" = "$skip_id" ]; then
      skip=true
      break
    fi
  done

  if [ "$skip" = false ]; then
    echo " - Deleting workflow run ID: $run_id (older than $TIME_THRESHOLD)"
    gh api -X DELETE "repos/$USER/$REPO/actions/runs/$run_id"
  else
    echo " X Skipping workflow run ID: $run_id (in skip list)"
  fi
done

echo -e "\nDone, bro! All runs older than $TIME_THRESHOLD are toast (except the ones you wanted to keep)!"

