import React from 'react';
import DescriptionIcon from '@material-ui/icons/Description';
import FolderIcon from '@material-ui/icons/Folder';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import DeleteIcon from '@material-ui/icons/Delete';
import CachedIcon from '@material-ui/icons/Cached';
import CloudDownloadIcon from '@material-ui/icons/CloudDownload';
import CreateIcon from '@material-ui/icons/Create';
import CloudUploadIcon from '@material-ui/icons/CloudUpload';

// Icons as functions
const File = DescriptionIcon; // We use DescriptionIcon (a "document") to represent a File.
const Folder = FolderIcon;
const FolderOpen = FolderOpenIcon;
const Rename = CreateIcon; // We use CreateIcon (a pencil) to represent Rename.
const Delete = DeleteIcon;
const Loading = CachedIcon; // We use CachedIcon (an anti-clockwise "loop") to represent Loading.
const Download = CloudDownloadIcon; // We use CloudDownloadIcon (a cloud with down arrow) to represent Download.
const Upload = CloudUploadIcon; // We use CloudUploadIcon (a cloud with up arrow) to represent Upload.

// Icons as objects
const IconObjects = {
    File: <File />,
    Folder: <Folder />,
    FolderOpen: <FolderOpen />,
    Rename: <Rename />,
    Delete: <Delete />,
    Loading: <Loading />,
    Download: <Download />,
    // TODO: Extras
    // Image: < />,
    // PDF: < />,
}

export {
    IconObjects,
    File,
    Folder,
    FolderOpen,
    Rename,
    Delete,
    Loading,
    Download,
    Upload,
}
