# Published findings

Every figure this skill quotes, with its source and its strength, plus the folklore numbers to refuse. Read this before repeating a number to a user, and cite the study rather than the skill.

## Strong sources

**Prana, Treude, Thung, Atapattu, Lo - _Categorizing the Content of GitHub README Files_, Empirical Software Engineering 2019 (arXiv 1802.06997).** 393 randomly sampled repositories, 4,226 manually annotated sections, inter-rater kappa 0.858.

Presence of each content category, per file:

| Category                                         | Present in | Reading                |
| ------------------------------------------------ | ---------- | ---------------------- |
| What (introduction, background)                  | 97.0%      | table stakes           |
| How (install, usage, config, requirements)       | 88.5%      | table stakes           |
| References (links out, API docs, support)        | 60.8%      | common                 |
| Who (team, contact, licence, credits)            | 52.9%      | common                 |
| Contribution                                     | 27.8%      | scarce                 |
| **Why (advantages, comparison to alternatives)** | **25.7%**  | **the differentiator** |
| **When (status, versions, roadmap)**             | **21.4%**  | **the differentiator** |
| Other (incl. tables of contents)                 | 6.9%       | -                      |

Size benchmark from the same sample: **median 7 sections**, with half of all files between **5 and 12**. Files under 2 KB were excluded as effectively empty.

Sample composition:

- end-user applications: 42%
- libraries: 27.9%
- learning resources: 17.4%
- frameworks: 7.3%
- UI: 5.4%

Association rules at file level:

- a file containing Who also contains What, with 0.98 confidence
- What implies How, with 0.89 confidence

Categories cluster, so a missing Why is rarely an isolated omission; it signals the README was written as a manual rather than as an evaluation aid.

**Trockman, Zhou, Kästner, Vasilescu - _Adding Sparkle to Social Coding: An Empirical Study of Repository Badges in the npm Ecosystem_, ICSE 2018 (DOI 10.1145/3180155.3180209).** 294,941 npm packages. The strongest evidence base on any single README element. The ACM DOI page itself still 403s to automated fetch, but the 46% badge-adoption figure below is independently re-confirmed via alternate access points (ResearchGate's hosted PDF, the ICSE companion PDF on src.acm.org, and the paper's own supplementary-materials repo, CMUSTRUDEL/npm-badges).

- 46% of packages carry at least one badge.
  - Of those, 66% adopted multiple kinds.
  - Of those, 82% did so within 24 hours - badges get added in a batch.
- 88 distinct badge types across six classes. The most common single badge (CI build status) appeared on 31.0% of packages.
- Survey of 32 maintainers and 57 contributors: 88% of maintainers treated badge presence as a quality indicator against only 53% of contributors, and 61% of contributors said badges do not influence their decision to contribute at all.
- Signalling theory distinction:
  - **assessment signals** - backed by a third-party check and cannot pass unless the underlying fact holds (CI status, coverage, dependency freshness)
  - **conventional signals** - merely state an intent ("PRs welcome") and are reliable evidence of nothing
- Popularity: "The adoption of a badge does not change popularity. However, popular projects are more likely to have badges." Among already-popular packages, excessive badge use correlates with _decreased_ popularity.

**GitHub 2017 Open Source Survey** (open dataset, Zenodo record 806811).

- Incomplete or outdated documentation is observed by 93% of respondents, yet 60% of contributors rarely or never contribute to documentation.
- 64% say an open source licence is very important in deciding whether to _use_ a project.
- 67% say it is very important in deciding whether to _contribute_.

The survey itself is robust. It is cross-sectional, so it measures perception at one point in time; it does not prove that adding a file raises adoption.

## Named methods worth citing

**Cognitive funneling** - named by Kira (hackergrrl) in _Art of README_ (CC-BY). Order sections by how quickly each lets a reader bail: name, one-liner, runnable usage, API, installation, licence - with a non-permissive licence moved up, because it disqualifies fastest.

- Stated posture, ego-checking: "Your job is to let them evaluate what your creation does as objectively as possible - not to maximize your downloads or userbase."
- Brevity rule: "The ideal README is as short as it can be without being any shorter."

Counterpoint worth surfacing to a user: makeareadme.com (Danny Guo) argues "too long is better than too short" and recommends moving overflow to another document rather than deleting it. Both agree overflow leaves the README; they disagree on the risk of cutting. The Good Docs Project frames the same rule as keeping "the minimum needed to start using the project" in the README.

**Readme Driven Development** - Tom Preston-Werner, 2010: "Write your Readme first. First. As in, before you write any code or tests or behaviors or stories or ANYTHING." And: "A perfect implementation of the wrong specification is worthless." Useful when the user is starting a project rather than fixing one.

**Ken Williams (Perl)**, on the completeness bar: "Your documentation is complete when someone can use your module without ever having to look at its code… the documentation, not the code, defines what a module does."

**Diátaxis** (Daniele Procida) classifies documentation into tutorial, how-to, reference and explanation. A README is none of them; it is a landing page and router that points at all four. Diátaxis therefore governs the docs site, not the README's internal order.

## Enterprise evaluation signals

- **Div Manickam**, on the buyer/user split: "Buyer: Needs to see ROI, security compliance, and long-term stability. User: Needs to see low latency, ease of implementation, and a strong community… Trust is earned in the code, not the copy."
- **Adam Jacob** (Chef, System Initiative), on the anxiety behind every enterprise trust signal: "Do I trust the supply chain? How do I know? Who do I call if there's a problem? What if there's a security vulnerability and the guy who does the patches is on vacation?… that tends to swing you right back to the proprietary one." (Changelog podcast #460.)
- **Mitchell Hashimoto**, on enterprise technical decision-makers: "90% of TDMs are motivated primarily by NOT GETTING FIRED… they follow secular trends supported by analysts and broad public sentiment." Reported secondhand via Simon Willison, 12 May 2026; attribute it that way.
- Foundation maturity is a real differentiator:
  - a graduated project in a major foundation has undergone a security audit and has a documented vulnerability disclosure process
  - a sandbox-tier project has not
- A recommended enterprise checklist, drawing on practitioners at CNCF and Buoyant:
  - maintainer depth and employer diversity
  - release cadence over the last twelve months
  - time to fix important bugs and security issues
  - governance clarity and roadmap ownership
- The OpenSSF Concise Guide for Evaluating Open Source Software, Microsoft's Code-with-Engineering Playbook, and several published dependency-evaluation checklists independently weight maintenance activity, licence clarity and security posture **above presentation quality**.

## Weak or vendor-influenced sources

- Attribution and "adoption funnel" framings mostly come from vendors selling the measurement product. The framing is useful; the implied causality is marketing.
- Demo-GIF and visual-content advice (short, dark-themed, legible, placed after the tagline) is practitioner consensus, not a measured result. Say so when recommending it.
- Growth-blog case studies bundle README work with a launch spike, so the README's isolated effect cannot be separated from the traffic event.

## Folklore - never cite

- "Repos with detailed READMEs get 50% more contributions, per GitHub's own research." The attribution to GitHub is invented.
- "A star-history chart increases star conversion by ~15%." Self-reported on a growth blog with no methodology.
- "62% of top-100 repos include GIF or video demos." Traces to unverifiable vendor blogs and is correlational at best.

If a user brings one of these, name it as unsourced rather than building the plan on it.
