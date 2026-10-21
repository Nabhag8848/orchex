---
name: current-state-docs
description: Create or update repository Markdown, especially READMEs and technical docs, so claims accurately reflect the codebase's current implemented state. Use whenever writing or revising project-facing .md files.
---

# Current-State Documentation

Ground documentation in the repository as it exists at the time of writing.

- Inspect the relevant source, configuration, manifests, scripts, tests, and existing docs before making factual claims.
- Describe implemented and verified behavior in the present tense.
- Distinguish clearly among working functionality, partial or experimental functionality, planned work, and unsupported behavior. Never present roadmap items as current features.
- Include a concise **Current status** section when it helps readers understand maturity, known gaps, or what works today. Do not add one mechanically when the same information is already obvious and maintained elsewhere.
- Prefer durable facts over volatile counts, dates, version numbers, or exhaustive file lists. When volatile details are necessary, derive them from authoritative repository sources and identify their source of truth.
- Ensure commands, paths, configuration keys, prerequisites, and examples match the current repository. Run safe verification commands when practical.
- Preserve useful existing content, but remove or correct stale, contradictory, aspirational, or unverifiable claims.
- If the current state cannot be established, label the uncertainty or ask for the missing information instead of guessing.

Before finishing, compare the changed Markdown against the inspected evidence and report any claims that could not be verified.
