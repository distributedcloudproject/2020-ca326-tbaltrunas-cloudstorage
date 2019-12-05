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


Our project – Distributed Cloud Storage – is essentially your own Google Drive/DropBox/OneDrive. Using our software you can turn a private cloud network into a cloud storage platform. Simply configure our storage software on your "nodes" (networked computers capable of storing data) and use our client software to access the data on your cloud with a File Explorer-like interface.

Why? We only have to do a search for "google drive privacy" or "google drive breach" to see the amount of trust that we place into giant corporations like Google or Microsoft when we use their "free" services. We can not be sure that our data will not be used as part of an advertisement campaign or not get stolen through some cyber attack. Thus, we provide software that you can use on top of your own secure private network and be in control of your data, where the protection of your data is only limited to the protection of the network itself.

Our storage or "node" software will interact with the underlying operating systems of your nodes, be it Linux, Windows or Mac OS X, to carry out its networking, storage, and other ad hoc functions. On the client-side we will have graphical user interfaces for desktop, web, and mobile devices.


### 1.2 Business Context


To test our project in practice, we will need to use our own computers and rent servers to create a private cloud and turn it into a storage platform.

As students we can use various cloud computing services for free. GCP free tier always lets us create a compute engine instance (Windows/Linux VM). Azure provides free Linux and Windows VM's for the first twelve months, and so does AWS.

Products that propose an alternative to "cloud storage giants" already exist in the market. See [Storj](https://storj.io) and [Sia](https://sia.tech/) for examples. However, such products are paid and manage the underlying cloud network. Our business model would be to provide free unmanaged software and charge for tech support.


### 1.3 Glossary

Define and technical terms used in this document. Only include those with which the reader may not be familiar.

**Cloud storage platform** - network of computers accessible via the Internet that stores user data.

**Node** - any computer system capable of participating in a cloud storage platform. Including but not limited to virtual machines, servers, and PC's.

**Go** - modern C-like general-purpose programming language.

**React Native** - framework for building hybrid (Android and iOS) mobile applications using JavaScript.

**VM** - virtual machine.

**GCP** - Google Cloud Platform, paid and free cloud services by Google.

**AWS** - Amazon Web Services, cloud services by Amazon.

**Azure** - Microsoft Azure, cloud services by Microsoft.

**Systemd** - Linux software for managing services (daemons).


## 2. General Description



### 2.1 Product / System Functions

The system comes in two parts:

* The "node" software.
* The "client" software.


**The "node" software** turns the machine it is running on into a participant of the cloud storage platform (a "node"). The machine must have the properties of a "node", including networking and storage capabilities.

After being configured the software will run as a daemon. 

There is no server. Multiple nodes will create a decentralised network.

The following is an overview of it functionalities:

* Perform file IO - read and write data persistently on the node.
* Perform network IO - receive and send packets over the Internet.
* Distribute data - split a file and share it with other nodes.
* Replicate data - send a copy of data to another node for redundancy.
* Select the best node - benchmark other nodes to pick the node to send data to.
* Resolve concurrent modification - check with other nodes to see if a file modification is allowed.


**The "client" software** allows the "end user" to access their data stored on their cloud storage platform.

The "client" is proposed to be implemented on many platforms - as a desktop program with a GUI, a web application, and a mobile application.

The following is an overview of its functionalities:

* Authenticate a user - the user running the client is capable of supplying credentials to authenticate with the cloud storage platform.
* Provide a File Explorer - the user can access, organise, and modify the files on the cloud storage platform.
* Encrypt and decrypt data - encrypt data before sending it from client to storage, and decrypt data locally once it is downloaded from the storage platform.
* Compress and uncompress data - before sending data compress it, and decompress data upon download.

*See the Functional Requirements section for a detailed description of functionalities.*

### 2.2 User Characteristics and Objectives

We can identify two user groups in our user community:

1. Node Administrator.
2. End User.

#### Node Administrator

The Node Administrator sets up and manages nodes on the cloud storage platform. This may be a systems administrator, software developer, or a self-taught individual.

The Node Operator's objective is to turn a private cloud into a cloud storage platform by configuring the nodes on their cloud.

This user needs to have basic systems administration skills such as installing and configuring software through command-line interfaces, graphical user interfaces, and configuration files on Windows/Linux/Mac OS X platforms, on their desktop, VM, or a remote machine. The user does not necessarily need to have any domain-specific knowledge about cloud storage except for basic computing terms such as encryption, compression, and disk space.

The Node Operator's wish list includes:

* To download and install the node software conveniently.
* To configure the node through a GUI (desirable), CLI, or a config file.
* To ensure node software runs in the background at all times through a GUI/CLI (desirable) or OS-specific program such as systemd.


#### End User

The End User accesses the cloud storage platform through a client program. This may be an employee or client in a business or a private individual.

The objective of the End User is to store their data on the cloud.

This user is non-technical and only needs basic computer skills. The user does not need to know anything about cloud storage.

The End User's wish list includes:

* To download, install, and launch the client software easily.
* To be able to authenticate with the cloud storage platform.
* To view, preview (desired), and download their files on the cloud.
* To upload, organise (desired), and delete their files on the cloud.
* To have quick response times from the cloud.
* To be able to encrypt their files.
* To be able to have access to their files reliably and at all times.


### 2.3 Operational Scenarios

This section should describe a set of scenarios that illustrate, from the user's perspective, what will be experienced when utilizing the system under various situations.

    In the article Inquiry-Based Requirements Analysis (IEEE Software, March 1994), scenarios are defined as follows:
    In the broad sense, a scenario is simply a proposed specific use of the system. More specifically, a scenario is a description of one or more end-to-end transactions involving the required system and its environment. Scenarios can be documented in different ways, depending up on the level of detail needed. The simplest form is a use case, which consists merely of a short description with a number attached. More detailed forms are called scripts.
    

Scenario ID: 1
User Objective: Upload file to the cloud.
User Action: Using a client the user selects their file, selects destination directory if any, initiates the upload, sees the progress of the upload, and sees the file appear in cloud directory structure.
Comment:

Concurrent modification.

Node goes down behind the scenes.

One node gets compromised by attacker.



### 2.4 Constraints

Lists general constraints placed upon the design team, including speed requirements, industry protocols, hardware platforms, and so forth.

#### Speed

The users expect the reading and writing of data to the cloud to be as fast as possible. Thus we must use node benchmarking, compression, etc.

#### Security

We boast of privacy and control, therefore we must upkeep it with encryption, redundancy/replication.

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


* **Requirement ID:**
* **Description:**
* **Criticality:**
* **Technical issues:** 
* **Dependencies:** 


* **Requirement ID:** 1
* **Description:** Node networking, including listening for requests, sending back responses, initiating connections with other nodes, and accepting streams of data.
* **Criticality:** Very high. Network IO is an essential component for this project.
* **Technical issues:** Nodes must be reachable by IP address and port number, i.e. be on a public network or port forwarded.
* **Dependencies:** N/A.

* **Requirement ID:** 2
* **Description:** Node set up, including storage allocation.
* **Criticality:** Very high. If there are no active nodes, then the cloud storage platform can not be used.
* **Technical issues:** N/A.
* **Dependencies:** Node networking.

* **Requirement ID:** 3
* **Description:** Node update status to the cloud, including its state (*active*/*inactive*), free space left, latency, bandwidth, etc.
* **Criticality:** Very high. If a node suddenly goes down this puts a risk to data. Also, knowing performance information about nodes is important for decisions.
* **Technical issues:** Communicating information with all nodes efficiently. 
* **Dependencies:** Node set up.

* **Requirement ID:** 4
* **Description:** Node distribute file across the cloud. This includes: 1. Splitting the file contents into chunks. 2. Replicating each chunk. 3. Distributing replicas to other nodes making the file redundant.
* **Criticality:** Very high. Distributing a file is essential for data redundancy.
* **Technical issues:** Managing the distribution of a single file is a complex routine with multiple steps. We need to think about whether to store only file contents or also file metadata. Also if only certain file chunks have changed, we do not need to change other chunks.
* **Dependencies:** Node set up, node status update.

* **Requirement ID:** 5
* **Description:** Node accept file from client and distribute file on the cloud.
* **Criticality:** Very high. One of the core functionalities of this project.
* **Technical issues:** Will need to call the "distribute file" routine.
* **Dependencies:** Node distribute file.

* **Requirement ID:** 6
* **Description:** Node collect and serve file (contents or metadata) to client.
* **Criticality:** Very high. The user must view and download whatever they uploaded to the cloud. Collecting back chunks of a file well is important for performance and data integrity.
* **Technical issues:** Collecting the file's chunks from remote locations is a complex task.
* **Dependencies:** Node accept file from client.

* **Requirement ID:** 7
* **Description:** Node daemonization, where the node software always runs as a background process and is managed by some daemon tool.
* **Criticality:** High. The node process is directly responsible for the availability of the node itself.
* **Technical issues:** Node software must be high performance to not unnecessarily use resources on its host machine. 
* **Dependencies:** N/A.

* **Requirement ID:** 8
* **Description:** Node uninstall software by shutting down the node and deleting all data on the host's data store.
* **Criticality:** High. Cleanly uninstalling software is an important feature for users.
* **Technical issues:** Communicating node uninstall to other nodes and moving data to other nodes to ensure no data loss.
* **Dependencies:** Node set up, node distribute file.

* **Requirement ID:** 9
* **Description:** Node configuration ability, including modifying the allocated storage space, network limits, etc.
* **Criticality:** Medium. This is a convenient feature.
* **Technical issues:** Node must be able to pick up the changes effectively without any or long downtime.
* **Dependencies:** Node set up.

* **Requirement ID:** 10
* **Description:** Node handle changes in case multiple users modify the same file.
* **Criticality:** Low. This is a proposed enhancement where multiple users share the same files and perform a concurrent write.
* **Technical issues:** We have to think about how to resolve concurrency conflicts.
* **Dependencies:** Node accept file, node serve file.




Client connect to cloud.

Client show files.

Client download files.

Client upload files.

Client delete files.

Client CRUD on files.

Client CRUD on directories.

Client create directory structure.

Client preview file.

Client encrypt and decrypt files.

Client compress and uncompress files.

Cache files on the client
Should be able to cache the files on the client device for quick access and freeing up space on the device (desired).


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