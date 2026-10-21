# Worked examples

Positive examples are quoted from real published READMEs and attributed. Negative examples are constructed anti-patterns - they are composites of common failures, not quotations of any real project, and should never be attributed to anyone.

## Table of Contents

- [The opening lines](#the-opening-lines)
- [Stating why over the alternative](#stating-why-over-the-alternative)
- [Stating project status](#stating-project-status)
- [The badge row](#the-badge-row)
- [The install section](#the-install-section)
- [The usage example](#the-usage-example)

## The opening lines

**Good** - `d3/d3`'s published opening:

> **D3** (or **D3.js**) is a JavaScript library for visualizing data using web standards. D3 helps you bring data to life using SVG, Canvas and HTML.

It names the technology, the job, and the medium in two sentences. A reader who does not work in the browser leaves immediately, which is the correct outcome for both parties.

**Good** - `jmilleralpine/ParallelGit`'s published opening:

> A high performance Java JDK 7 nio in-memory filesystem for Git.

Fourteen words carrying language, runtime requirement, mechanism, and purpose.

**Bad** - constructed:

> # flowmatic
>
> [badge][badge][badge][badge][badge][badge][badge]
>
> <img src="logo.png" width="700">
>
> Flowmatic makes it easy to work with flows. Built with ❤️ and TypeScript.

The reader is four scrolls in and still does not know what a "flow" is. "Makes it easy" says nothing measurable, and the logo pushes the only informative sentence below the fold.

## Stating why over the alternative

**Bad** - constructed:

> ## Why flowmatic?
>
> Flowmatic is simple, fast, lightweight and developer-friendly.

Every project claims this, so it carries no information and costs a section.

**Better** - constructed, showing the shape to aim for:

> ## Why flowmatic
>
> Use it when you need retries and backoff but not a scheduler. It has no runtime dependencies and adds about 12 kB gzipped, against roughly 300 kB for a full workflow engine. If you need cron scheduling, durable state across restarts, or a UI, use a workflow engine instead - flowmatic deliberately has none of those.

Concrete axes, a number the reader can verify, and an explicit statement of when _not_ to use it. Naming the case against your own project is the fastest way to earn a developer's trust, and it filters out the users who would have churned anyway.

## Stating project status

**Better** - constructed:

> Flowmatic is in early beta. Retry and backoff are stable; scheduling and durable state are not built yet. That said, we run it ourselves every day, and we'd rather you find the rough edges than we find them for you.

Honest about maturity, specific about the risk, and still gives a reason to try it.

**Bad** - constructed:

> ## Status
>
> Production ready! 🎉

Unfalsifiable, and contradicted the moment a reader checks the issue tracker.

## The badge row

**Bad** - constructed: nine badges covering build status, coverage, three package registries, licence, contributor count, chat, and "made with love".

Ask of each one: what decision does a first-time visitor make with this? Coverage percentage and contributor count almost never change an adoption decision. Build status tells the reader the maintainers' CI is red, which is information for the maintainers.

**Good** - four badges: latest released version, licence, supported runtime versions, and last-release date or maintenance status. Each answers a question from the cold-reader test.

## The install section

**Bad** - constructed:

> ## Installation
>
> ```
> npm install flowmatic
> ```

No runtime constraint, no import line, no verification step. This is the single most common cause of a blocking A2 failure in the scorecard.

**Better** - constructed:

> ## Install
>
> Requires Node 20 or later.
>
> ```bash
> npm install flowmatic
> ```
>
> Verify:
>
> ```bash
> node -e "console.log(require('flowmatic').version)"
> # 2.4.1
> ```

The verification step converts a silent failure into an immediate, self-diagnosing one.

## The usage example

**Bad** - constructed:

> ## API
>
> - `createFlow(opts)`
> - `run(flow, input)`
> - `cancel(handle)`

Signatures without a scenario. The reader cannot tell what a flow is, what `opts` accepts, or what running one produces.

**Better** - constructed:

> ## Usage
>
> Retry a flaky HTTP call three times with exponential backoff:
>
> ```js
> import { retry } from "flowmatic";
>
> const body = await retry(() => fetch("https://api.example.com/health"), {
>   attempts: 3,
>   backoff: "exponential",
> });
>
> console.log(body.status);
> // 200
> ```
>
> This example is in [`examples/retry.js`](./examples/retry.js) and runs with `node examples/retry.js`.

One scenario, real code, the output shown, and the snippet exists as a file so a reader who clones the repository can run it unmodified.
