# Core Concepts

Foundational concepts that underpin the Continuous Delivery Model. These articles establish the building blocks you need to understand before diving into the 12-stage model and implementation practices.

**Start here** if you're new to the CD Model.

---

## [Unit of Flow](unit-of-flow.md)

Understanding the four interconnected components that enable Continuous Delivery.

The Unit of Flow provides a mental model for how Continuous Delivery works by breaking it down into four discrete components:

1. **Trunk**: The version-controlled timeline where code lives (Git repository)
2. **Deployable Unit**: The discrete body of work that gets built and deployed
3. **Deployment Pipeline**: The automated process that validates and delivers changes (12 stages)
4. **Live**: The runtime environment where software serves users

**Key topics:**

- How the four components relate to each other
- Polyrepo vs monorepo patterns
- Integration with CD Model stages 1-12
- Common architectural patterns

**Read this article to understand**: The big picture of how Continuous Delivery components work together.

---

## [Deployable Units](deployable-units.md)

Understanding what gets built, versioned, and deployed through the CD Model.

A Deployable Unit is the fundamental building block of Continuous Delivery - the discrete body of work that is built, tested, and delivered as a cohesive whole.

**Key topics:**

- Definition and characteristics
- Two types: Runtime Systems (services, apps) vs Versioned Components (libraries, containers)
- Six versioning strategies: Implicit, CalVer, Release Number, SemVer, API versioning
- Immutable artifacts and why they matter
- Dependency management (internal and external)
- Choosing the right granularity and boundaries

**Read this article to understand**: What you're actually building and deploying, how to version it, and how to manage dependencies.

---

## Why These Concepts Matter

Understanding these core concepts is essential because:

**Unit of Flow** provides the mental model for:

- How code flows from development to production
- Why trunk-based development matters
- How deployable units relate to repositories
- Where the deployment pipeline fits in

**Deployable Units** establish:

- What gets versioned and released
- How to structure your repositories
- When to create new deployable units
- How to manage dependencies between units

These concepts are referenced throughout the CD Model documentation and are prerequisite knowledge for understanding implementation patterns.

---

## Next Steps

Once you understand these core concepts, you're ready to explore:

- **[CD Model](../cd-model/index.md)**: The complete 12-stage framework
- **[Workflow](../workflow/index.md)**: Trunk-based development and branching strategies
- **[Testing](../testing/index.md)**: Testing strategy across all stages
- **[Architecture](../architecture/index.md)**: Environment and repository patterns
- **[Security](../security/index.md)**: Security integration throughout the pipeline

Or return to the **[Continuous Delivery Overview](../index.md)** for the complete navigation.
