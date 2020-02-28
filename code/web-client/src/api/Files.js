import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// Perform a GET request for all files.
function GetFiles(callback) {
    console.log("Called API method: GetFiles")
    const url = urljoin(Constants.GetBase(), '/files')
    console.log('Computed URL: ' + url)
    axios.get(url)
        .then(response => {
            if (response.status === 200) {
                callback(response.data)
            } else {
                console.log(response.status)
            }
        })
        .catch(error => {
            console.log(error)
        })
}

// Perform a GET request for a file's metadata.
function GetFile(fileID, callback) {
    console.log("Called API method: GetFile")
    const url = urljoin(Constants.GetBase(), '/files/`${fileID}`')
    console.log('Computed URL: ' + url)
}

// Perform a GET request for a file's contents.
function GetFileContents(fileID, callback) {
    console.log("Called API method: GetFileContents")
    const url = urljoin(Constants.GetBase(), '/files/`${fileID}`', '?type=contents')
    console.log('Computed URL: ' + url)
}

// Perform a POST request with a file.
function CreateFile(file, callback) {
    console.log("Called API method: CreateFile")
    // const reader = file.stream().getReader();
    // const contents = Uint8Array();
    // const numRead = 0;
    // reader.read()
    //     .then(res => {
    //         const {value, done} = res;
    //         console.log(value)
    //         console.log(done)
            
    //         if (done) {
    //             return;
    //         }

    //         numRead += value.length;
    //     })
    //     .catch(error => {
    //         console.error(error)
    //     })
    
    // if (file.size != numRead) {
    //     console.error('File size does not match the number of bytes read: ' + numRead)
    // }

    const formData = new FormData();
    formData.append('file', file)

    axios.post(urljoin(Constants.GetBase(), 
                       '/files', 
                       `?name=${file.name}`,
                       `&size=${file.size}`,
                       `&type=${file.type}`,
                       `&lastModified=${file.lastModified}`
    ), formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        }
    }
    ).then(response => {
        if (response.status !== 200) {
            console.warn(response.status)
        }
    }).catch(error => {
        console.error(error)
    })
}

// Perform a PUT request on a file.
function UpdateFile(fileID, file, callback) {
    console.log("Called API method: UpdateFile")
}

// Perform a DELETE request on a file.
function DeleteFile(fileID, callback) {
    console.log("Called API method: DeleteFile")
}

// Perform a GET request for a folder.
function GetFolder(folderID, callback) {
    console.log("Called API method: GetFolder")
}

// Perform a POST request with a folder.
function CreateFolder(folderID, folder, callback) {
    console.log("Called API method: CreateFolder")
}

// Perform a PUT request on a folder.
function UpdateFolder(folderID, folder, callback) {
    console.log("Called API method: UpdateFolder")
}

// Perform a DELETE request on a folder.
function DeleteFolder(folderID, callback) {
    console.log("Called API method: DeleteFolder")
}

export {
    GetFiles,
    GetFile,
    CreateFile,
    UpdateFile,
    DeleteFile,
    GetFolder,
    CreateFolder,
    UpdateFolder,
    DeleteFolder,
};
