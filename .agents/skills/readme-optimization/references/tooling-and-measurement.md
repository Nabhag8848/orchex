# Tooling and measurement

An optional integration note. The skill's instructions are written to be tool-agnostic; this file names the specific tools that exist so you can suggest one when the user's environment allows it. Verify a tool is still alive before recommending it - several in this space have been sunset.

## Structural README scorers

| Tool                                            | What it measures                                                                                            | Scale          |
| ----------------------------------------------- | ----------------------------------------------------------------------------------------------------------- | -------------- |
| ScoreMe / `readme-score` (Clay Allsopp)         | structural complexity of README text                                                                        | 0-100          |
| `readme-inspector` (wraps the readme-score API) | existence and quality of a README, usable in CI                                                             | 0-100          |
| trymarkdownviewer README Health Score           | heading hierarchy 20%, GFM compliance 30%, image alt text 15%, link integrity 25%, task-list completion 10% | 0-100, A+ to F |
| readmecodegen README Score Checker              | structural proxies with contextual weighting                                                                | 0-100          |

ScoreMe's own description is the honest one: it measures "complexity, which is generally correlated with quality", and it explicitly disclaims judging whether one README is better than another. Treat "complexity correlates with quality" as the author's unvalidated assumption.

The code host's own **community-standards view** is a binary added/not-added checklist over the recognized health files (README, code of conduct, contributing, licence, security policy, issue and PR templates, support, funding, citation, governance). It is a hygiene baseline, not a score.

## Security and package-health scorers

Relevant only when the audience is an enterprise evaluator.

- **OpenSSF Scorecard** - automated security checks (security policy present, packaging, CI tests, code review, best-practices badge), each with a documented risk level. Enterprises do use it in dependency acceptance criteria.
- **OpenSSF Best Practices Badge** (formerly CII) - passing, silver, gold tiers. Unlike Scorecard it accepts claims with justifications from maintainers, trading manual effort for fewer false negatives. Gold requires multiple developers on the project.
- **SLSA** - supply-chain provenance levels, designed for a reader judging organizational risk.
- **Libraries.io SourceRank** - weights how many other projects depend on a package plus documentation and SemVer presence; explicitly positioned as an alternative to star counts. A 2.0 redesign may not be live; verify first.
- **Snyk Advisor** - rated packages across popularity, maintenance, security and community. Being sunset: the standalone site folded into Snyk's security database and went offline in January 2026. Do not recommend it.

## Prose and Markdown linters

- **Vale** - syntax-aware prose linter that understands Markdown, so it excludes code blocks from prose rules. Bundles Microsoft and Google style guides and can enforce a readability score.
- **markdownlint** - structure and formatting.
- Companions: `alex` (inconsiderate wording), `write-good`, `proselint`, `textlint` - heuristic and largely non-configurable.

The common practitioner pattern is Vale plus markdownlint in CI as a merge gate, which stops the README regressing after the rewrite.

**LintMe** (Mynampaty et al., CHI 2026, DOI 10.1145/3772318.3791597) combines programmatic checks with LLM-based content evaluation such as jargon detection. Its own finding is the clearest academic statement of the ceiling: "what constitutes a good README varies across audiences and contexts", and existing linters "concentrate on form rather than substance".

## Terminal demo recording

- **asciinema** records a terminal session as a lightweight text `.cast` file, convertible to GIF with `agg`. Small, and stays crisp at any zoom.
- **VHS** (Charm) renders a GIF, MP4 or WebM from a `.tape` script specifying window size, theme, typing speed and pauses. The advantage over a hand-recorded capture is that the tape file is diffable and re-renderable, so the demo can be regenerated when the CLI's real output changes rather than silently going stale.

Real costs, independent of the conversion question: a GIF can dominate the page weight, and animated content has no alt text, so it is invisible to screen readers and to reading the raw file. Never let a demo carry information that appears nowhere in text.

## Measurement surfaces

- **Code-host traffic insights** - views, unique visitors, clones, referring sites. First-party and needs no integration, but retention is short (daily data is a rolling two-week window on GitHub), so snapshot it rather than relying on it retroactively.
- **Scarf** - cookie-free attribution that works on surfaces the code host cannot see: a gateway redirect that attributes downloads by company via reverse IP lookup, and a pixel embedded in the README or docs.
  - It is the only practical way to attribute a registry-page or locally-rendered README view.
  - Scarf sells this product and frames documentation views as a funnel step; weight that framing accordingly.
  - It states it respects Do Not Track and purges personal data after metadata lookup - confirm the current privacy posture against your own policy before embedding a pixel in a public README.
- **Star-history trackers** - useful for a launch narrative only. Stars are cumulative and monotonic, so they cannot show a regression.
- **UTM-tagged outbound links** from the README to the docs site give a clean click-through number without any third-party pixel.

Funnel frameworks that place a README as a step: AAARRRP (Phil Leggetter, 2016, adapted from Dave McClure's pirate metrics) and the Orbit Model (Josh Dzielak and Patrick Woods), which replaces funnel stages with engagement-depth orbits precisely because community behaviour is not linear. Practitioners including Ben Greenberg have argued publicly against measuring developer relations with a strict sales funnel at all.

## The validity ceiling

Every tool above measures structural completeness, hygiene, or traffic. None measures whether a reader was persuaded to try the project. A high score on any of them is compatible with a README nobody understands.

Use a scorer to catch mechanical regressions, and use the cold-reader test in `audit_scorecard.md` for everything that actually decides adoption.
