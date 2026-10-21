# README skeletons by project type

Pick the skeleton that matches what the reader is deciding, not what the code happens to be. A library and a CLI can share a language and still need opposite orderings, because one reader is deciding "does this fit my codebase" and the other "does this replace a tool I already run".

Every skeleton below is a starting order, not a template to fill blindly. Drop sections the project does not need; the median README has seven sections and half sit between five and twelve.

## Library or SDK

The reader is evaluating whether to add a dependency. They care about fit and cost of reversal before they care about features.

1. Name and one-line description
2. Badge row (≤4)
3. Two-to-four sentence description: the problem, the approach, the shape of the API
4. Why this over the alternatives - concrete axes only
5. Usage: the smallest complete example, with its output
6. Install, including version/runtime requirements
7. Core API: the handful of entry points, not the full surface
8. Status and stability guarantees (semver policy, breaking-change history)
9. Links out: full documentation, more examples, support
10. Licence, maintainers, contributing pointer

Put usage above install: a reader who dislikes the API shape does not care how to install it, and one who likes it will scroll.

## CLI tool

The reader is comparing against a tool already on their machine. Show the invocation and the output immediately.

1. Name and one-line description
2. A terminal transcript: one real command and its real output
3. Why this over the incumbent tool
4. Install, per platform, with the package-manager one-liners first
5. Common invocations - three or four, each with a comment saying what it is for
6. Configuration: only the options most users touch; link the rest
7. Status, versioning, upgrade notes
8. Licence, maintainers, contributing pointer

A recorded terminal session is fine as an addition, never as the only place the commands appear - a reader on a registry page or in a terminal sees nothing.

## Application, service, or self-hosted product

The reader is deciding whether to run infrastructure. Operational cost dominates.

1. Name, one-line description, screenshot with alt text
2. What it does and who it is for
3. Why this over the hosted or incumbent option
4. Try it: hosted demo, one-command container run, or sandbox
5. Requirements: runtime, memory, storage, external services
6. Install and first-run configuration
7. Architecture in a short paragraph, or one diagram plus a text summary
8. Security and data handling posture
9. Status, release cadence, upgrade and backup notes
10. Support channels, licence, contributing pointer

## Framework or platform

The reader is deciding whether to build on top of you, which is the least reversible choice of all.

1. Name and one-line description
2. The mental model in one paragraph: the two or three concepts everything else derives from
3. Why this over the alternatives, including what it deliberately does not do
4. A complete small application, end to end
5. Install and scaffold command
6. Ecosystem and extension points
7. Stability and long-term-support commitments - this reader will ask about the five-year horizon
8. Documentation, community, governance, licence

## Template, starter, or learning resource

The reader is deciding whether to copy it.

1. Name and one-line description
2. What you will have after using it, stated as an outcome
3. Prerequisites and assumed knowledge
4. Use it: the clone/scaffold command and the first thing to run
5. What is inside: the directory map, briefly
6. What to change first
7. Currency - which versions this tracks and when it was last verified, because a stale starter is worse than none
8. Licence and attribution terms

## Minimal skeleton

For a small, single-purpose project, four sections are a complete README. Padding it does not make it look more serious.

1. Name and one-liner
2. Usage example with output
3. Install
4. Licence

## Sections that usually belong elsewhere

Move these out and link to them; each has a conventional home that tooling already understands.

| Content                                          | Where it belongs                                  |
| ------------------------------------------------ | ------------------------------------------------- |
| Full contribution process, dev environment setup | `CONTRIBUTING.md`                                 |
| Release history                                  | `CHANGELOG.md`                                    |
| Behaviour expectations                           | `CODE_OF_CONDUCT.md`                              |
| Vulnerability reporting                          | `SECURITY.md`                                     |
| Full API reference                               | documentation site or generated reference         |
| Step-by-step teaching material                   | tutorial in the docs                              |
| Hand-written table of contents                   | nothing - hosts generate an outline from headings |

Leave a one-line pointer in the README for each. The point is not to hide the content; it is to keep the README at the length where a reader still reads it.
