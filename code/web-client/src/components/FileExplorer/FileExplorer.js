import React from 'react';
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';
import FileBrowser from 'react-keyed-file-browser';
import '../../../node_modules/react-keyed-file-browser/dist/react-keyed-file-browser.css';
import * as FileExplorerIcons from './Icons';
import './FileExplorer.css';
import Upload from './Upload';
import Download from './Download';
import { FilesAPI } from '../../api';
import * as UpdateUI from './UpdateUI';

export default class FileExplorer extends React.Component {
    state = {
      files: [],
      selectedFile: 'test',
    }

    componentDidMount() {
      const files = FilesAPI.GetFiles((files) => {
        this.setState({ files: files })
      })
    }

    // handleCreateFiles adds new files to existing files.
    // files is an Array of DOM File objects (created by input[type=file]).
    handleCreateFiles = (files) => {
      console.log(files)
      files.forEach(file => {
        FilesAPI.CreateFile(file)
      });
      
      const prefix = ''; // TODO: handle prefix
      this.setState({state: UpdateUI.UpdateUICreateFiles(this.state, files, prefix)})
    }

    // handleRenameFile renames an existing file.
    handleRenameFile = (oldKey, newKey) => {
      FilesAPI.UpdateFile()
      this.setState({state: UpdateUI.UpdateUIRenameFile(this.state, oldKey, newKey)})
    }

    // handleDeleteFile deletes an existing file.
    handleDeleteFile = (fileKey) => {
      FilesAPI.DeleteFile()
      // FIXME: doesn't get deleted.
      this.setState({state : UpdateUI.UpdateUIDeleteFile(this.state, fileKey)})
    }

    // handleCreateFolder creates a new folder.
    handleCreateFolder = (key) => {
      FilesAPI.CreateFolder()
      this.setState({state: UpdateUI.UpdateUICreateFolder(this.state, key)})
    }

    // handleRenameFolder renames an existing folder.
    handleRenameFolder = (oldKey, newKey) => {
      FilesAPI.UpdateFolder()
      this.setState({state: UpdateUI.UpdateUIRenameFolder(this.state, oldKey, newKey)})
    }

    // handleDeleteFolder deletes an existing folder.
    handleDeleteFolder = (folderKey) => {
      FilesAPI.DeleteFolder()
      // FIXME: doesn't get deleted.
      this.setState({state: UpdateUI.UpdateUIDeleteFolder(this.state, folderKey)})
    }
  
    render() {
      return (
        <Container>

          <Upload callback={this.handleCreateFiles} />
          <Download file={this.state.selectedFile} />

          <FileBrowser
            files={this.state.files}
            icons={FileExplorerIcons.IconObjects}

            // Handlers
            onCreateFolder={this.handleCreateFolder}
            onCreateFiles={this.handleCreateFiles}
            onMoveFolder={this.handleRenameFolder}
            onMoveFile={this.handleRenameFile}
            onRenameFolder={this.handleRenameFolder}
            onRenameFile={this.handleRenameFile}
            onDeleteFolder={this.handleDeleteFolder}
            onDeleteFile={this.handleDeleteFile}
          />
        </Container>
      )
    }
}
