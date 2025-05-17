
# Scripts description and releases testing

## Scripts

- `run-jobs.sh`: This script is used to run all workflows manually, with 
  different parameters, to ensure new releases are working fine.

- `install-gwy.sh`: This script is a modified copy of GWY repo's install 
  script modified to check out the latest release candidate branch instead 
  of the master branch, so we can test fully the latest release candidate 
  install process in this repository (GWY Playground).

- `update-gwy.sh`: This script will install GWY in the `ci-base` branch 
  using `install-gwy.sh` modified copy of GWY install script, and rebase all 
  the branches in this repo with the new version installed so it can be 
  tested fully before releasing the latest release working candidate.

- `remove-jobs.sh`: This script allows you to remove all run jobs older 
  than a given time from a specified repository.

## How to test a GWY feature branch or new release candidate?

1. Edit `install-gwy.sh` BRANCH variable with the name of the branch in GWY 
   repository holding the version we want to test.

2. Copy to project root the `update-gwy.sh` script and run it, this script 
   will use `install-gwy.sh` to install the specified GWY branch code in 
   `ci-base` branch, which all branches in this repo branch from, and rebase 
   all branches with the new version, updating all branches GWY files with 
   the version we want to test :)

3. Do whatever manual tests are required and never forget to run `run-jobs.sh` 
   script and check the dozens of workflows triggered manually to ensure all 
   workflows are working correctly.
4. 
