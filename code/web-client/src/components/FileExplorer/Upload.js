import React, { useRef } from 'react';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import * as FileExplorerIcons from './Icons';

export default function Upload(props) {
    // Attach refs to component.
    const fileUploaderEl = useRef(null);

    function handleButtonClick(e) {
        fileUploaderEl.current.click();
    }

    function handleInputChange(e) {
        e.preventDefault();
        console.log('Received files: ' + e.target.files)
        const filesArray = Array.from(e.target.files)
        props.callback(filesArray)
    }

    // TODO: Upload multiple files or a folder.
    // TODO: Upload field or box.
    return (
        <Button 
            className='icon-button' 
            onClick={handleButtonClick} >
            <Form.Control
                type="file" 
                id="file" 
                ref={fileUploaderEl} 
                hidden
                onChange={handleInputChange} >
            </Form.Control>
            <FileExplorerIcons.Upload/>
            Upload
        </Button>
    );
}
