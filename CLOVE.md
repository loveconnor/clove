# CLOVE.md

## Project Context

CLOVE is a trust-first GitHub alternative: a review-centered software forge for teams that want GitHub-compatible workflows without GitHub's drift toward opaque pricing, bundled platform risk, AI ambiguity, degraded review ergonomics, and vendor lock-in.

The product is not "GitHub, but cheaper." The wedge is dependability, trust, portability, and clarity.

Core thesis:

> Build the most dependable review-centered forge for teams that want GitHub compatibility without GitHub drift.

The product should feel familiar enough for GitHub users to migrate, but more legible, predictable, and respectful of user agency.

---

## Primary Product Promise

CLOVE should help teams do three things exceptionally well:

1. **Review code safely**
   - Fast, stable pull-request workflows.
   - Clear review state.
   - Explicit patch-set history.
   - Human comments, automated checks, and AI suggestions visually separated.
   - Strong keyboard navigation and accessibility.

2. **Run delivery without surprise**
   - CI/CD and package registries should feel native.
   - Compute, artifact, and package surfaces must be isolated from the core forge.
   - Pricing must include hard caps, preflight estimates, transparent metering, and budget controls.
   - Failures in CI or packages should not paralyze repository access or code review.

3. **Administer access without pain**
   - SAML, SCIM, SSO, audit logs, external guest accounts, custom roles, and policy inheritance should be first-class early.
   - Admins must be able to answer: “Who can do what, and why?”
   - Permissions, billing, identity, and compliance should be visible and explainable.

---

## Target Users

CLOVE is for:

- Serious engineering teams that like GitHub-style workflows but no longer fully trust GitHub as the whole delivery stack.
- Startups that need predictable collaboration and CI costs.
- Enterprises that need identity, policy, audit, guest access, export, and compliance from day one.
- Open-source maintainers who want better review ergonomics, project discoverability, and data portability.
- Sovereignty-conscious teams that care about self-hosting, exportability, and vendor escape hatches.

CLOVE is not primarily for:

- Users looking only for the cheapest Git host.
- Teams that want a minimal personal Git server with almost no product surface.
- Communities that exclusively prefer email-first patch workflows.
- Organizations that are already deeply locked into Atlassian or Azure DevOps and do not care about public forge/network effects.

---

## Product Principles

### 1. Trust is the product

Every feature should increase one of these forms of trust:

- Operational trust: the core forge is reliable.
- Economic trust: pricing is predictable.
- Data trust: customer code and metadata are not exploited.
- Migration trust: users can leave without punishment.
- Authorship trust: humans remain in control of public collaboration surfaces.
- Admin trust: access and policy behavior is explainable.
- Review trust: code-review state is clear, stable, and auditable.

If a feature makes the platform feel more opaque, manipulative, bundled, or surprising, it is probably wrong.

### 2. Review is the center

Pull requests / merge requests are the core product surface. Optimize for reviewers, maintainers, and admins before optimizing for dashboards or marketing surfaces.

A PR should have:

- Stable patch sets.
- Explicit “changes since last review.”
- Clear review requests.
- Clear unresolved-thread state.
- Suggested changes.
- Review snapshots.
- Merge queue support.
- Code owners.
- Approval rules.
- Required checks.
- Fast diff loading.
- Keyboard-first navigation.
- Clear separation between human comments, bot output, CI checks, and AI suggestions.

### 3. Compatibility reduces migration fear

CLOVE should be GitHub-compatible where compatibility matters:

- Git protocol over SSH and HTTPS.
- Git LFS.
- GitHub Actions workflow import or near-compatibility.
- Webhook parity where practical.
- Branch protection semantics.
- Reviewer and approver mapping.
- Package registry mirroring.
- URL durability and redirects.
- Issue, PR, label, comment, release, and wiki import.
- API completeness relative to the web UI.

Do not build a cosmetic GitHub clone. Build a workflow-compatible forge with better trust boundaries.

### 4. Product boundaries must be explicit

Avoid putting everything in one failure domain.

Separate:

- Forge core: Git read/write, refs, permissions, PR/review state.
- Execution plane: CI orchestration, hosted runners, self-hosted runners, build logs.
- Artifact plane: package registry, artifacts, container registry, package mirrors.
- Search plane: indexing, symbol search, regex search, semantic layers.
- Identity/policy plane: SSO, SCIM, audit, policy inheritance, guest access.
- Eventing plane: durable webhooks, event replay, dead-letter queues.

Users should know which subsystem is degraded and what is still safe to use.

### 5. AI must be optional, bounded, and visible

AI should assist only inside explicit user-initiated boundaries.

Rules:

- AI is opt-in at organization, repository, and user levels.
- No training on private/customer repository content by default.
- No AI-authored comments, PR descriptions, titles, or public collaboration text unless the user explicitly requested that exact action.
- AI output must be visually distinct from human authorship.
- AI should never be required for the core forge experience.
- No ads, product “tips,” sponsored content, or growth prompts inside collaboration surfaces.
- AI features should be useful add-ons, not the business model foundation.

The core product should remain complete without AI.

---

## Launch Feature Set

The first credible version should include a strong P0 core:

### Repository Core

- Git hosting over SSH and HTTPS.
- Git LFS support.
- Repository import and mirroring.
- Repository export.
- Branch protections.
- Protected tags.
- Code owners.
- Repository visibility controls.
- Web UI for commits, branches, tags, diffs, blame, and history.

### Code Review

- Pull requests / merge requests.
- Review snapshots / patch sets.
- Inline comments.
- Thread resolution.
- Suggested changes.
- Required reviewers.
- Approval rules.
- Merge queue.
- Squash, merge commit, and rebase merge options.
- “Changes since last review.”
- Clear bot/check/AI/human separation.
- Fast keyboard navigation.
- Mobile triage for urgent review actions.

### Search and Discovery

- All-branch code search.
- Lexical search.
- Symbol search.
- Regex search.
- Issue, PR, discussion, docs, and package search.
- Saved views.
- Search freshness indicators.
- Visible indexing scope.
- Multilingual repository metadata.
- README language selection or language landing pages.
- Healthy/open-maintainer discovery signals.

### Issues, Discussions, and Project Metadata

- Issues.
- Labels.
- Milestones.
- Discussions.
- Releases.
- Wikis or docs.
- Basic project health metadata.
- Maintainer responsiveness indicators.
- Security policy metadata.

### CI/CD

- First-party CI orchestration.
- GitHub Actions-compatible workflow import where feasible.
- Hosted runners.
- Self-hosted runners.
- Runner groups.
- Build logs.
- Required status checks.
- Budget caps.
- Preflight cost estimates.
- External CI integration.

### Packages and Artifacts

- Package/artifact service.
- Container registry or mirroring support.
- Package mirroring.
- Artifact retention policy.
- Hard storage budgets.
- External registry escape hatches.

### Enterprise Administration

- SAML SSO.
- SCIM.
- Audit logs.
- Organization/team hierarchy.
- Custom roles.
- Policy inheritance.
- External guest accounts.
- Fine-grained tokens.
- Admin dashboards showing effective permissions.
- Exportable compliance/audit reports.

### Accessibility and Internationalization

- WCAG 2.2 AA targets.
- Keyboard-only review flows.
- Strong focus management.
- Screen reader support.
- Contrast-safe UI.
- Zoom-safe layouts.
- Locale-aware timestamps.
- Locale-aware first day of week.
- Date/time formatting controls.
- Right-to-left readiness.
- Multilingual project metadata.

---

## UX Rules

### Pull Requests

The PR page is sacred. Keep it fast, legible, and reviewer-centered.

Do:

- Make review state obvious.
- Show what changed since each reviewer last reviewed.
- Keep human discussion distinct from automation.
- Make unresolved threads impossible to miss.
- Preserve patch-set history.
- Support stacked-review workflows.
- Prioritize low-latency diff navigation.

Do not:

- Hide important review state behind menus.
- Inject promotional AI text.
- Collapse human and bot comments into the same visual hierarchy.
- Add extra clicks to core review paths.
- Make checkout/rebase instructions ambiguous.
- Make the user infer whether a check is stale.

### Search

Search should expose mode, scope, and freshness.

Search modes should be distinct:

- Code.
- Symbols.
- Issues.
- Pull requests.
- Discussions.
- Packages.
- Docs.
- Users/organizations.

Do:

- Show which branches/scopes are indexed.
- Show last indexed time.
- Offer regex and exact search.
- Offer semantic reranking as optional, not magical.
- Support saved multi-repo searches.

Do not:

- Silently switch between search engines.
- Hide rate limits.
- Index only default branches without making that clear.
- Treat search as a secondary feature.

### Administration

Admin UX should make policy legible.

Do:

- Show inherited policies.
- Show effective permissions.
- Explain why a user can or cannot perform an action.
- Make guest/external access easy to audit.
- Make token scopes understandable.
- Provide policy simulation or dry-run tools.

Do not:

- Split critical settings across unrelated pages.
- Require admins to reverse-engineer access.
- Hide billing or security consequences behind vague labels.

---

## Architecture Guidance

Use a cell-based architecture with isolated planes.

High-level components:

- Web, CLI, IDE, and mobile clients.
- Edge/CDN/API gateway.
- Repository gateway.
- Immutable Git object storage.
- Ref, branch, and permission service.
- Metadata database per cell.
- PR and review service.
- Diff and patch-set cache.
- Append-only event log.
- Webhook delivery and replay.
- CI orchestration.
- Hosted and self-hosted runner pools.
- Package and artifact service.
- Search ingest pipeline.
- Lexical, symbol, regex, and semantic search indexes.
- Identity, SSO, SCIM, policy, and audit services.
- Observability, tracing, SLOs, status, and abuse controls.

### Cell-Based Model

Each cell should own a bounded tenant or regional slice and include:

- Repository gateway.
- Metadata database.
- Search ingest.
- Review cache.
- CI coordinator.
- Webhook workers.
- Relevant observability and abuse controls.

Benefits:

- Reduced blast radius.
- Component-level status.
- Easier tenant isolation.
- More legible incidents.
- Better regional scaling.

### Data Consistency

Git objects are mostly immutable and can be broadly replicated.

Refs, permissions, review decisions, identity state, billing state, and audit state require stronger consistency and careful failure handling.

### Eventing

Events should be durable, replayable, and inspectable.

Requirements:

- Append-only event log.
- Idempotency keys.
- Dead-letter queues.
- Per-delivery payload inspection.
- Replay from UI and API.
- Webhook delivery history.
- Retention windows suitable for incident recovery.

The model should be “durable event distribution,” not best-effort callbacks.

### Search Architecture

Use a hybrid pipeline:

- Trigram/lexical retrieval for literal matching.
- Symbol index for definitions and references.
- Regex support for maintainers.
- Semantic reranking as an optional layer.
- Freshness indicators.
- Explicit indexed scope.

Never let users wonder whether search results are complete.

### Operational Priorities

The forge core has the strictest reliability target.

Suggested SLO priority order:

1. Repository read/write.
2. PR viewing and review actions.
3. Permission checks and identity.
4. Webhooks and status checks.
5. Search.
6. Issues/discussions.
7. CI orchestration.
8. Package and artifact publication.
9. Discovery/social features.

A delayed build is less damaging than a broken merge, inaccessible repository, incorrect permission, or missing review state.

---

## Migration Strategy

Migration is product, not documentation.

Support two adoption modes:

### Incremental Adoption

- Mirror repositories from GitHub.
- Sync Git LFS.
- Import issues, PRs, labels, comments, releases, and wiki data.
- Run CLOVE CI in parallel.
- Mirror packages.
- Dual-deliver webhooks.
- Let teams move repository by repository.
- Preserve GitHub as social upstream during transition if needed.

### Enterprise Cutover

1. Inventory and discovery.
2. Repository mirror and LFS sync.
3. Metadata import: issues, PRs, releases, wiki, labels, comments.
4. Identity mapping: users, teams, SSO, guests, SCIM.
5. Dual-run phase: CI, webhooks, package mirror, status checks.
6. Freeze window and delta sync.
7. Cutover: URLs, webhooks, runners, package endpoints.
8. Redirects, audit validation, rollback window, decommission.

Identity mapping is the hardest part. Treat it as a first-class migration workflow, not a support edge case.

---

## Business Model Guidance

CLOVE should align pricing with the trust thesis.

Recommended model:

- Seat pricing for collaboration, review, admin, and governance.
- Transparent metering for expensive compute/storage surfaces.
- Managed cloud as the default growth engine.
- Early self-hosted or single-tenant architecture path.
- OSS/public project support subsidized by sponsors, partnerships, or enterprise revenue.

Pricing must include:

- Included quotas.
- Hard budget caps.
- Stop-on-budget controls.
- Preflight cost estimates.
- Annual rate cards.
- Price-protection windows.
- Clear compute/storage separation.
- No surprise billing surfaces.

Do not monetize through confusion.

---

## Open Source and Community Strategy

CLOVE should be credible for open-source maintainers without relying only on donation-funded operations.

Important features:

- Public project discovery.
- Maintainer health signals.
- Good import/export.
- No ads in collaboration surfaces.
- Transparent moderation.
- Abuse-resistant onboarding.
- Project sponsorship or funding integrations.
- Package and CI abuse controls.
- Clear policy due process.
- Public incident reports.

Community trust depends on both values and operational capacity.

---

## Anti-Patterns to Avoid

### Cosmetic GitHub Clone

Copying GitHub’s layout without solving review state, migration, search, trust, and admin legibility is not enough.

### One-Tribe Product

Do not become useful only for one narrow community, such as email-first maintainers, self-hosting purists, or Atlassian-heavy teams.

### Underfunded Operations

Running a public forge requires serious SRE, abuse response, DDoS mitigation, storage durability, spam handling, and incident communication.

### Single Failure Domain

Do not let repository hosting, CI, packages, identity, billing, and search all fail together.

### Surprise Monetization

Do not create unpredictable billing surfaces. Cost volatility is a trust failure.

### AI Authorship Erosion

Do not let AI speak for users, modify public collaboration surfaces without explicit consent, or train on private data by default.

### Accessibility as Backlog

Accessibility and internationalization are not polish. They are part of review correctness and global usability.

---

## Engineering Defaults

When building CLOVE, prefer:

- Explicit state over inferred state.
- Durable logs over ephemeral callbacks.
- Exportable data over lock-in.
- Hard boundaries over hidden coupling.
- Clear permissions over clever abstractions.
- Fast reviewer workflows over decorative UI.
- Visible freshness over magical search.
- Budget caps over metering surprises.
- Optional AI over platform coercion.
- Accessible defaults over retrofitted compliance.

---

## Tone and Brand

CLOVE should feel:

- Dependable.
- Fast.
- Calm.
- Legible.
- Professional.
- Developer-respecting.
- Admin-friendly.
- Anti-surprise.
- Portable.
- Trustworthy.

Avoid:

- Growth-hacky copy.
- AI hype.
- Dark patterns.
- Enterprise jargon without product clarity.
- Social-network clutter.
- Ads in collaboration surfaces.
- “One platform to rule them all” positioning.

Suggested positioning:

> CLOVE is a review-first, Git-compatible forge built for teams that need dependable code collaboration, predictable delivery, enterprise-legible administration, and a clean exit path.

Alternative short version:

> GitHub-compatible collaboration without GitHub-style lock-in, surprise costs, or AI ambiguity.

---

## Agent Instructions for Future Work

When working on this project:

1. Prioritize repository, review, search, identity, migration, and audit flows before peripheral features.
2. Treat PR review UX as the primary workflow.
3. Keep AI optional, visually distinct, and user-initiated.
4. Design every expensive subsystem with budget controls.
5. Keep CI/CD and package registries native but separable.
6. Build import/export and migration fidelity early.
7. Make admin permissions explainable.
8. Design for accessibility and internationalization from the start.
9. Avoid features that increase lock-in without increasing user trust.
10. Document operational boundaries and failure behavior clearly.
11. Use GitHub compatibility strategically, but do not blindly copy GitHub.
12. Optimize for serious teams, maintainers, and enterprise admins rather than casual repository hosting alone.

When unsure, ask:

- Does this make code review safer or clearer?
- Does this reduce migration fear?
- Does this improve operational trust?
- Does this make cost more predictable?
- Does this preserve user agency?
- Does this make permissions easier to understand?
- Does this avoid bundling unrelated failure domains?
- Does this respect human authorship?

If the answer is no, the feature probably does not belong in the core product yet.
