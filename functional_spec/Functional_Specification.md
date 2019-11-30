# Distributed Cloud Storage – Functional Specification

**Student Name:** Tomas Baltrunas
**Student Number:** 17350793

**Student Name:** Bartosz Śwituszak
**Student Number:** 17437072

**Supervisor:** Ray Walshe
**Date:** 06/12/2019

## 0. Table of contents

A table of contents with pages numbers indicated for all sections / headings should be included.


## 1. Introduction



### 1.1 Overview

Provides a brief overview of the system / product to be developed. It should include a description of the need for the system, briefly describe its functions and explain how it will work with other systems (if appropriate).

Our project – Distributed Cloud Storage – is essentially your own "Google Drive/DropBox/OneDrive". Using our software you can turn your private cloud network into a cloud storage platform. Simply install our storage software onto your "nodes" (networked computers capable of storing data) and use our client software to access your data on your cloud with a File Explorer-like interface.

Why? We only have to search for the terms "google drive privacy" or "google drive breach" to see the amount of trust that we place into giant corporations like Google or Microsoft when we use their "free" services. We can not be sure that our data will not be used as part of an advertisement campaign or not get stolen through some cyber attack. For this reason, where privacy is critical we provide our software that you are fully in control of, that you can use on top of your own secure private network. It is true that your data will only be as protected as far as the network itself is protected. However, control of data is still transferred to the user.

Our storage or "node" software will interact with the underlying operating systems of your nodes, be it Linux, Windows or Mac OS X, to carry out its networking, storage, and other ad hoc functions. On the client-side we will have graphical user interfaces for desktop, web, and mobile.


### 1.2 Business Context

Provides an overview of the business organization sponsoring the development of this system / product or in which the system / product will / could be deployed. Note - may not be applicable to all projects

Since the project involves dealing with a network of computers, we will need to rent a network like that for testing purposes. 

As students we can use AWS, Google Cloud, or Microsoft Azure servers.


### 1.3 Glossary

Define and technical terms used in this document. Only include those with which the reader may not be familiar.

**Storage Cloud** - user-controlled private network of computers, that stores data using our software.

**Node** - any computer system capable of participating in the storage cloud. Including but not limited to virtual machines, servers, PC's.

**Go** - modern C-like general-purpose programming language.

**React Native** - framework for building hybrid mobile applications using JavaScript.


## 2. General Description



### 2.1 Product / System Functions

Describes the general functionality of the system / product.


### 2.2 User Characteristics and Objectives

Describes the features of the user community, including their expected expertise with software systems and the application domain. Explain the objectives and requirements for the system from the user's perspective. It may include a "wish list" of desirable characteristics, along with more feasible solutions that are in line with the business objectives.


### 2.3 Operational Scenarios

This section should describe a set of scenarios that illustrate, from the user's perspective, what will be experienced when utilizing the system under various situations.

    In the article Inquiry-Based Requirements Analysis (IEEE Software, March 1994), scenarios are defined as follows:
    In the broad sense, a scenario is simply a proposed specific use of the system. More specifically, a scenario is a description of one or more end-to-end transactions involving the required system and its environment. Scenarios can be documented in different ways, depending up on the level of detail needed. The simplest form is a use case, which consists merely of a short description with a number attached. More detailed forms are called scripts.


### 2.4 Constraints

Lists general constraints placed upon the design team, including speed requirements, industry protocols, hardware platforms, and so forth.


## 3. Functional Requirements

This section lists the functional requirements in ranked order. Functional requirements describes the possible effects of a software system, in other words, what the system must accomplish. Other kinds of requirements (such as interface requirements, performance requirements, or reliability requirements) describe how the system accomplishes its functional requirements.

As an example, each functional requirement could be specified in a format similar to the following:

    Description - A full description of the requirement.
    Criticality - Describes how essential this requirement is to the overall system.
    Technical issues - Describes any design or implementation issues involved in satisfying this requirement.
    Dependencies with other requirements - Describes interactions with other requirements.
    Others as appropriate


## 4. System Architecture

This section describes a high-level overview of the anticipated system architecture showing the distribution functions across (potential) system modules. Architectural components that are reused or 3rd party should be highlighted.


## 5. High-Level Design

This section should set out the high-level design of the system. It should include one or more system models showing the relationship between system components and the systems and its environment. These might be object-models, DFD, etc.


## 6. Preliminary Schedule

This section provides an initial version of the project plan, including the major tasks to be accomplished, their interdependencies, and their tentative start/stop dates. The plan also includes information on hardware, software, and wetware resource requirements.

The project plan should be accompanied by one or more PERT or GANTT charts.


## 7. Appendices

Specifies other useful information for understanding the requirements. 


### 7.1 References
	