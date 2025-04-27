# Onedriver Design Documentation

This directory contains design documentation for the onedriver project, including class diagrams, sequence diagrams, and mappings between design elements and code artifacts.

## Class Diagrams

Class diagrams represent the static structure of the system, showing the classes, their attributes, methods, and relationships.

- [Core Engine Class Diagram](core_engine_class_diagram.puml) - Represents the core filesystem implementation
- [Graph API Class Diagram](graph_api_class_diagram.puml) - Represents the Microsoft Graph API integration
- [UI Class Diagram](ui_class_diagram.puml) - Represents the UI components and command-line interface

## Sequence Diagrams

Sequence diagrams represent the dynamic behavior of the system, showing the interactions between objects over time.

- [Authentication Workflow](auth_sequence_diagram.puml) - Shows the authentication process with Microsoft Graph API
- [File Access Workflow](file_access_sequence_diagram.puml) - Shows how files are accessed from OneDrive
- [File Modification Workflow](file_modification_sequence_diagram.puml) - Shows how files are modified and uploaded to OneDrive
- [Delta Synchronization Workflow](delta_sync_sequence_diagram.puml) - Shows how changes are synchronized between OneDrive and the local filesystem

## Mappings

- [Design to Code Mapping](design_to_code_mapping.md) - Provides a mapping between design elements and code artifacts

## Viewing the Diagrams

The PlantUML diagrams can be viewed using various tools:

1. **Online PlantUML Server**: Copy the content of the .puml file and paste it into the [PlantUML Server](http://www.plantuml.com/plantuml/uml/)
2. **VS Code with PlantUML Extension**: Install the PlantUML extension for VS Code and use Alt+D to preview the diagram
3. **IntelliJ IDEA with PlantUML Plugin**: Install the PlantUML plugin for IntelliJ IDEA and use the preview button
4. **Command Line**: Use the PlantUML JAR file to generate images from the command line:
   ```
   java -jar plantuml.jar diagram.puml
   ```

## Architecture Overview

onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing the entire OneDrive content. The architecture consists of several key components:

1. **Filesystem Implementation (fs package)**: Implements the FUSE filesystem interface to provide a native filesystem experience
2. **Graph API Integration (fs/graph package)**: Handles communication with Microsoft's Graph API for accessing OneDrive
3. **Cache Management**: Manages local caching of file content and metadata to improve performance and enable offline access
4. **Command Line Interface (cmd/onedriver package)**: Provides a command-line interface for mounting and configuring onedriver
5. **Graphical User Interface (ui package and cmd/onedriver-launcher package)**: Provides a graphical interface for managing onedriver mountpoints

For a more detailed overview of the architecture, see the [Software Architecture Document](software_architecture_document.md).