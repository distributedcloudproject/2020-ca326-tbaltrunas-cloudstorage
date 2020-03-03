# Distributed Cloud Storage – User Manual

This is a 5 to 10 page user instruction guide on how to use the software system. It should include a step by step guide on how to use the product major components and should be written for a user and not technical audience (unless the system/product is intended for use by technical persons). You may consider including screen shots. 

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
    
Use a binary. Any OS.

#### Manage files

[Multiple screenshots for each operation]

#### Manage directories

[Multiple screenshots for each operation]

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
