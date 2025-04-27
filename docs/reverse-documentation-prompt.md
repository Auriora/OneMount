## Summary
This response delivers a set of targeted AI prompts for each phase of reverse documenting the OneDriver project—covering feature/module extraction, requirements drafting, architecture mapping, design specification, test case generation, and peer-review automation  ([I test ChatGPT for a living - 7 secrets to instantly up your prompt game](https://www.tomsguide.com/ai/i-test-chatgpt-for-a-living-7-secrets-to-instantly-up-your-prompt-game?utm_source=chatgpt.com)). Each prompt leverages proven prompt engineering techniques like role prompting and chain-of-thought guidance to maximize clarity and precision  ([Dynamic Prompt Engineering: Revolutionizing How We Interact with AI](https://medium.com/%40rahulholla1/dynamic-prompt-engineering-revolutionizing-how-we-interact-with-ai-386795e7f432?utm_source=chatgpt.com)). Templates referenced include ISO/IEC/IEEE 29148 SRS prompts  ([Best SRS AI Prompts - DocsBot AI](https://docsbot.ai/prompts/tags?tag=SRS&utm_source=chatgpt.com)) and Faqprime test case prompts  ([AI Prompts for Test Case Generation - Faqprime](https://faqprime.com/en/ai-prompts-for-test-case-generation/?utm_source=chatgpt.com)).

---

## Phase 1: Planning & Setup
### Prompt 1: List All Features and Modules
```
You are a product discovery specialist. Analyze the root and src directories of the OneDriver repository and output a JSON array where each element contains:
- moduleName: string
- description: one-sentence summary
- fileCount: number of files
- dependencies: list of imported modules
Only include modules that define at least one function or class.
```  
([ChatGPT prompt for generating app feature list - Promptmatic](https://promptmatic.ai/prompts/chatgpt/generate-app-feature-list?utm_source=chatgpt.com))  

### Prompt 2: Identify Stakeholders and Goals  
```
You are a business analyst. From the OneDriver README and code comments, extract all stakeholders, their roles, and the primary goals they wish to achieve. Present as a markdown table with columns: Stakeholder, Role, Goal.
```  
([30 ChatGPT Prompts for Software Development Engineers](https://kms-technology.com/emerging-technologies/ai/30-best-chatgpt-prompts-for-software-engineers.html?utm_source=chatgpt.com))  

---

## Phase 2: Reverse Analysis  
### Prompt 3: Static Code Analysis  
```
You are a static analysis specialist. Perform static code analysis on the OneDriver codebase to list all classes and functions along with their cyclomatic complexity. Output a CSV with columns: Name, Type (class/function), FilePath, Complexity.
```  
([Prompt Security Launches Static Analysis Security Testing for AI ...](https://www.prompt.security/press/prompt-security-launches-static-analysis-security-testing-for-ai-generated-code?utm_source=chatgpt.com))  

### Prompt 4: Dynamic Tracing Instrumentation  
```
You are a software instrumentation engineer. Inject logging into all public methods in the core module to record entry, exit, and parameter values at runtime. Provide a unified patch or diff using the existing logging framework.
```  
([Get granular LLM observability by instrumenting your LLM chains](https://www.datadoghq.com/blog/llm-observability-chain-tracing/?utm_source=chatgpt.com))  

### Prompt 5: Behavior Mapping  
```
You are a reverse engineering expert. Execute the primary workflows in OneDriver (e.g., file upload, download, conflict resolution) and log the sequence of invoked functions. Present a sequence diagram in PlantUML syntax illustrating these flows.
```  
([Get granular LLM observability by instrumenting your LLM chains](https://www.datadoghq.com/blog/llm-observability-chain-tracing/?utm_source=chatgpt.com))  

---

## Phase 3: Documentation Drafting  
### Prompt 6: Requirements Specification (SRS)  
```
You are a requirements engineer. Using the ISO/IEC/IEEE 29148 SRS structure, draft functional and nonfunctional requirements for OneDriver’s file synchronization, authentication, and error handling features. Format each requirement as:
- ID: SRS-Fxx
- Description: one-sentence requirement
- Rationale: short justification
- Acceptance Criteria: bullet points
```  
([SRS Functional Requirements - AI Prompt - DocsBot AI](https://docsbot.ai/prompts/technical/srs-functional-requirements?utm_source=chatgpt.com))  

### Prompt 7: Use Case Generation  
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

### Prompt 8: Architecture Document  
```
You are a software architect. Produce a Software Architecture Document using the Views & Beyond template, including:
- Context View
- Logical View
- Development View
- Process View
- Deployment View
```  
([Best SRS AI Prompts - DocsBot AI](https://docsbot.ai/prompts/tags?tag=SRS&utm_source=chatgpt.com))  

### Prompt 9: Detailed Design Specification  
```
You are a design engineer. For each major component, generate:
- Class diagrams in PlantUML
- Sequence diagrams in PlantUML
- OpenAPI YAML for any RESTful APIs  
  Provide clear mappings between design elements and code artifacts.
```  
([AI Prompts for Code Reviews - Faqprime](https://faqprime.com/en/ai-prompts-for-code-reviews/?utm_source=chatgpt.com))  

### Prompt 10: Test Case Creation  
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

## Phase 4: Review & Validation  
### Prompt 11: Peer Review Checklist  
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

### Prompt 12: Automated CI Validation  
```
You are a DevOps engineer. Write a CI pipeline YAML snippet that:
- Verifies SRS IDs in code comments via regex
- Generates and publishes PlantUML diagrams
- Runs static complexity analysis and fails on thresholds  
  Include comments explaining each step.
```  
([Prompt Security Launches Static Analysis Security Testing for AI ...](https://www.prompt.security/press/prompt-security-launches-static-analysis-security-testing-for-ai-generated-code?utm_source=chatgpt.com))