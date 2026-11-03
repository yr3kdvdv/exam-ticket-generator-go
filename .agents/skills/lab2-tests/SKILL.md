---
name: lab2-tests
description: Run and diagnose the Go unit tests, Fyne workflow tests, and GitHub Actions checks for the Lab 2 ticket generator.
---

Run from the repository root.

1. Run `go test -count=1 ./internal/lab2` for ticket parsing, student data, and repeat issuance.
2. Run `go test -count=1 -v ./cmd/lab2` for the virtual Fyne windows. The issuance test uses copies of the bundled input files and writes results to a temporary directory.
3. Run `go test ./...` and `go vet ./...` for the full project. If native Fyne headers are missing on Linux, use the dependency list in `.github/workflows/go.yml`.
4. For GitHub Actions, match the run to the requested commit with `gh run list --commit <sha>`. Inspect a failing job with `gh run view <run-id> --log-failed`. If asked to trigger CI without pushing a commit, use `gh workflow run go.yml --ref <branch>` once the workflow with manual dispatch is on the default branch.

Report what passed or failed. Distinguish the virtual GUI test and Windows build from running the GUI on a Windows desktop.
