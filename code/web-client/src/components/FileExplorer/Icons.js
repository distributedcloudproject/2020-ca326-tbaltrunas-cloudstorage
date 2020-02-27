import React from 'react';
import DescriptionIcon from '@material-ui/icons/Description';
import FolderIcon from '@material-ui/icons/Folder';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import DeleteIcon from '@material-ui/icons/Delete';
import CachedIcon from '@material-ui/icons/Cached';
import CloudDownloadIcon from '@material-ui/icons/CloudDownload';
import CreateIcon from '@material-ui/icons/Create';
import CloudUploadIcon from '@material-ui/icons/CloudUpload';

const FileExporerIcons = {
    // We use DescriptionIcon (a "document") to represent a File.
    File: <DescriptionIcon />,
    Folder: <FolderIcon style={{ color: 'blue' }} />,
    FolderOpen: <FolderOpenIcon style={{ color: 'blue' }}/>,
    // We use CreateIcon (a pencil) to represent Rename.
    Rename: <CreateIcon />,
    Delete: <DeleteIcon />,
    // We use CachedIcon (an anti-clockwise "loop") to represent Loading.
    Loading: <CachedIcon />,
    // We use CloudDownloadIcon (a cloud with down arrow) to represent Download.
    Download: <CloudDownloadIcon />,

    // Extras
    // Image: < />,
    // PDF: < />,

    // Officially unrendered
    // We use CloudUploadIcon (a cloud with up arrow) to represent Upload.
    Upload: <CloudUploadIcon />,
}

export default FileExporerIcons;
