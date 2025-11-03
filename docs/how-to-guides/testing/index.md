# Testing How-to Guides

Step-by-step guides for testing tasks.

---

## Overview

These guides provide task-oriented instructions for working with the three-layer testing approach (ATDD, BDD, TDD).

---

## Available Guides

### [Setup Gauge](setup-gauge.md)

Install and configure Gauge for ATDD acceptance testing.

**When to use**: Before creating your first acceptance.spec file

**What you'll learn**:

- Install Gauge on your system
- Install the Go plugin for Gauge
- Verify installation
- Configure project for Gauge

### [Setup Godog](setup-godog.md)

Install and configure Godog for BDD behavior testing.

**When to use**: Before creating your first behavior.feature file

**What you'll learn**:

- Install Godog
- Configure Godog for your project
- Verify installation
- Create initial test harness

### [Create Feature Spec](create-feature-spec.md)

Create acceptance.spec and behavior.feature files for a new feature.

**When to use**: Starting work on a new feature

**What you'll learn**:

- Create feature directory structure
- Write acceptance.spec file
- Write behavior.feature file
- Link files with Feature ID
- Implement step definitions

### [Run Example Mapping Workshop](run-example-mapping.md)

Facilitate an Example Mapping workshop to discover requirements.

**When to use**: Before implementing any new feature

**What you'll learn**:

- Prepare for the workshop
- Run the 15-25 minute session
- Use colored cards effectively
- Convert cards to test files
- Handle questions and blockers

### [Run Tests](run-tests.md)

Execute tests at all three layers (ATDD, BDD, TDD).

**When to use**: During development and in CI/CD

**What you'll learn**:

- Run all tests
- Run specific feature tests
- Run tests by tags
- Generate reports
- Integrate with CI/CD

---

## Quick Start

**New to testing in this project?** Follow this sequence:

1. [Setup Gauge](setup-gauge.md) - Install ATDD tool
2. [Setup Godog](setup-godog.md) - Install BDD tool
3. [Run Example Mapping Workshop](run-example-mapping.md) - Discover requirements
4. [Create Feature Spec](create-feature-spec.md) - Create test files
5. [Run Tests](run-tests.md) - Validate implementation

---

**Want to understand WHY?** See [Testing Explanation](../../explanation/testing/index.md)

**Need technical reference?** See [Testing Reference](../../reference/testing/index.md)
