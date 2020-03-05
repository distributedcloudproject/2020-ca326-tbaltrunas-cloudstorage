# Distributed Cloud Storage – Technical Manual
<!-- This is an 8 to 12 page detailed design document, which reflects both the initial design and the current design, incorporating any major changes made after initial systems design. The contents of each Technical Specification document will vary depending on the nature of the project. However, all projects Technical Specification's must contain the following information as a minimum. The specific format, layout and contents of each document is at the discretion of its authors. -->

## 0. Table of contents

Distributed Cloud Storage – Technical Manual
- 1. Introduction
  - 1.1. Overview
  - 1.2. Glossary
- 2. System Architecture
  - 2.1. Operational Overview
  - 2.2. Class Diagram
  - 2.3. Communications Overview
  - 2.4. REST API Reference
- 3. High Level Design
  - 3.1. Initial Design
  - 3.2. Current Design
  - 3.3. Major Design Considerations
- 4. Problems and Solutions
  - 4.1. Network communications
  - 4.2. Cloud and Network data structures
  - 4.3. File Storage data structures
  - 4.4. Desktop Client
  - 4.5. Web Client
    - 4.5.1. Frontend and Backend
    - 4.5.2. Secure Communications
  - 4.6. Automation Tools
    - 4.6.1. Deployment
    - 4.6.2. Scripting
    - 4.6.3. Dependency Management
- 5. Installation Guide
  - 5.1. Cloud CLI
  - 5.2. Cloud desktop GUI
  - 5.3. Web client 
- 6. Testing
  - 6.1. GitLab CI
  - 6.2. Unit & Integration Tests
  - 6.3. System Tests
  - 6.4. User Tests
  - 6.5. Directory `/tests`

## 1. Introduction

### 1.1 Overview
<!-- Provides a brief (half page) overview of the system / product that was developed. Include a description of how it works with other systems (if appropriate). -->

Cloud storage. Distributed (decentralized) and private.

Interacts with OS.


### 1.2 Glossary
<!-- Define and technical terms used in this document. Only include those with which the reader may not be familiar. -->

Node.

REST API.

React.

## 2. System Architecture
<!-- This section describes the high-level overview of the system architecture showing the distribution functions across (potential) system modules. Architectural components that are reused or 3rd party should be highlighted. Unlike the architecture in the Functional Specification - this description must reflect the design components of the system as it is demonstrated. -->

Go library. TCP.

Desktop GUI.

Desktop CLI.

Web app. Website. HTTPS

Secure communications.

## 3. High-Level Design
<!-- This section should set out the high-level design of the system. It should include system models showing the relationship between system components and the systems and its environment. These might be object-models, DFD, etc. Unlike the design in the Functional Specification - this description must reflect the design of the system as it is demonstrated. -->

Class diagram.

Communications diagrams. TCP. HTTP (auth).
web frontend <-> web backend <-> go library


## 4. Problems and Solutions
<!-- This section should include a description of any major problems encountered during the design and implementation of the system and the actions that were taken to resolve them. -->

Data structure design (files, network).

Distribution algorithm (Calculating node benchmaks).

Auth - JWT. Download. Auth middleware.

Secure comms - HTTPS certs.

## 5. Installation Guide
<!-- This is a 1 to 2 page section which contains a step by step software installation guide. It should include a detailed description of the steps necessary to install the software, a list of all required software, components, versions, hardware, etc. -->

Obtain our binary.
Or compile from source.
Need Go. Go deps.
create-react-app.
Optional: Makefile, Ansible.

Any OS.

Need own servers to make a cloud.

## 6. Testing

Unit and integration tests.

System tests.

User testing.
