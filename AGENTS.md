# AGENTS.md

## Collaboration Mode

This repository is built in pair-programming mode.

Do not generate, scaffold, or edit implementation code unless the user explicitly asks for it. Default to explaining tradeoffs, reviewing code, suggesting next steps, and answering questions.

Documentation changes are allowed when requested.

## Project Plan

Use `docs/implementation-plan.md` as the product and architecture reference. The plan is guidance, not automatic permission to implement.

Keep the plan current as the project progresses. When a milestone is completed, changed, deferred, or replaced, update the relevant section so the document reflects the current project direction.

## Go Style Guide

Follow Uber's Go Style Guide for Go code style and review feedback.

Prefer:

- clear package boundaries
- small interfaces owned by consumers
- explicit error handling
- `context.Context` as the first parameter where applicable
- table-driven tests for meaningful behavior
- `gofmt` and idiomatic standard-library patterns

## Review Expectations

When reviewing completed features, focus first on:

- correctness bugs
- security issues
- data ownership and tenant isolation
- error handling
- test coverage gaps
- maintainability and readability

Give suggestions as review feedback, not automatic edits.

## Implementation Requests

When the user explicitly asks for implementation:

- inspect the existing code first
- keep changes narrowly scoped
- preserve user changes
- explain the approach before substantial edits
- verify with relevant tests or commands when available
