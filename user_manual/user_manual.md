# Distributed Cloud Storage – User Manual

<!-- This is a 5 to 10 page user instruction guide on how to use the software system. It should include a step by step guide on how to use the product major components and should be written for a user and not technical audience (unless the system/product is intended for use by technical persons). You may consider including screen shots.  -->

## Table of Contents

## Intro

Guide for our two primary users - node administrator and end-user.
We present how to set up the cloud, and how to use it with desktop and website clients.

## Node Administrator Guide

Node administrator will set up the cloud on nodes (machines or servers).

Software is OS-independent. 

### Cloud set up (command-line interface)

Cloud binary. See help flag. Examples are on Unix.

[Code sample]

#### Prerequisites

Get binary.

#### Initiate the cloud

Initialize first node to create a cloud.

Pass attributes:
* Private RSA key

Optional attributes:
* Network name

[Code sample]

#### Add a node on the cloud

Bootstrap (connect) other nodes to an already created cloud (already connected node).

Pass address of existing node.

Authenticate with a private RSA key and a whitelist file.

[Code sample]

#### Configure an added node

Configure certain attributes:
* IP and port.
* Save file for cloud state.
* Storage allocated.
* Storage directory.

[Code sample]

Misc attributes:
* Logging.
* Fancy display for cloud state.

[Code sample]

### Cloud set up (desktop graphical user interface)

Binary. Same binary as for end user.

#### Add a node to an existing network

[Multiple screenshots of menu]

#### See administrative information of a node

[Screenshots]

### Website client set up

Set up a website client for end-users.

The web client exposes a File Explorer UI to the storage cloud through the web.

#### 5.3.1. Web Application

Also need:

Install the web application (web backend):

postgresql. Apply schema. Create user with bcrypt password.

Obtain HTTPS TLS certificates.
- Testing: self-signed generate with:

Obtain the cloud binary as from section ...

Run the program with the web backend enabled.

#### 5.3.2. Web Server

npm

Obtain frontend source

install dependencies

Serve with react-create-app or custom web server
npm start

#### 5.3.3. Website

Browser. Open address

#### Prerequisites

Get our web app and front-end code. Database.

#### Enable web app backend on a node

Serve a web app to our website client from a node.

Pass a command-line argument.

[Code sample]

#### Serve the website on a web server

Serve the website.

[Code sample]

#### Change user accounts.

Connect to database.

[Code sample or screenshot]

## End User Guide

User that wishes to store files on the cloud.

### Desktop client

#### Prerequisites
    
Use the desktop binary provided. Supports Windows, Linux and Mac OS X.
Requires a graphics driver to be installed to run.

#### Manage files

The main overview of the file explorer in the desktop GUI looks like this:
![View of the File Explorer](file&#32;explorer.png)

You can navigate through folders by double clicking on them.

Each file has 3 main operations. Sync, Download and Delete.

The sync option, will create a sync between the cloud file and a file locally stored on the system. This sync acts similarly to a symlink. Any change to the file locally will be reflected on the cloud, as well as any changes to the file on the cloud will be reflected locally.

The download option downloads the file from the cloud to a local file. It will download the file chunks from any node that contains it.

The final operation, is the delete operation. This deletes the file from the cloud.

Additionally, at the bottom of the window, there are 2 buttons. `Add` and `Sync`.

The `Add` button will upload a selected file from the local system to the cloud.
The `Sync` button will upload the selected file to the cloud, and create a link between the local version and the cloud version.

#### Manage directories

There are 2 main operations for the folder.

The sync operation on the folder, will sync the cloud folder to a local folder. Similarly to a file sync, but instead it syncs the folder. This is also like Google Drive, Dropbox, etc... where you have a local folder being synced with the cloud. Any changes to the folder and files inside it, will be kept up to date with the cloud.

The other operation is deleting a folder. The folder has to be empty to be deleted.

### Website client

#### Prerequisites

Browser.

Web address.

#### Log in

Credentials provided by administrator.

#### Overview of home screen

[Screenshot pointing out things on the layout]

#### Manage files

[Multiple screenshots]

#### Manage directories

[Multiple screenshots]

#### Search for a file

## Other

More information:
Technical specification link. Source code link.

Contact support team.
