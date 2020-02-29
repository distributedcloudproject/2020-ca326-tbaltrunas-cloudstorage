import React from 'react';
import Button from 'react-bootstrap/Button';
import * as FileExplorerIcons from './Icons';
import { GetFileContentsLink} from '../../api/FilesDownload';

export default function Download(props) {
    function handleButtonClick(e) {
        // FIXME: No page refresh when download is clicked.
        // e.preventDefault();
        return false
    }

    return (
        <Button 
            className="icon-button"
            onClick={ handleButtonClick } 
            href={ GetFileContentsLink(props.file) } 
            download >
                < FileExplorerIcons.Download/>
                Download
        </Button>
    );
}
