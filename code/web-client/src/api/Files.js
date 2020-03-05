import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// Perform a POST request with a file.
export function CreateFile(file) {
    console.log("Called API method: CreateFile")
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
            console.error(response.status)
        }
    }).catch(error => {
        console.error(error)
    })
}

// Perform a GET request for all files.
export async function GetFiles() {
    console.log("Called API method: GetFiles")
    const url = urljoin(Constants.GetBase(), '/files')
    console.log('Computed URL: ' + url)
    try {
        const response = await axios.get(url)
        if (response.status !== 200) {
            console.error(response.status)
        }
        return response.data
    } catch (error) {
        console.error(error)
    }
}

// Perform a GET request for a file's metadata.
// See FilesDownload.GetFileContentsLink for a file's contents.
export function GetFile(fileKey) {
    console.log("Called API method: GetFile")
    const url = urljoin(Constants.GetBase(), `/files/${fileKey}`)
    console.log('Computed URL: ' + url)
    axios.get(url, {
        responseType: 'blob',
    }).then(response => {
        if (response.status === 200) {
            
            // callback(response.data)
        } else {
            console.log(response.status)
        }
    })
    .catch(error => {
        console.log(error)
    })
}


// Perform a PUT request on a file.
export function UpdateFile(fileKey, newFileKey) {
    console.log("Called API method: UpdateFile")
    const url = urljoin(Constants.GetBase(), `/files/${fileKey}`, `?path=${newFileKey}`)
    console.log('Computed URL: ' + url)
    axios.put(url)
        .then(response => {
            if (response.status !== 200) {
                console.error(response.status)
            }
        }).catch(error => {
            console.error(error)
        })   
}

// Perform a DELETE request on a file.
export function DeleteFile(fileKey) {
    console.log("Called API method: DeleteFile")
    const url = urljoin(Constants.GetBase(), `/files/${fileKey}`)
    console.log('Computed URL: ' + url)
    axios.delete(url)
        .then(response => {
            if (response.status !== 200) {
                console.error(response.status)
            }
        }).catch(error => {
            console.error(error)
        })
}

// Perform a POST request with a folder.
export function CreateFolder(folderKey) {
    console.log("Called API method: CreateFolder")
    console.log("Creating new folder: " + folderKey)
    const url = urljoin(Constants.GetBase(), '/directories', `?path=/${folderKey}`)
    axios.post(url)
        .then(response => {
            if (response.status !== 200) {
                console.error(response.status)
            }
        }).catch(error => {
            console.error(error)
        })
}

// Perform a GET request for all folders.
export async function GetFolders() {
    console.log("Called API method: GetFolders")
    const url = urljoin(Constants.GetBase(), '/directories')
    console.log("Computed URL: " + url)
    try {
        const response = await axios.get(url)
        if (response.status !== 200) {
            console.error(response.status)
        }
        return response.data
    } catch (error) {
        console.error(error)
    }
}

// Perform a GET request for a folder.
export function GetFolder(folderID, callback) {
    console.log("Called API method: GetFolder")
}

// Perform a PUT request on a folder.
export function UpdateFolder(folderID, folder, callback) {
    console.log("Called API method: UpdateFolder")
}

// Perform a DELETE request on a folder.
export function DeleteFolder(folderKey) {
    console.log("Called API method: DeleteFolder")
    console.log("Deleting folder: "+ folderKey)
    const url = urljoin(Constants.GetBase(), '/directories', `?path=/${folderKey}`)
    console.log("Computed URL: " + url)
    axios.delete(url)
        .then(response => {
            if (response.status !== 200) {
                console.error(response.status)
            }
        }).catch(error => {
            console.error(error)
        })
}

