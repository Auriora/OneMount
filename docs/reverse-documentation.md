# Reverse Documentation Approach and Prompts for OneDriver

## Overview

This document outlines a structured four-phase process to reverse document the OneDriver project, along with industry-standard templates and AI prompt examples to automate each phase. The recommended phases are:

1. **Planning & Setup** – Define scope, gather existing code and artifacts, and select documentation standards.
2. **Reverse Analysis** – Perform static and dynamic analysis to extract modules, dependencies, and behavior.
3. **Documentation Drafting** – Populate templates for Requirements Specification, Architecture Document, Design Specification, Use Cases, and Test Cases.
4. **Review & Validation** – Conduct peer reviews, refine with automated feedback, and validate against the codebase.

For artifact templates, I recommend:

- **Software Requirements Specification (SRS)** based on ISO/IEC/IEEE 29148 ([ISO/IEC/IEEE 29148 Requirements Specification Templates](https://www.reqview.com/doc/iso-iec-ieee-29148-templates/?utm_source=chatgpt.com)).
- **Software Architecture Document** using the "Views and Beyond" approach ([Example: Software Architecture Document](https://www.ecs.csun.edu/~rlingard/COMP684/Example2SoftArch.htm?utm_source=chatgpt.com)).
- **Use Case** descriptions and UML diagrams ([UML Use Case Diagram Tutorial - Lucidchart](https://www.lucidchart.com/pages/uml-use-case-diagram?utm_source=chatgpt.com)).
- **Test Case** templates for consistency and completeness ([Test Case Template with Examples: Free Excel & Word Sample for ...](https://katalon.com/resources-center/blog/test-case-template-examples?utm_source=chatgpt.com)).

---

## Proposed Process

### Phase 1: Planning & Setup

1. **Define Objectives & Scope**  
   Clarify which features, modules, and use cases in OneDriver need documentation.
2. **Gather Existing Artifacts**  
   Collect the code repository, any existing READMEs, comments, and test suites ([8 steps to the reverse-engineering process - Control Design](https://www.controldesign.com/design/development-platforms/article/55252541/8-steps-to-the-reverse-engineering-process?utm_source=chatgpt.com)).
3. **Select Standards & Templates**
    - Requirements: ISO/IEC/IEEE 29148 SRS ([ISO/IEC/IEEE 29148 Requirements Specification Templates](https://www.reqview.com/doc/iso-iec-ieee-29148-templates/?utm_source=chatgpt.com))
    - Architecture: IEEE/SEI Views & Beyond ([Software Architecture Documentation Template](https://wiki.sei.cmu.edu/confluence/display/SAD/Software%2BArchitecture%2BDocumentation%2BTemplate?utm_source=chatgpt.com))
    - Test Cases: Smartsheet or Katalon templates ([Free Test Case Templates | Smartsheet](https://www.smartsheet.com/test-case-templates-examples?srsltid=AfmBOoqbbG8DOp0t26QcdJYO4e133y_HLkJrFSSUg8aQVJwVflv1_jCF&utm_source=chatgpt.com), [Test Case Template with Examples: Free Excel & Word Sample for ...](https://katalon.com/resources-center/blog/test-case-template-examples?utm_source=chatgpt.com))
    - Use Cases: UML diagrams from Lucidchart/Visual Paradigm ([UML Use Case Diagram Tutorial - Lucidchart](https://www.lucidchart.com/pages/uml-use-case-diagram?utm_source=chatgpt.com), [Use Case Diagram Templates - Visual Paradigm Online](https://online.visual-paradigm.com/diagrams/templates/use-case-diagram/?utm_source=chatgpt.com))

#### Phase 1 Prompts

**Prompt 1: List All Features and Modules**
```
You are a product discovery specialist. Analyze the root and src directories of the OneDriver repository and output a JSON array where each element contains:
- moduleName: string
- description: one-sentence summary
- fileCount: number of files
- dependencies: list of imported modules
Only include modules that define at least one function or class.
```  
([ChatGPT prompt for generating app feature list - Promptmatic](https://promptmatic.ai/prompts/chatgpt/generate-app-feature-list?utm_source=chatgpt.com))  

**Prompt 2: Identify Stakeholders and Goals**  
```
You are a business analyst. From the OneDriver README and code comments, extract all stakeholders, their roles, and the primary goals they wish to achieve. Present as a markdown table with columns: Stakeholder, Role, Goal.
```  
([30 ChatGPT Prompts for Software Development Engineers](https://kms-technology.com/emerging-technologies/ai/30-best-chatgpt-prompts-for-software-engineers.html?utm_source=chatgpt.com))  

---

### Phase 2: Reverse Analysis

1. **Static Code Analysis**  
   Extract module hierarchy, class/function lists, and dependency graphs.
2. **Dynamic Tracing & Logging**  
   Instrument the code to capture execution paths and data flows.
3. **Disassembly & Behavior Mapping**  
   Work from high-level modules down to individual algorithms to understand functionality ([8 steps to the reverse-engineering process - Control Design](https://www.controldesign.com/design/development-platforms/article/55252541/8-steps-to-the-reverse-engineering-process?utm_source=chatgpt.com)).
4. **Extract Key Artifacts**
    - Module descriptions
    - Data models
    - Protocols or APIs

#### Phase 2 Prompts

**Prompt 3: Static Code Analysis**  
```
You are a static analysis specialist. Perform static code analysis on the OneDriver codebase to list all classes and functions along with their cyclomatic complexity. Output a CSV with columns: Name, Type (class/function), FilePath, Complexity.
```  
([Prompt Security Launches Static Analysis Security Testing for AI ...](https://www.prompt.security/press/prompt-security-launches-static-analysis-security-testing-for-ai-generated-code?utm_source=chatgpt.com))  

**Prompt 4: Dynamic Tracing Instrumentation**  
```
You are a software instrumentation engineer. Inject logging into all public methods in the core module to record entry, exit, and parameter values at runtime. Provide a unified patch or diff using the existing logging framework.
```  
([Get granular LLM observability by instrumenting your LLM chains](https://www.datadoghq.com/blog/llm-observability-chain-tracing/?utm_source=chatgpt.com))  

**Prompt 5: Behavior Mapping**  
```
You are a reverse engineering expert. Execute the primary workflows in OneDriver (e.g., file upload, download, conflict resolution) and log the sequence of invoked functions. Present a sequence diagram in PlantUML syntax illustrating these flows.
```  
([Get granular LLM observability by instrumenting your LLM chains](https://www.datadoghq.com/blog/llm-observability-chain-tracing/?utm_source=chatgpt.com))  

---

### Phase 3: Documentation Drafting

#### 3.1 Software Requirements Specification (SRS)

Use an ISO/IEC/IEEE 29148–compliant template such as the one provided by ReqView ([ISO/IEC/IEEE 29148 Requirements Specification Templates](https://www.reqview.com/doc/iso-iec-ieee-29148-templates/?utm_source=chatgpt.com), [Example Software Requirements Specification (SRS) - ReqView](https://www.reqview.com/doc/iso-iec-ieee-29148-srs-example/?utm_source=chatgpt.com)):

- **1 Introduction**
    - Purpose
    - Scope
    - Definitions, Acronyms, Abbreviations
- **2 Overall Description**
    - Product Perspective
    - User Characteristics
    - Constraints
- **3 Specific Requirements**
    - Functional Requirements (numbered)
    - Nonfunctional Requirements (performance, security, usability)
- **4 Appendices**
    - Glossary
    - References

#### 3.2 Use Cases

Adopt a standardized use case template with accompanying UML diagrams ([UML Use Case Diagram Tutorial - Lucidchart](https://www.lucidchart.com/pages/uml-use-case-diagram?utm_source=chatgpt.com), [Use Case Diagram – Unified Modeling Language (UML)](https://www.geeksforgeeks.org/use-case-diagram/?utm_source=chatgpt.com)):

| Field             | Description                        |
|-------------------|------------------------------------|
| Use Case ID       | UC-XX                              |
| Name              | Brief, descriptive title           |
| Actors            | List of users/external systems     |
| Preconditions     | What must be true before execution |
| Postconditions    | State after successful completion  |
| Main Flow         | Step-by-step primary scenario      |
| Alternative Flows | Variations, error conditions       |

#### 3.3 Software Architecture Document (SAD)

Follow the "Views and Beyond" approach with multiple architectural views ([Example: Software Architecture Document](https://www.ecs.csun.edu/~rlingard/COMP684/Example2SoftArch.htm?utm_source=chatgpt.com), [Software Architecture Documentation Template](https://wiki.sei.cmu.edu/confluence/display/SAD/Software%2BArchitecture%2BDocumentation%2BTemplate?utm_source=chatgpt.com)):

- **1 Introduction & Context**
- **2 Stakeholder Viewpoints & Concerns**
- **3 Architectural Views**
    - Context View
    - Logical View
    - Development View
    - Process View
    - Deployment View
- **4 Crosscutting Concerns**
    - Security
    - Performance

#### 3.4 Design Specification

Document detailed design elements:

- **Class & Component Diagrams**
- **Sequence & Collaboration Diagrams**
- **API Specifications**
- **Data Model Definitions**

#### 3.5 Test Cases

Use a consistent test case template to ensure coverage ([Test Case Template with Examples: Free Excel & Word Sample for ...](https://katalon.com/resources-center/blog/test-case-template-examples?utm_source=chatgpt.com), [5 Test Case Template Examples - Monday.com](https://monday.com/blog/rnd/test-case-template/?utm_source=chatgpt.com)):

| Field           | Description                   |
|-----------------|-------------------------------|
| Test Case ID    | TC-XX                         |
| Title           | Brief scenario title          |
| Description     | Objective of the test         |
| Preconditions   | Setup before execution        |
| Steps           | Detailed step-by-step actions |
| Expected Result | What should happen            |
| Actual Result   | Filled in during test         |
| Status          | Pass/Fail                     |

#### Phase 3 Prompts

**Prompt 6: Requirements Specification (SRS)**  
```
You are a requirements engineer. Using the ISO/IEC/IEEE 29148 SRS structure, draft functional and nonfunctional requirements for OneDriver's file synchronization, authentication, and error handling features. Format each requirement as:
- ID: SRS-Fxx
- Description: one-sentence requirement
- Rationale: short justification
- Acceptance Criteria: bullet points
```  
([SRS Functional Requirements - AI Prompt - DocsBot AI](https://docsbot.ai/prompts/technical/srs-functional-requirements?utm_source=chatgpt.com))  

**Prompt 7: Use Case Generation**  
```
You are a UML use-case specialist. Generate detailed use cases for the following scenarios: UploadFile, DownloadFile, HandleConflict. For each, include:
- Use Case ID
- Name
- Actors
- Preconditions
- MainFlow
- AlternateFlows
- Postconditions

Also produce a PlantUML use-case diagram.
```  
([How to write AI prompts for architecture - Adobe Firefly](https://www.adobe.com/products/firefly/discover/ai-architecture-prompts.html?utm_source=chatgpt.com))  

**Prompt 8: Architecture Document**  
```
You are a software architect. Produce a Software Architecture Document using the Views & Beyond template, including:
- Context View
- Logical View
- Development View
- Process View
- Deployment View
```  
([Best SRS AI Prompts - DocsBot AI](https://docsbot.ai/prompts/tags?tag=SRS&utm_source=chatgpt.com))  

**Prompt 9: Detailed Design Specification**  
```
You are a design engineer. For each major component, generate:
- Class diagrams in PlantUML
- Sequence diagrams in PlantUML
- OpenAPI YAML for any RESTful APIs  
  Provide clear mappings between design elements and code artifacts.
```  
([AI Prompts for Code Reviews - Faqprime](https://faqprime.com/en/ai-prompts-for-code-reviews/?utm_source=chatgpt.com))  

**Prompt 10: Test Case Creation**  
```
You are a QA engineer. Create test cases following the template: ID, Title, Description, Preconditions, Steps, ExpectedResult. Cover:
- Successful synchronization
- Network failure and retry
- Invalid credentials
- Conflict resolution
- Large file handling
```  
([AI Prompts for Test Case Generation - Faqprime](https://faqprime.com/en/ai-prompts-for-test-case-generation/?utm_source=chatgpt.com))  

---

### Phase 4: Review & Validation

1. **Peer Reviews**  
   Conduct walkthroughs of each document with stakeholders.
2. **Automated Validation**
    - Link requirements to code coverage in CI pipelines.
    - Generate diagrams automatically from updated code ([Reverse Engineering (Code to Architecture Documentation) - Reddit](https://www.reddit.com/r/softwarearchitecture/comments/1dhcryn/reverse_engineering_code_to_architecture/?utm_source=chatgpt.com)).
3. **Iteration**  
   Update documentation based on feedback and new code changes.

#### Phase 4 Prompts

**Prompt 11: Peer Review Checklist**  
```
You are a senior engineer. Generate a comprehensive review checklist covering:
- Code style and conventions
- Documentation completeness and accuracy
- Requirements traceability
- Architecture-to-code alignment
- Test coverage metrics  
  Output as a markdown bullet list.
```  
([AI Prompts for Code Reviews - Faqprime](https://faqprime.com/en/ai-prompts-for-code-reviews/?utm_source=chatgpt.com))  

**Prompt 12: Automated CI Validation**  
```
You are a DevOps engineer. Write a CI pipeline YAML snippet that:
- Verifies SRS IDs in code comments via regex
- Generates and publishes PlantUML diagrams
- Runs static complexity analysis and fails on thresholds  
  Include comments explaining each step.
```  
([Prompt Security Launches Static Analysis Security Testing for AI ...](https://www.prompt.security/press/prompt-security-launches-static-analysis-security-testing-for-ai-generated-code?utm_source=chatgpt.com))

---

## Summary

This document delivers a set of targeted AI prompts for each phase of reverse documenting the OneDriver project—covering feature/module extraction, requirements drafting, architecture mapping, design specification, test case generation, and peer-review automation. Each prompt leverages proven prompt engineering techniques like role prompting and chain-of-thought guidance to maximize clarity and precision. Templates referenced include ISO/IEC/IEEE 29148 SRS prompts and Faqprime test case prompts.

By following this structured process and leveraging the provided templates and AI prompts, you'll be able to generate clear, comprehensive documentation—complete with test cases, design artifacts, architecture views, and requirements specifications—for the OneDriver project.