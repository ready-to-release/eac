# Compliance-as-Code Basics

**Status:** Placeholder - Content coming soon

**Prerequisites:** [Your First Specification](../getting-started/first-specification.md), understanding of compliance requirements

## Planned Content

This tutorial teaches you how to implement compliance automation by tagging specifications with control references, generating risk assessments, and producing audit evidence.

### What You'll Learn

- Tag specifications with `@control:` references to OSCAL catalogs
- Map requirements to compliance frameworks (21 CFR Part 11, GAMP 5, etc.)
- Generate risk assessments: `r2r create risk-assess`
- Validate control tags: `r2r validate control-tags`
- Generate audit evidence from CI/CD pipeline
- Understand OSCAL (Open Security Controls Assessment Language)

### Tutorial Structure

1. **Understanding compliance-as-code**
   - Traditional compliance: manual documentation, periodic audits
   - Compliance-as-code: automated validation, continuous evidence
   - Benefits: shift-left, traceability, reduced audit prep time

2. **OSCAL catalogs and controls**
   - OSCAL standard overview
   - Control catalogs (NIST 800-53, FDA controls, etc.)
   - Control structure: ID, title, parameters, guidance
   - Example: AC-1 (Access Control Policy)

3. **Tagging specifications**
   - Add control references to features
   - Tag format: `@control:AC-1` or `@control:AC-1(a)`
   - Multiple controls: `@control:AC-1 @control:AC-2`
   - Compliance tags: `@cfr21part11`, `@gamp5`

4. **Example: Access Control Specification**

   ```gherkin
   @L2 @ov @deps:go @control:AC-1 @control:AC-2 @cfr21part11
   Feature: user-auth_access-control

     As a security officer
     I want to enforce access controls
     So that only authorized users can access the system

     Rule: Users must authenticate before accessing protected resources

       Scenario: Deny access without authentication
         Given I am not authenticated
         When I attempt to access protected resource "/api/data"
         Then I should receive a 401 Unauthorized response
   ```

5. **Validating control tags**
   - Check tags against catalog: `r2r validate control-tags`
   - Ensure all referenced controls exist
   - Verify tag format correctness

6. **Generating risk assessments**
   - Create assessment: `r2r create risk-assess`
   - Produces OSCAL assessment-results document
   - Maps specifications to controls
   - Includes test evidence

7. **Understanding audit evidence**
   - Test results as evidence
   - Build artifacts and signatures
   - CI/CD logs and approvals
   - Traceability matrix

8. **Compliance frameworks**
   - 21 CFR Part 11 (FDA electronic records)
   - GAMP 5 (pharmaceutical GxP)
   - NIST 800-53 (federal security controls)
   - Custom organizational controls

### Compliance Workflow

The tutorial demonstrates a complete compliance workflow:

```bash
# 1. Tag specifications with control references
vim specs/user-auth/access-control/specification.feature
# Add @control:AC-1 @control:AC-2

# 2. Validate control tags
r2r validate control-tags
# ✓ All control references valid

# 3. Run tests (generates evidence)
r2r test user-auth --suite acceptance

# 4. Generate risk assessment
r2r create risk-assess
# Creates .r2r/oscal/assessment-results.json

# 5. Review assessment
cat .r2r/oscal/assessment-results.json
```

### Key Concepts Covered

- Compliance-as-code principles
- OSCAL standard and catalogs
- Control tagging in specifications
- Risk assessment generation
- Evidence collection and traceability
- Mapping to regulatory frameworks

### OSCAL Document Types

- **Catalog**: Control definitions (e.g., NIST 800-53)
- **Profile**: Control baseline (selected controls)
- **Assessment Plan**: How controls will be tested
- **Assessment Results**: Test results and findings
- **POA&M**: Plan of Action & Milestones

### Benefits of Compliance-as-Code

- **Continuous validation**: Every commit generates evidence
- **Automated traceability**: Requirements → Controls → Tests → Evidence
- **Reduced audit prep**: Evidence collected automatically
- **Version controlled**: Compliance artifacts in git
- **Efficient audits**: Auditors review automated evidence

### Real-World Example

Scenario: FDA-regulated medical device software

1. Tag specifications with 21 CFR Part 11 controls
2. Run tests in CI/CD (generates evidence)
3. Generate assessment results
4. Export traceability matrix
5. Submit to auditors with automated evidence
6. Result: 80% reduction in audit prep time

### Best Practices

- Tag specifications at feature level
- Validate control tags in pre-commit hooks
- Generate assessments in CI/CD
- Version control OSCAL documents
- Review control coverage regularly
- Keep catalogs up to date

### Next Steps

After completing this tutorial, you'll understand compliance automation basics. Continue to [CI/CD Integration](./ci-cd-integration.md) to learn how to automate compliance validation in your pipeline.
