# README audit scorecard

Score the README as it exists today, before proposing any rewrite. Report every check with the evidence that decided it - a line number, a quoted sentence, or a command you ran. A check with no evidence is an opinion, and the user is entitled to reject it.

State the declared audience - solo developer, company or procurement evaluator, or both - on the first line of the report. Several checks resolve differently depending on it, so a scorecard with no declared audience cannot be audited by anyone else.

## Band A - blocking checks

Each is pass or fail. A single failure means the README is not fit to publish, whatever the rest scores.

**A1 - Identity in the first two lines.** The project name and one sentence saying what it is appear before anything else a reader must scroll past. Every unusual term in that sentence is either defined inline or linked. Fail if the first thing under the title is a badge wall, a logo the size of a screen, or a sentence that only makes sense to someone who already uses the project.

**A2 - The first command runs.** Take the install command and the first usage example exactly as written, run them in a clean environment, and record what happens. Fail on any undocumented prerequisite, any command that errors, any output that contradicts the README. This is the check that catches the most real damage, because a reader who hits a broken install leaves and does not file an issue.

**A3 - Every claim is verifiable.** List every command, flag, default, environment variable, config key, endpoint, exported symbol, file path, version requirement, and behavioural assertion in the file, then check each against its own source of truth in the repository - the argument parser, the definition site, the route table, the changelog. Fail on any documented flag the parser does not register, any signature that does not match, and any performance number, compatibility matrix or scale claim with no benchmark, CI matrix or changelog entry behind it. Report the count of unverifiable claims on its own line, because it is the finding maintainers most often dispute and most often lose.

**A4 - Status is stated honestly.** A reader can tell whether the project is experimental, actively maintained, feature-frozen, or looking for a new maintainer, without opening the commit history. Fail if the README implies more stability than the project has. Only 21.4% of READMEs carry status content at all (Prana et al. 2019), so stating it is both rare and cheap.

**A5 - Licence is reachable from the README.** Named in the README, linked to the file. Fail if a reader has to guess, and fail harder if the licence is non-permissive and buried at the bottom; that is the fastest disqualifier for a commercial reader, so it belongs near the top where it saves their time. 64% of respondents to GitHub's 2017 Open Source Survey called the licence very important in deciding whether to use a project, the highest-rated documentation element in that survey.

## Band B - scored checks

Score each 0, 1, or 2. Zero means absent, one means present but weak, two means it does its job.

**B1 - Differentiation.** The README says why to choose this over the obvious alternative, in the project's own words, without disparaging anyone. This is the scarcest content category measured - only 25.7% of READMEs have any (Prana et al. 2019) - so it is usually the highest-leverage check on the board.

- 2 points: a concrete axis - a benchmark, a dependency count, a supported-platform difference, a design trade-off deliberately taken.
- 1 point: a vague "simple and fast".

**B2 - Time to a copyable command.** A command the reader can paste appears within roughly the first screen.

- 2 points: inside the first 400 words.
- 1 point: arrives after a long preamble.
- 0 points: no code block at all.

**B3 - Runnable usage example.** The example is real code a reader could execute, and it shows the output or the effect. Two points if the same snippet also exists as a file in the repository. Zero for a bare list of function signatures; that is API reference material, not usage.

**B4 - Funnel ordering.** Sections are ordered so a reader can disqualify the project as early as possible: identity, then differentiation, then usage, then install, then detail, then background. Two points if a reader can stop at any point without having missed something that would have changed their decision.

**B5 - Badge discipline.** Each badge should be an assessment signal answering a question a visiting reader actually has: is it maintained, what version, what licence, does it work on my platform.

- 2 points: four badges or fewer.
- 1 point: a tolerable row.
- 0 points: a badge wall, or a row built on conventional signals such as "PRs welcome".

Trockman et al. (ICSE 2018, 294,941 npm packages) found that adopting badges does not raise popularity, and that among already-popular packages excessive badge use correlates with decreased popularity; the _direction_ is sourced, the cap of four is this skill's own baseline. Build-status badges mostly serve maintainers, who are better served by a failing-build notification.

**B6 - Nothing critical trapped in an image.** Every image has alt text, and no instruction, command, or decision-relevant fact exists only inside a screenshot, diagram, or animated demo. Readers meet READMEs in terminals, in package-registry pages that strip HTML, and through screen readers.

**B7 - Links resolve everywhere.** Relative links resolve on the default branch, and the README still makes sense when rendered on the package registry page, where relative paths and raw HTML often break. Two points requires actually checking the registry render, not assuming it. Confirmed on both npm and PyPI: both registries render the README on their own domain rather than the source repository's, so a relative image or link path that works on the code host resolves to a dead link on the registry page - PyPI in particular has no support for relative image paths at all, and the standard fix is converting to absolute URLs (e.g. `raw.githubusercontent.com`) at publish time.

**B8 - Overflow lives elsewhere.** Long reference material, tutorials, configuration tables, and community health content (contributing, code of conduct, security policy, support) are linked, not inlined. Half of all sampled READMEs sit between 5 and 12 sections with a median of 7 (Prana et al. 2019), so a count in that band is the normal signal; more is fine when the project genuinely has that many top-level concerns, but say why.

## Pass threshold

The README passes when all five Band A checks pass **and** Band B totals **14 of 16 or more**.

**This threshold is this skill's own baseline, not a published standard.** The automated README scorers in the field measure structural proxies rather than persuasion. Treat 14/16 as the default and let the user move it for a stated reason.

Below the threshold, rewrite and re-score rather than shipping with a note. Iterate until it is met or the user explicitly accepts a lower score for a stated reason; record that reason in the report so the next reviewer does not relitigate it.

## The cold-reader test

Run this after the rewrite, as the outcome check that the scorecard is only a proxy for.

Give the README to a reader who has never seen the project - a person if the user can find one, otherwise a fresh agent session with no repository context. Allow 60 seconds, from the top, no scrolling past the fold more than once. Ask five questions:

1. What is this?
2. Why would I use it instead of the alternative I already know?
3. Does it fit my language, platform, and licence constraints?
4. What is the first command I would run?
5. Is this project alive?

Pass is five correct answers within the minute. The 60-second window, the five questions, and the five-of-five bar are this skill's own construction, like the 14/16 threshold - not a published protocol - move them with the user for a stated reason. Anything less than a pass points at a specific check, which makes the failure directly actionable: a wrong answer to question 2 is a B1 problem, to question 5 an A4 problem.

## Reporting format

```
README audit ; acme/flowmatic   (audience: solo developer)
Band A: A1 pass · A2 FAIL (npm install needs Node >=20, undocumented) · A3 FAIL · A4 pass · A5 pass
Claim check: 2 unverifiable ; "handles 10k req/s" (no benchmark in repo), --watch flag not in parser
Band B: 11/16  (B1 0, B2 2, B3 1, B4 2, B5 0, B6 2, B7 2, B8 2)
Verdict: does not pass ; A2 and A3 blocking; B1 returns the most per hour spent
Cold-reader test: 3/5 ; readers could not answer "why this over X" or "is it alive"

Fix in this order:
1. Document the Node version requirement and re-run the install from clean (A2)
2. Remove or source the throughput claim; drop --watch from the docs (A3)
3. Add three sentences on why this over <alternative> (B1)
4. Cut the badge row from 9 to 3 (B5)
```

The fix list is ordered by return per hour spent, using the ordering in SKILL.md § Order the fixes, and re-ranked against the user's deadline and effort ceiling. Say in the report which of their answers moved which fix.

For an enterprise or procurement audience, add one line listing which trust signals are present and which are missing, in the order SKILL.md § Step 8 ranks them:

`support expectation == release cadence > security policy > named adopters > governance and maintainer depth > supply-chain attestation > foundation affiliation`

That line doubles as the order to fix them in. Those are not scored; they are pass/absent facts a buyer will check anyway.
