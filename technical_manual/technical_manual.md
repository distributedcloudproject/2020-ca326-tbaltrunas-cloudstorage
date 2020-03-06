# Distributed Cloud Storage – Technical Manual
<!-- This is an 8 to 12 page detailed design document, which reflects both the initial design and the current design, incorporating any major changes made after initial systems design. The contents of each Technical Specification document will vary depending on the nature of the project. However, all projects Technical Specification's must contain the following information as a minimum. The specific format, layout and contents of each document is at the discretion of its authors. -->

## 0. Table of contents

Distributed Cloud Storage – Technical Manual
- 1. Introduction
  - 1.1. Overview
  - 1.2. Glossary
    - 1.2.1. Project-specific Terms
    - 1.2.2. General Terms
- 2. System Architecture
  - 2.1. Operational Overview
  - 2.2. Class Diagram
  - 2.3. Communications Overview
  - 2.4. API Reference
    - 2.4.1. Cloud API Reference
    - 2.4.2. REST API Reference
- 3. High Level Design
  - 3.1. Initial Design
  - 3.2. Current Design
  - 3.3. Major Design Considerations
    - 3.3.1. Go Library
- 4. Problems and Solutions
  - 4.1. Network Communications
    - 4.1.1. RPC (Communication between Nodes)
    - 4.1.2. Message Structure
    - 4.1.3. Authentication
  - 4.2. Cloud and Network data structures
  - 4.3. File Storage data structures
  - 4.4. Distribution Algorithm
  - 4.5. Desktop Client
  - 4.6. Web Client
    - 4.5.1. Frontend
    - 4.5.2. Backend
    - 4.5.3. Secure Communications
  - 4.7. Automation Tools
    - 4.6.1. Deployment
    - 4.6.2. Scripting
    - 4.6.3. Dependency Management
- 5. Installation Guide
  - 5.1. Node Software
    - 5.1.1. Obtain from a release
    - 5.1.2. Compile from source
  - 5.2. Desktop GUI Client
    - 5.2.1. Obtain from a release
    - 5.2.2. Compile from source.
  - 5.3. Web Client
- 6. Testing
  - 6.1. GitLab CI
  - 6.2. Unit & Integration Tests
  - 6.3. System Tests
  - 6.4. User Tests
  - 6.5. Directory `/tests`

## 1. Introduction

### 1.1 Overview
<!-- Provides a brief (half page) overview of the system / product that was developed. Include a description of how it works with other systems (if appropriate). -->

Distributed Cloud Storage – a set of programs that can turn your private servers into a cloud storage platform (think "Google Drive", "iCloud", or "Dropbox"). Our "node software" uses the Internet to connect your servers ("nodes") into a cloud designed for storage. Use one of our "client programs" to connect to your network and upload/download files, all as if the network was a single cloud entity.

Distributed (de-centralised), secure, intelligent.

- Leverages the nodes' underlying Operating Systems for persistent storage.
- Intelligent routing of files to the most optimal node in terms of storage load and network benchmarks.
- Reliability and privacy of storage at all times through redundancy and encrypted communications.
- Minimised single points of failure. Each node acts both as a client and as a server (a distributed system).

Portable (cross-platform), easily installable "node software" for technical/industry users requiring off-the-shelf private cloud storage solutions. Configure through a graphical or command-line interface.

"Client programs" including a mobile friendly website client and graphical desktop client for the end-users of storage. Modern file explorer UI/UX to interact with the cloud storage platform.

### 1.2 Glossary
<!-- Define and technical terms used in this document. Only include those with which the reader may not be familiar. -->

#### 1.2.1 Project-specific Terms

**Node** - a server or computer system capable of participating in a storage cloud (capabilities: network stack, persistent file system, etc.)

**Node software** - a program that joins the computer it is running on into a storage cloud, intended to be used by technical users.

**Client program** - a program or interface that connects the user to the storage cloud and allows them to store and download their files, intended for end-users that may not be as technical.

**Node administrator** - a user that interacts with the cloud storage in a technical way, ensuring set up of "node software" and some "client software" such as the website.

**End user** - a user that interacts with the cloud storage non-technically, to upload and download files.

#### 1.2.2 General Terms

**Cloud** - network of computers connected via the Internet that expose some interface to the outside world.

**Storage Cloud** - a cloud designed to expose file storage.

**Go, Golang** - performant, concurrent, C like general-purpose programming language [https://golang.org].

**RPC (Remote Procedure Call)** - executing a function on a different computer.

**Gob** - Go standard library package for encoding/decoding variables into binary and vice versa, used for RPC [https://golang.org/pkg/encoding/gob/].

**Binary executable** - a single file that can be distributed and executed as a complete program (for example, .exe on Windows).

**REST API** - a style for web (HTTP) API's, important aspects include a client-server architecture and stateless requests (server treats each request as if the request had everything that was needed to serve it).

**GCP (Google Cloud Platform)** - cloud services provided by Google, including ability to rent a virtual machine with an external IP [https://cloud.google.com/free/].

**Fyne** - Go third party library for desktop-based portable GUI's [https://github.com/fyne-io/fyne].

**React.js** - JavaScript front-end web development library, declarative and stateful [https://reactjs.org/].

**Bootstrap** - CSS front-end library for mobile-friendly user interfaces [https://getbootstrap.com/].

**PostgreSQL** - a relational (SQL) database [https://www.postgresql.org/].

**SPA (Single Page Application)** - a website that is served once and updates dynamically instead of from browser refresh.

**Ansible** - an automation tool for deploying software onto machines via SSH using a declarative configuration [https://www.ansible.com/].

**Make** - a Unix tool for automatically building software via a set of rules [https://www.gnu.org/software/make/].

## 2. System Architecture
<!-- This section describes the high-level overview of the system architecture showing the distribution functions across (potential) system modules. Architectural components that are reused or 3rd party should be highlighted. Unlike the architecture in the Functional Specification - this description must reflect the design components of the system as it is demonstrated. -->

### 2.1. Operational Overview

The following is an Operational Overview Diagram

![operational overview diagram](Operational&#32;Overview.png)

### 2.2. Class Diagram

Go library

### 2.3. Communications Overview

Secure comms (TCP)

HTTPS

### 2.4. API Reference

## 3. High-Level Design
<!-- This section should set out the high-level design of the system. It should include system models showing the relationship between system components and the systems and its environment. These might be object-models, DFD, etc. Unlike the design in the Functional Specification - this description must reflect the design of the system as it is demonstrated. -->

### 3.3. Major Design Considerations

#### 3.3.1. Communication Layer
One of the major design considerations is the communication layer between nodes. As the nodes are in constant communication, it plays a key part in the cloud.

The main options we considered for communication are:
1. Using existing RPC libraries, such as gRPC.
2. Acting as HTTP server and communicating using Rest API.
3. Using websockets to communicate.
4. Our own communication library.

**RPC libraries, gRPC**
The first consideration was using an existing RPC library, mainly considering gRPC as this is the most known/widely used library.

One of the main benefits of gRPC is cross-language communication. gRPC is supported by most modern programming languages that are used, such as C++, Java, Golang, Python, etc... However this would not benefit us, as we had no plans to implement the cloud in other languages.

Another benefit of gRPC is it's speed. As gRPC files requires to be compiled seperately, it allows for much faster encoding and decoding of variables than using gob, as the structure is known before. Unfortunately, it does need to be compiled seperately which is less than ideal.

The drawback of gRPC is it's built for client-to-server communication. As our cloud is decentralized, every node is a client and a server. With this, twice as many sockets would have to be opened. Additionally, it would be impossible for nodes to participate in the network if their ports are closed. As they would only be able to send requests, and not receive. We decided it was not worth the trade off for a tiny performance improvement with encoding and decoding.

**Rest API**
The second consideration was to create a HTTP server and expose it using rest API and passing the data with json. This had the same drawback of being client-to-server communication like gRPC. Additionally, it would not have the performance improvements as gRPC would, as the encoding and decoding has to be flexible on the data structures passed.

Having to work with the data as json would also add to the work required when implementing functions. As instead of receiving fixed data structures, we would end up with working on a json object.

**Websockets**
Websockets was another consideration. It would allow bi-directional communication (client-to-client). It would not provide much more benefits over than using pure TCP sockets, and as such we decided not to go with it.

**Creating our own communication library**
We decided to create our own library to facilitate communication between nodes. The library is built upon TCP sockets to ensure reliability. It allows bi-directional communication (client-to-client) and gives us full control over the communication layer. By encoding and decoding to gob, it allows for minimal overhead.

All communication is encrypted, using a system based of TLS. Public keys are exchanged. A symmetric key is generated, encrypted using the public key, and sent over. The symmetric key is used for encrypting and decrypting the data that is sent as opposed to the public key, as symmetrical encryption is much faster than asymmetrical.

#### 3.3.2. Go Library

As there are multiple uses of the cloud (Desktop CLI, Desktop GUI, Web app), we created a Go library to handle the cloud backbone. Making it easy to use and control outside. This library was created to be simple in use but offer as much control as possible.

Connecting to an existing network is as easy as:
```
cloud, err := network.BootstrapToNetwork("networkip:port", network.Node{
			Name:      "Node Name",
			IP:        ":9002",
			PublicKey: key.PublicKey,
		}, privateKey, network.CloudConfig{FileStorageDir: "/path/to/dir/to/store})
```

Allowing for easy interfacing with the cloud:
```
// Syncs local folder to a folder on the cloud.
c.SyncFolder("/folder/on/cloud", "/local/folder");

// Adds a local file to the cloud.
c.AddFile(fileMetadata, "/path/on/cloud", "/local/file")
```

#### 3.3.3. Go GUI Library

Creating GUI in Golang is not the smoothest experience. There is no official library to do this, and many libraries are still under-developed. When choosing a GUI library for the desktop client, we mainly looked at those factors:
1. No runtime requirements - The compiled executable should be enough to run the program. There should be no third-party programs that are required to be installed in order to run the program.
2. Cross compatibility - One of the main benefits of Golang is it's cross compatibility. As such, it was very important to keep this.
3. Lightweight - The library should be lightweight and performant to run. This eliminated a lot of libraries that depend on HTML/CSS/JS combo to run.
4. Decent design - Having a decent working design that can be used is really beneficial. It will save a lot of time designing our own from scratch.
5. Flexible - Having full control over the GUI was an important factor.

With all of those factors considered, we decided to settle on [fyne.io](fyne.io) library. It ticked all of the boxes above.

## 4. Problems and Solutions
<!-- This section should include a description of any major problems encountered during the design and implementation of the system and the actions that were taken to resolve them. -->

### 4.1. Network Communications

#### 4.1.1. RPC (Communication between Nodes)
Nodes are in constant communication between each other, making it important to get the communications right. All communications between nodes is done using a TCP socket, ensuring the communication between them is reliable. The data sent is encoded and decoded using Gob on the fly. This allows for highely performant communication while maintaining ease of use. This allows passing Go structs as parameters and the data will be decoded/encoded in the communication layer.

The function that is responsible for sending RPC communications is `SendMessage(functionName string, args...interface{})`. In that function, the arguements passed are serialized into gob data. The call is blocking until a response is received. The variables received are returned as an array. Additionally, the `SendMessage` returns an error. If the function on the receiving end returns an error, the error is seperated from outputs and placed into the `err` variable. Additionally, the error can be caused by failure to send the request, or the request timing out.

An example demonstrating the ease of writing communication functions between nodes. Node 1 calls `GetNetwork()`, which calls `OnGetNetworkRequest()` on Node 2, returning any information to node 1.
```
type Network struct {
	Name string
	Nodes []Node
}

func GetNetwork() (Network, error) {
	network, err := client.SendMessage(GetNetworkMsg)
  // network[0] is Network casted as interface{} received from OnGetNetworkRequest(). The err is the error received from that same function, or an error if the request could not be sent, or if the response was not received in given time.
  return network[0].(Network), err
}

func OnGetNetworkRequest() (Network, error) {
	return Network{...}, nil
}
```

#### 4.1.2. Message Structure

The message structure for socket communication starts off with 9 bytes of headers.
The first byte, determines if the message is a response to another message.
The next 4 bytes hold the message ID. The message ID is an unsigned int and is incremental. This is used to identify messages and send responses back to specific messages. As the communication is done over a single TCP socket connection, this allows multiple messages to be sent and received at the same time. It is safe for the message ID to overflow.
The next 4 bytes, the last 4 bytes, hold the message length. As it is not guaranteed for the packet to arrive in it's entirety, this is used to combine multiple packets into one request. Allowing for much larger requests than packet size.

The remaining data, of which length is determined by the last 4 bytes in the header, contains the function name to call and the corresponding data. In the case of the message being a response, function name is omitted.

The function name is escaped by a NULL (\0) character. Anything after that is encrypted gob data of variables that are passed.
The function name determines which function on the receiving end to call for the request.

Each function name will have a handler, which points towards a function in the code to call.


#### 4.1.3. Authentication

Every node connecting to the cloud network has to authenticate before participating in the network. Each node has a unique identifier, which is sha256 sum of a public key. Any node in the network can add a node ID to be able to join the network. Upon establishing a socket connection, public keys are exchanged. The node on on the network generates a symmetric key that will be used for encrypting and decrypting all communication between those two nodes, and encrypts it with the connecting node's public key and sends it to them. This ensures that the connecting node owns the private key corresponding to the public key and cannot fake another node's identity.

### 4.2. Cloud and Network data structures

Cloud and Network

### 4.3. File Storage data structures

Chunks

### 4.4. Distribution Algorithm

Distribution algorithm (Calculating node benchmaks).

### 4.5. Desktop Client

### 4.6. Web Client

Frontend - bootstrap

Secure comms - HTTPS certs.

Auth - JWT. Download. Auth middleware.
DL.
Login.
Postgresql

### 4.7. Automation Tools

Need Go. Go deps.


## 5. Installation Guide
<!-- This is a 1 to 2 page section which contains a step by step software installation guide. It should include a detailed description of the steps necessary to install the software, a list of all required software, components, versions, hardware, etc. -->

Simple steps to install our node software and clients.

Where there are command-line examples, it is assumed that the environment is Unix (corresponding commands can be found for Windows).

See the User Manual (Node Administrator section) for more details on set up of the software.

All our software is cross-platform and compatible with most modern operating systems, including Windows, Linux, Mac OS X [https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies](). Some clients such as the web client work anywhere where there is a web browser.

We do not require special hardware. The software can manage both powerful and less powerful machines. 

In the case of executable binaries we provide precompiled releases on our GitLab for different architectures. Another option is to compile from source.

### 5.1. Node Software

Obtain a binary distribution of our node software (named "cloud").

It is recommended for the node machine to have enhanced storage hardware (in storage space, RAID, etc.) and good or excellent network connectivity.

#### 5.1.1. Obtain from a release.

<!-- TODO -->
See our GitLab releases.

#### 5.1.2. Compile from source.

Clone the project's GitLab repository
```
git clone https://gitlab.computing.dcu.ie/baltrut2/2020-ca326-tbaltrunas-cloudstorage.git
```

Change directory into the node software
```
cd 2020-ca326-tbaltrunas-cloudstorage
cd code/cloud
```

Compile the software into a binary
`go build cloud`

Optionally with the `Make` tool:
```
make
```

Find the node software executable binary under the name `cloud`.

### 5.2. Desktop GUI Client 

Obtain a binary of the desktop GUI client.

The client requires the machine to have a graphical monitor.

#### 5.2.1. Obtain from a release.

<!-- TODO -->
See our GitLab releases.

#### 5.2.2. Compile from source.

Clone the project's GitLab repository
```
git clone https://gitlab.computing.dcu.ie/baltrut2/2020-ca326-tbaltrunas-cloudstorage.git
```

Change directory into the desktop client's directory
```
cd 2020-ca326-tbaltrunas-cloudstorage
cd code/cloud/des
```

Compile the software into a binary
```
go build
```

Optionally with the `Make` tool:
```
make
```

Find the binary.

<!-- TODO -->

### 5.3. Web Client

The web client runs as a website. See the user guide for in depth details of how to set up the web client from a node administrator's point of view.

## 6. Testing

Unit and integration tests.

System tests.

User testing.
