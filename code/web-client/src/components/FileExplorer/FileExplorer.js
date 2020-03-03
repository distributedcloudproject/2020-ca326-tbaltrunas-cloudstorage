import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
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
      selectedFile: {key: 'test'},
    }

    componentDidMount() {
      FilesAPI.GetFiles((files) => {
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

    // handleSelectFile updates component state with the current file.
    // This allows the user to download the file, etc.
    handleSelectFile = (file) => {
      console.log('Select: ' + file.key);

      this.setState({selectedFile: {key: file.key}})

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

          {/* Can also drag and drop into the FileBrowser to upload. */}
          <Row className='m-2'>
            <Col>
              <Upload callback={this.handleCreateFiles} />
            </Col>
            <Col>
              <Download file={this.state.selectedFile} />
            </Col>
          </Row>

          <Row className='m-2'>
            <Col>
            <FileBrowser
              files={this.state.files}
              icons={FileExplorerIcons.IconObjects}

              // TODO: buttons to specify sort order

              // Handlers
              onCreateFolder={this.handleCreateFolder}
              onCreateFiles={this.handleCreateFiles}
              onMoveFolder={this.handleRenameFolder}
              onMoveFile={this.handleRenameFile}
              onRenameFolder={this.handleRenameFolder}
              onRenameFile={this.handleRenameFile}
              onDeleteFolder={this.handleDeleteFolder}
              onDeleteFile={this.handleDeleteFile}

              onDownloadFile={(keys) => {
                console.log('Download: ' + keys);
              }}

              // onSelect: PropTypes.func,
              onSelectFile={this.handleSelectFile}
              // onSelectFolder: PropTypes.func,
          
              onPreviewOpen={(file) => {
                console.log('Preview: ' + file);
              }}
              onPreviewClose={(file) => {
                console.log('Preview close: ' + file);
              }}
          
              // detailRenderer={() => {}}
            />
            </Col>
          </Row>
        </Container>
      )
    }
}
