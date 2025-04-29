# Requirements-Architecture Gap Analysis

This document identifies and analyzes gaps between the requirements specified in the Software Requirements Specification (SRS) and the architectural elements described in the Software Architecture Specification (SAS) for the onedriver project.

## 1. Introduction

The purpose of this gap analysis is to identify areas where:
1. Requirements are not fully addressed in the architecture
2. Architectural elements exist without clear requirements driving them
3. Inconsistencies exist between requirements and architectural decisions

This analysis will help ensure that the architecture fully supports all requirements and that all architectural decisions are properly justified by requirements.

## 2. Methodology

The analysis was conducted by:
1. Reviewing the Software Requirements Specification (SRS)
2. Examining the Software Architecture Specification (SAS)
3. Analyzing the Requirements Traceability Matrix
4. Identifying gaps and inconsistencies between these documents

## 3. Gap Analysis

### 3.1 Requirements Not Fully Addressed in Architecture

| Requirement ID | Requirement Description | Gap Description | Severity | Recommendation |
|----------------|-------------------------|-----------------|----------|----------------|
| NFR-REL-002 | The system shall recover gracefully from crashes. | The SAS mentions "Crash recovery mechanisms" in section 4.3 (Other Crosscutting Concerns - Availability), but doesn't provide specific architectural mechanisms for crash detection, recovery, or state preservation. | High | Enhance the SAS to detail the crash recovery architecture, including how state is preserved, how crashes are detected, and the recovery process.<br>**Status: Implemented.** The SAS has been updated to include detailed information about the crash recovery architecture, including state preservation using bbolt database, crash detection through stale lock files, and the recovery process. |
| NFR-REL-004 | The system shall handle API rate limiting gracefully. | The SAS mentions "Graceful handling of API rate limits" in section 4.3 (Other Crosscutting Concerns - Reliability), but doesn't describe the architectural approach to rate limit detection, backoff strategies, or user notification. | Medium | Expand the SAS to include the architectural approach to handling API rate limits, including detection mechanisms, backoff strategies, and user notification.<br>**Status: Implemented.** The SAS has been updated to include detailed information about API rate limit handling, including detection mechanisms through HTTP 429 responses, exponential backoff strategies with jitter, request prioritization, rate tracking with sliding window counters, and user notification with different severity levels. |

### 3.2 Architectural Elements Without Clear Requirements

| Architectural Element | Description | Gap Description | Severity | Recommendation                                                                                                                                                                                                                                                                                                                 |
|----------------------|-------------|-----------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Thumbnail Cache | The SAS mentions a thumbnail cache in sections 3.2.3 (Key Abstractions) and 4.2.3 (Caching Strategy), but there's no specific requirement for thumbnail caching in the SRS. | This architectural feature lacks a clear requirement justification. | Low | Add a requirement for thumbnail caching to the SRS, or clarify how this architectural element supports existing requirements.<br>**Status: Implemented.** Requirements FR-FS-007 (thumbnail caching) and FR-FS-008 (direct thumbnail endpoint) have been added to the SRS and mapped in the Requirements Traceability Matrix. |
| Subscription-based Change Notification | The design-to-code mapping document mentions "Subscription-based Change Notification" and "fs/subscription.go", but this isn't clearly described in the SAS or traced to specific requirements. | This architectural feature lacks visibility in the SAS and clear requirement traceability. | Medium | Update the SAS to include details about the subscription-based change notification mechanism and clarify which requirements it supports.<br>**Status: Implemented.** Requirement FR-FS-009 (subscription-based change notifications) has been added to the SRS and the SAS has been updated with details about the subscription-based change notification mechanism in sections 3.4.1 (Runtime Processes), 4.2.2 (Scalability), and 4.2.2.1 (Subscription-based Change Notification). The requirement has been mapped in the Requirements Traceability Matrix. |
| QuickXORHash Implementation | The SAS mentions "fs/graph/quickxorhash" in section 3.3.1 (Module Organization), but there's no specific requirement for this hashing algorithm. | This architectural element lacks a clear requirement justification. | Low | Add a requirement for file integrity verification using QuickXORHash to the SRS, or clarify how this architectural element supports existing requirements.<br>**Status: Implemented.** Requirement NFR-REL-005 (QuickXORHash for file integrity verification) has been added to the SRS and mapped in the Requirements Traceability Matrix. |

### 3.3 Inconsistencies Between Requirements and Architecture

| Requirement ID | Architectural Element | Inconsistency Description | Severity | Recommendation |
|----------------|------------------------|---------------------------|----------|----------------|
| FR-FS-006 | Conflict Resolution | The requirement states "The system shall handle file conflicts between local and remote changes," but the architectural description in SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) doesn't provide a clear strategy for how conflicts are detected, presented to users, and resolved. | High | Enhance the architectural description to clearly define the conflict detection mechanism, user notification approach, and resolution strategies.<br>**Status: Implemented.** The SAS has been updated to include detailed information about the conflict detection mechanism (based on file modification times, ETags, and local change tracking), resolution strategy (creating conflict copies with timestamps while preserving local changes), and user notification approach (using extended attributes, D-Bus signals, and UI indicators). |
| FR-OFF-003 | Network Connectivity Detection | The requirement states "The system shall automatically detect network connectivity changes," but the SAS doesn't describe the architectural mechanism for detecting network changes. | Medium | Update the SAS to include the architectural approach for network connectivity detection, including the components responsible and the detection mechanism.<br>**Status: Implemented.** The SAS has been updated to include detailed information about the network connectivity detection approach, including the passive detection mechanism that monitors API call success/failure, the components responsible (delta synchronization loop and IsOffline function), and the detection process that marks the filesystem as offline or online based on API call results. |
| FR-INT-006 | Extended Attributes Fallback | The requirement states "The system shall fall back to extended attributes if D-Bus is not available," but the SAS section 3.4.2 (Process Communication) doesn't provide details on how this fallback mechanism works. | Medium | Expand the SAS to include details about the extended attributes fallback mechanism, including how the system detects D-Bus unavailability and transitions to using extended attributes.<br>**Status: Implemented.** The SAS has been updated to include detailed information about the extended attributes fallback mechanism, including how the system detects D-Bus unavailability during filesystem initialization, how it always stores file status as extended attributes regardless of D-Bus availability, and how the transition between D-Bus and extended attributes is transparent to users and applications. |

## 4. Summary of Findings

The gap analysis identified several areas where the architecture documentation could be improved to better address the requirements:


1. **Reliability Mechanisms**: The architectural approaches for crash recovery (NFR-REL-002) and API rate limit handling (NFR-REL-004) need more detailed descriptions.

2. **Architectural Elements Without Clear Requirements**: Several architectural elements (Thumbnail Cache, Subscription-based Change Notification, QuickXORHash) lack clear requirement justification.

3. **Inconsistent Architectural Descriptions**: Some requirements (FR-FS-006, FR-OFF-003, FR-INT-006) have architectural elements that lack sufficient detail to fully address the requirement.

## 5. Recommendations

Based on the gap analysis, we recommend:


1. **Improve Reliability Architecture**: Enhance the architectural descriptions of crash recovery and API rate limit handling mechanisms.

2. **Update Requirements or Clarify Architectural Elements**: Either add requirements for architectural elements that lack clear requirement justification, or clarify how these elements support existing requirements.

3. **Expand Architectural Descriptions**: Provide more detailed architectural descriptions for conflict resolution, network connectivity detection, and extended attributes fallback.

4. **Review and Update Traceability Matrix**: After addressing the gaps, update the Requirements Traceability Matrix to ensure all requirements are properly traced to architectural elements.

By addressing these gaps, the architecture documentation will better align with the requirements, ensuring that all requirements are properly addressed and all architectural decisions are justified.
