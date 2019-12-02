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

To test our project in real life, we will need to use our own computers and rent servers in order to create our private cloud storage platform.

As students we can use various cloud computing services for free. GCP free tier always lets us create a compute engine instance (Windows/Linux VM). Azure provides free Linux and Windows VM's for the first twelve months, similarly to AWS.

Similar products have been created before. However, our unique USP is that we will provide software.

### 1.3 Glossary

Define and technical terms used in this document. Only include those with which the reader may not be familiar.

**Cloud storage platform** - storing data on a network of computers accessible via the Internet.

**Node** - any computer system capable of participating in a cloud storage platform. Including but not limited to virtual machines, servers, PC's.

**VM** - virtual machine.

**GCP** - Google Cloud Platform, paid and free cloud services by Google.

**AWS** - Amazon Web Services, cloud services by Amazon.

**Azure** - Microsoft Azure, cloud services by Microsoft.

**Go** - modern C-like general-purpose programming language.

**React Native** - framework for building hybrid (Android and iOS) mobile applications using JavaScript.


## 2. General Description



### 2.1 Product / System Functions

Describes the general functionality of the system / product.

The product comes in two parts
	* The "node".
	* The "client".
	
The "node" software turns the current machine into a participant of the cloud storage platform. It is like a daemon.

The "client" software allows the "end user" to access the data stored on the cloud storage platform uniformly and easily through a UI.

The "client" will be many different things - a desktop program with a GUI, a website hosted or ran locally, and a mobile application.


### 2.2 User Characteristics and Objectives

Describes the features of the user community, including their expected expertise with software systems and the application domain. Explain the objectives and requirements for the system from the user's perspective. It may include a "wish list" of desirable characteristics, along with more feasible solutions that are in line with the business objectives.

We can identify two user groups (which may overlap):
1. Systems Administrator - sets up nodes to create a cloud storage platform.
2. Non-technical or end user - acesses the cloud storage platform through a client.

The Systems Administrator

Should be able to install the software through some standard procedure such as download and run the software as an executable.

Should be able to control the node daemon through CLI/GUI (desired) or OS-specific daemon manager such as systemd (feasible).

The user should be able to follow a GUI (desired) or a command-line interface (feasible) to set up the node.

The user should modify the node's settings through a GUI (desired) or a configuration file (feasible).

**This user does not necessarily need to know anything about cloud storage. Such details must be abstracted away.**

The End User

If needed, the user should be able to download and install the client easily through an executable. They should be able to launch the client easily through an executable and stop it through graphical controls.

Should be able to enter their credentials to access their cloud storage platform.

Should be able to see a list of files currently stored on the platform as an intuitive directory structure, view file metadata, preview the files (desired), and download the files [read permission]. 

Should be able to upload new files, create a directory structure, move the files within the directory structure, delete directories [write permission].

Should be able to cache the files on the device running the client for quick access (desired).

**This user does not need to know anything about the underlying cloud network, how the data is distributed and replicated, etc.**



### 2.3 Operational Scenarios

This section should describe a set of scenarios that illustrate, from the user's perspective, what will be experienced when utilizing the system under various situations.

    In the article Inquiry-Based Requirements Analysis (IEEE Software, March 1994), scenarios are defined as follows:
    In the broad sense, a scenario is simply a proposed specific use of the system. More specifically, a scenario is a description of one or more end-to-end transactions involving the required system and its environment. Scenarios can be documented in different ways, depending up on the level of detail needed. The simplest form is a use case, which consists merely of a short description with a number attached. More detailed forms are called scripts.
    

Scenario ID: 1
User Objective: Upload file to the cloud.
User Action: Using a client the user selects their file, selects destination directory if any, initiates the upload, sees the progress of the upload, and sees the file appear in cloud directory structure.
Comment:


### 2.4 Constraints

Lists general constraints placed upon the design team, including speed requirements, industry protocols, hardware platforms, and so forth.

Speed - the users expect the reading and writing of data to the cloud to be as fast as possible. Thus we must use node benchmarking, compression, etc.

Security - we boast of privacy and control, therefore we must upkeep it with encryption, redundancy/replication.

Operating Systems - we must support Linux, Windows, and Mac OS X for our nodes and desktop client software. We must support Android and iOS for our mobile client software. We will prioritise Linux and Windows support at first as that is the operating system of most servers. Through the use of portable technologies like Go or React Native we may not need additional coding for different hardware, but our testing may be limited to the hardware that we own as a team.

Time - there may not be enough time for everything
	* One of the team members is not as experienced at Go.
	* The team may not be as experienced at desktop GUI applications, web and mobile development.
	* A plethora of clients could be created. If there is not enough time we may not do the mobile client and/or the web client.


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

* https://cloud.google.com/free/ - Google Cloud Platform free tier.
* https://azure.microsoft.com/en-us/free/free-account-faq/ - Microsoft Azure free tier.
* https://aws.amazon.com/free/ - AWS free tier.
* https://facebook.github.io/react-native/docs/out-of-tree-platforms - React Native platform support.