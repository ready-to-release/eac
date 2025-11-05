# Testing Taxonomy (L0-L4)

This taxonomy defines test categories based on execution environment and scope.

## Test Levels

| Level | Name                   | Shift Direction | Execution Environment | Scope                | External Dependencies                     | Determinism | Domain Coherency |
|-------|------------------------|-----------------|-----------------------|----------------------|-------------------------------------------|-------------|------------------|
| L0    | Unit Tests             | LEFT            | devbox or agent       | Source and binary    | All replaced with test doubles            | Highest     | Lowest           |
| L1    | Unit Tests             | LEFT            | devbox or agent       | Source and binary    | All replaced with test doubles            | Highest     | Lowest           |
| L2    | Emulated System Tests  | LEFT            | devbox or agent       | Deployable artifacts | All replaced with test doubles            | High        | High             |
| L3    | In-Situ Vertical Tests | LEFT            | PLTE                  | Deployed system      | All replaced with test doubles            | Moderate    | High             |
| L4    | Testing in Production  | RIGHT           | Production            | Deployed system      | All production, may use live test doubles | High        | Highest          |

## Out-of-Category for Automated QA

| Level          | Name                  | Shift Direction | Execution Environment      | Scope           | External Dependencies                        | Determinism | Domain Coherency |
|----------------|-----------------------|-----------------|----------------------------|-----------------|----------------------------------------------|-------------|------------------|
| Horizontal E2E | Horizontal End-to-End | non-shifted     | Shared testing environment | Deployed system | Tied up to non-production "test" deployments | Lowest      | High             |

Old-school horizontal pre-production environments where multiple teams' pre-prod services are linked together are highly fragile and non-deterministic. This taxonomy explicitly advocates shifting LEFT (to L0-L3) and RIGHT (to L4) to avoid this pattern.

If you decide to break this fundamental rule, all you need to do is to connect your L3 PLTE to something not a test double and you have automated tests running in L3 with external dependencies. This is not a technical constraint, its an inherent constrant in the nature of Horizontal E2E: Noone can control the variables, so you are not performing verification, you are playing games.

It is recommended to have a horizontally connected e2e environment and calling it Demo - and use it for explorative testing and demos.

It is NOT recommended to tie this environment up to any automated tests, other than smoke tests like "able to deploy, site starts" etc.

## Taxonomy Constraints

The taxonomy constrains two key aspects:

- Execution requirements (what binaries, tooling, and configuration are needed)
- Test scope (vertical vs horizontal boundaries, and test double usage)

## Determinism vs Coherency Trade-off

Lower Lx values provide higher determinism but lower domain coherency, while higher Lx values provide lower determinism but higher domain coherency.
