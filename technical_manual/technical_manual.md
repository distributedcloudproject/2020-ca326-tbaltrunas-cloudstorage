# Distributed Cloud Storage – Technical Manual
This is an 8 to 12 page detailed design document, which reflects both the initial design and the current design, incorporating any major changes made after initial systems design. The contents of each Technical Specification document will vary depending on the nature of the project. However, all projects Technical Specification's must contain the following information as a minimum. The specific format, layout and contents of each document is at the discretion of its authors.

## 0. Table of contents
A table of contents with pages numbers indicated for all sections / headings should be included.

## 1. Introduction

### 1.1 Overview
Provides a brief (half page) overview of the system / product that was developed. Include a description of how it works with other systems (if appropriate).

Cloud storage. Distributed (decentralized) and private.

Interacts with OS.


### 1.2 Glossary
Define and technical terms used in this document. Only include those with which the reader may not be familiar.

Node.

REST API.

React.

## 2. System Architecture
This section describes the high-level overview of the system architecture showing the distribution functions across (potential) system modules. Architectural components that are reused or 3rd party should be highlighted. Unlike the architecture in the Functional Specification - this description must reflect the design components of the system as it is demonstrated.

Go library. TCP.

Desktop GUI.

Desktop CLI.

Web app. Website. HTTPS

Secure communications.

## 3. High-Level Design
This section should set out the high-level design of the system. It should include system models showing the relationship between system components and the systems and its environment. These might be object-models, DFD, etc. Unlike the design in the Functional Specification - this description must reflect the design of the system as it is demonstrated.

Class diagram.

Communications diagrams. TCP. HTTP (auth).
web frontend <-> web backend <-> go library


## 4. Problems and Resolution
This section should include a description of any major problems encountered during the design and implementation of the system and the actions that were taken to resolve them.

Data structure design (files, network).

Distribution algorithm (Calculating node benchmaks).

Auth - JWT.

Testing.

## 5. Installation Guide
This is a 1 to 2 page section which contains a step by step software installation guide. It should include a detailed description of the steps necessary to install the software, a list of all required software, components, versions, hardware, etc.

Obtain our binary.
Or compile from source.
Need Go. Go deps.
create-react-app.
Optional: Makefile, Ansible.

Any OS.

Need own servers to make a cloud.
