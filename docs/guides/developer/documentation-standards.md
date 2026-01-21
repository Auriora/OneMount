# Documentation Standards for OneMount

**Version**: 1.0  
**Last Updated**: 2025-01-21  
**Owner**: Development Team

## Purpose

This document defines the standards and processes for maintaining documentation in the OneMount project. Following these standards ensures that documentation remains accurate, current, and useful for all stakeholders.

## Documentation Types

### 1. Architecture Documentation

**Location**: `docs/2-architecture/`

**Purpose**: Describes the high-level structure of the system, component interactions, and architectural decisions.

**Update Frequency**: Within 1 week of structural changes

**Required Content**:
- Component diagrams
- Component interactions
- Interface descriptions
- Architectural decision records (ADRs)
- System context diagrams

**Standards**:
- Use PlantUML for all diagrams
- Include both diagram source and rendered images
- Document rationale for architectural decisions
- Keep diagrams up-to-date with implementation

### 2. Design Documentation

**Location**: `docs/2-architecture/`

**Purpose**: Describes the detailed design of components, data models, and algorithms.

**Update Frequency**: Within 1 week of data model or design changes

**Required Content**:
- Class diagrams
- Sequence diagrams
- Data model descriptions
- Design patterns used
- API signatures

**Standards**:
- Use PlantUML for all diagrams
- Document design patterns explicitly
- Include rationale for design decisions
- Keep API signatures current

### 3. API Documentation (Godoc)

**Location**: Inline in source code

**Purpose**: Documents the public API of the system for developers.

**Update Frequency**: In same commit as code changes

**Required Content**:
- Function/method description
- Parameter descriptions
- Return value descriptions
- Error conditions
- Usage examples (for complex APIs)

**Standards**:
```go
// FunctionName performs [action] on [object].
//
// This function [detailed description of behavior].
//
// Parameters:
//   - param1: Description of param1
//   - param2: Description of param2
//
// Returns:
//   - returnValue: Description of return value
//   - error: Description of error conditions
//
// Example:
//   result, err := FunctionName(arg1, arg2)
//   if err != nil {
//       // Handle error
//   }
//   // Use result
func FunctionName(param1 Type1, param2 Type2) (ReturnType, error) {
    // Implementation
}
```

### 4. User Documentation

**Location**: `docs/guides/user/`

**Purpose**: Guides end users on how to use the system.

**Update Frequency**: Within 2 weeks of user-facing changes

**Required Content**:
- Installation instructions
- Configuration guides
- Usage examples
- Troubleshooting guides
- FAQ

**Standards**:
- Use clear, non-technical language
- Include screenshots/examples
- Provide step-by-step instructions
- Keep troubleshooting guides current

### 5. Developer Documentation

**Location**: `docs/guides/developer/`

**Purpose**: Guides developers on how to contribute to the project.

**Update Frequency**: As needed

**Required Content**:
- Development environment setup
- Build instructions
- Testing guidelines
- Contribution guidelines
- Code review checklist

**Standards**:
- Assume minimal prior knowledge
- Provide complete setup instructions
- Include troubleshooting for common issues
- Keep build instructions current

## Documentation Update Process

### 1. During Development

**When making code changes**:

1. **Identify Documentation Impact**
   - Does this change affect architecture?
   - Does this change affect design?
   - Does this change affect public APIs?
   - Does this change affect user-facing behavior?

2. **Update Documentation**
   - Update architecture docs for structural changes
   - Update design docs for data model changes
   - Update godoc comments for API changes
   - Update user docs for behavior changes

3. **Create ADRs**
   - Create ADR for significant architectural decisions
   - Document rationale and consequences
   - Link ADR to related code changes

### 2. Pull Request Process

**PR Template Checklist**:

```markdown
## Documentation Updates

- [ ] Architecture documentation updated (if applicable)
- [ ] Design documentation updated (if applicable)
- [ ] API documentation (godoc) updated (if applicable)
- [ ] User documentation updated (if applicable)
- [ ] ADR created (if applicable)
- [ ] N/A - No documentation changes required

## Documentation Changes

[Describe what documentation was updated and why]
```

**Code Review Checklist**:

- [ ] All public APIs have godoc comments
- [ ] Godoc comments are accurate and complete
- [ ] Architecture docs reflect structural changes
- [ ] Design docs reflect data model changes
- [ ] User docs reflect behavior changes
- [ ] ADRs created for significant decisions
- [ ] Diagrams are up-to-date

### 3. Regular Reviews

**Weekly Documentation Review**:
- Review recent code changes for documentation gaps
- Update documentation backlog
- Prioritize documentation tasks
- Assign documentation work

**Quarterly Documentation Audit**:
- Comprehensive review of all documentation
- Verify alignment with implementation
- Update outdated sections
- Archive deprecated documentation
- Generate documentation metrics

## Documentation Ownership

### Component Owners

Each major component has a designated documentation owner responsible for:
- Keeping documentation current
- Reviewing documentation changes
- Conducting regular documentation audits

| Component | Owner | Documentation Location |
|-----------|-------|----------------------|
| Filesystem | TBD | `docs/2-architecture/`, `internal/fs/` |
| Graph API | TBD | `docs/2-architecture/`, `internal/graph/` |
| Authentication | TBD | `docs/2-architecture/`, `internal/graph/` |
| Cache Management | TBD | `docs/2-architecture/`, `internal/fs/` |
| UI Components | TBD | `docs/2-architecture/`, `internal/ui/` |
| Metadata Store | TBD | `docs/2-architecture/`, `internal/metadata/` |
| State Management | TBD | `docs/2-architecture/`, `internal/fs/` |

## Automated Documentation Checks

### 1. Godoc Linting

**Tool**: `golangci-lint` with `godoc` linter enabled

**Checks**:
- All exported functions have godoc comments
- Godoc comments start with function name
- Godoc comments are complete sentences

**Configuration**:
```yaml
linters:
  enable:
    - godoc

linters-settings:
  godoc:
    check-all: true
```

### 2. Link Checking

**Tool**: `markdown-link-check`

**Checks**:
- All links in markdown files are valid
- No broken internal links
- No broken external links

**Usage**:
```bash
find docs -name "*.md" -exec markdown-link-check {} \;
```

### 3. Spell Checking

**Tool**: `aspell` or `codespell`

**Checks**:
- No spelling errors in documentation
- Technical terms in whitelist

**Usage**:
```bash
codespell docs/
```

### 4. Diagram Validation

**Tool**: `plantuml`

**Checks**:
- All PlantUML diagrams compile successfully
- Rendered images are up-to-date

**Usage**:
```bash
find docs -name "*.puml" -exec plantuml -checkonly {} \;
```

## Documentation Tools

### 1. Diagram Generation

**Tool**: PlantUML

**Usage**:
```bash
# Generate all diagrams
find docs -name "*.puml" -exec plantuml {} \;

# Generate specific diagram
plantuml docs/2-architecture/resources/component-diagram.puml
```

**Standards**:
- Store source in `.puml` files
- Generate PNG images for documentation
- Commit both source and rendered images

### 2. API Documentation

**Tool**: `godoc`

**Usage**:
```bash
# Generate API documentation
godoc -http=:6060

# View at http://localhost:6060/pkg/github.com/jstaf/onedriver/internal/
```

**Standards**:
- All public APIs must have godoc comments
- Include examples for complex APIs
- Link to related documentation

### 3. Documentation Hosting

**GitHub Pages**: User documentation  
**godoc.org**: API documentation  
**Internal Wiki**: Development documentation

## Documentation Metrics

### Success Metrics

- [ ] 100% of public APIs have godoc comments
- [ ] 100% of PRs include documentation updates (when applicable)
- [ ] 0 broken links in documentation
- [ ] 0 spelling errors in documentation
- [ ] Weekly documentation review completed
- [ ] Quarterly documentation audit completed

### Tracking

**Weekly Metrics**:
- Number of PRs with documentation updates
- Number of documentation gaps identified
- Number of documentation tasks completed

**Quarterly Metrics**:
- Percentage of public APIs with godoc comments
- Number of broken links
- Number of spelling errors
- Documentation coverage by component

## Architectural Decision Records (ADRs)

### Purpose

ADRs document significant architectural decisions, their context, and consequences.

### When to Create an ADR

Create an ADR when:
- Making a significant architectural decision
- Changing an existing architectural decision
- Introducing a new design pattern
- Deviating from established patterns
- Making a decision with long-term impact

### ADR Template

**Location**: `docs/2-architecture/decisions/`

**Filename**: `ADR-XXX-title-in-kebab-case.md`

**Template**:
```markdown
# ADR-XXX: [Title]

## Status

[Proposed | Accepted | Deprecated | Superseded by ADR-YYY]

## Context

[What is the issue that we're seeing that is motivating this decision or change?]

## Decision

[What is the change that we're proposing and/or doing?]

## Consequences

### Positive

[What becomes easier because of this change?]

### Negative

[What becomes more difficult because of this change?]

### Neutral

[What is neither easier nor more difficult?]

## Alternatives Considered

[What other options were considered and why were they rejected?]

## References

[Links to related documentation, issues, PRs, etc.]
```

### ADR Lifecycle

1. **Proposed**: ADR is created and under discussion
2. **Accepted**: ADR is approved and implemented
3. **Deprecated**: ADR is no longer recommended but still in use
4. **Superseded**: ADR is replaced by a newer ADR

## Documentation Style Guide

### General Principles

1. **Clarity**: Write clearly and concisely
2. **Accuracy**: Ensure documentation matches implementation
3. **Completeness**: Include all necessary information
4. **Consistency**: Follow established patterns and conventions
5. **Maintainability**: Make documentation easy to update

### Writing Style

- Use active voice
- Use present tense
- Use second person ("you") for user documentation
- Use third person for technical documentation
- Avoid jargon unless necessary
- Define technical terms on first use

### Formatting

- Use Markdown for all documentation
- Use code blocks for code examples
- Use tables for structured data
- Use lists for sequential or related items
- Use headings to organize content

### Code Examples

- Include complete, working examples
- Use realistic scenarios
- Include error handling
- Add comments to explain non-obvious code
- Test examples to ensure they work

## Training and Onboarding

### New Developer Onboarding

1. **Documentation Overview**
   - Review this documentation standards guide
   - Review existing documentation structure
   - Understand documentation ownership

2. **Documentation Tools**
   - Setup PlantUML
   - Setup godoc
   - Setup linting tools

3. **Practice**
   - Update documentation for a small change
   - Create an ADR for a hypothetical decision
   - Review documentation in a PR

### Ongoing Training

- Quarterly documentation workshops
- Share documentation best practices
- Review documentation metrics
- Celebrate documentation improvements

## Conclusion

Following these documentation standards ensures that OneMount documentation remains accurate, current, and useful for all stakeholders. Regular reviews and automated checks help maintain documentation quality over time.

**Questions or Suggestions?**

Contact the documentation owner or open an issue in the project repository.
