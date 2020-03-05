import React from 'react';
import Button from 'react-bootstrap/Button';
import urljoin from 'url-join';
import * as FileExplorerIcons from './Icons';
import { FilesDownloadAPI, APIConstants } from '../../api';

export default function Download(props) {
    async function handleButtonClick(e) {
        // FIXME: No page refresh when download link is clicked.
        // e.preventDefault();
        try {
            const fileURL = await FilesDownloadAPI.GetFileDownloadLink(props.file);
            console.log('Computed temporary file URL: ', fileURL);

            // Create a temporary hidden link and click it.
            const link = document.createElement('a');
            document.body.appendChild(link);
            link.href = fileURL;
            link.setAttribute('type', 'hidden')
            link.download = true;

            link.click()
        } catch (error) {
            console.error(error)
        }
        
    }

    return (
        <Button 
            className="icon-button"
            onClick={handleButtonClick} >
                < FileExplorerIcons.Download/>
                Download
        </Button>
    );
}
