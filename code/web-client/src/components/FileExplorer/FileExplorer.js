import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import FileBrowser from 'react-keyed-file-browser';
import '../../../node_modules/react-keyed-file-browser/dist/react-keyed-file-browser.css';
import * as FileExplorerIcons from './Icons';
import './FileExplorer.css';
import Upload from './Upload';
import { FilesAPI, FilesDownloadAPI } from '../../api';
import * as UpdateUI from './UpdateUI';
import * as CustomRenderers from './fileBrowserCustom';

export default class FileExplorer extends React.Component {
    state = {
      // A list of objects. 
      // Files have key (path), size, modified since (optional).
      // Folders have key (path) only. Must have a "/" at the end of the path.
      files: [],
      selectedFile: {key: 'test'},
    }

    async componentDidMount() {
      // Get existing state of cloud files and folders
      try {
        const files = await FilesAPI.GetFiles();
        console.log('Received number of files: ' + files.length);
        console.log(files);
        const folders = await FilesAPI.GetFolders();
        console.log('Received number of folders: ' + folders.length);
        console.log(folders);
        // TODO: merge files and folders

        const mergedFiles = files.concat(folders);

        this.setState({ files: mergedFiles });
      } catch (error) {
        console.error(error)
      }
      // TODO: might want to call GetFiles on each UI update so that UI doesn't get out of sync with backend
      // FIXME: if can't connect to backend, get undefined files error.
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

    handleReadFiles = async (files) => {
      console.log(files)
      const file = files[0]
      try {
        const fileURL = await FilesDownloadAPI.GetFileDownloadLink(file);
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

    // handleSelectFile updates component state with the current file.
    // This allows the user to download the file, etc.
    handleSelectFile = (file) => {
      console.log('Select: ' + file.key);

      this.setState({selectedFile: {key: file.key}})

    }

    // handleRenameFile renames an existing file.
    handleRenameFile = (oldKey, newKey) => {
      FilesAPI.UpdateFile(oldKey, newKey)
      this.setState({state: UpdateUI.UpdateUIRenameFile(this.state, oldKey, newKey)})
    }

    // handleDeleteFile deletes an existing file.
    handleDeleteFile = (fileKeys) => {
      fileKeys.forEach(fileKey => {
        FilesAPI.DeleteFile(fileKey)
        this.setState({state : UpdateUI.UpdateUIDeleteFile(this.state, fileKey)})
      });
    }

    // handleCreateFolder creates a new folder.
    handleCreateFolder = (folderKey) => {
      FilesAPI.CreateFolder(folderKey)
      this.setState({state: UpdateUI.UpdateUICreateFolder(this.state, folderKey)})
      console.log(this.state.files)
    }

    // handleRenameFolder renames an existing folder.
    handleRenameFolder = (oldKey, newKey) => {
      FilesAPI.UpdateFolder(oldKey, newKey)
      this.setState({state: UpdateUI.UpdateUIRenameFolder(this.state, oldKey, newKey)})
    }

    // handleDeleteFolder deletes an existing folder.
    handleDeleteFolder = (folderKeys) => {
      folderKeys.forEach(folderKey => {
        // FIXME: may want to wrap FilesAPI calls in  try-catch, and only update UI state on success.
        FilesAPI.DeleteFolder(folderKey)
        this.setState({state: UpdateUI.UpdateUIDeleteFolder(this.state, folderKey)})  
      });
    }
  
    render() {
      return (
        <Container className='col-xs-12 m-1'>

          {/* Can also drag and drop into the FileBrowser to upload. */}
          <Row className='mb-3'>
            <Col>
              <Upload callback={this.handleCreateFiles} />
            </Col>
          </Row>

          <Row className='mt-3'>
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
                onDownloadFile={this.handleReadFiles}

                // onSelect: PropTypes.func,
                onSelectFile={this.handleSelectFile}
                // onSelectFolder: PropTypes.func,
            
                onPreviewOpen={(file) => {
                  console.log('Preview: ' + file);
                }}
                onPreviewClose={(file) => {
                  console.log('Preview close: ' + file);
                }}
                
                // headerRenderer={() => {return null}}
                // fileRenderer={() => {return null}}
                // folderRenderer={() => {return null}}

                actionRenderer={CustomRenderers.ActionRenderer}
                // actionRenderer={() => {return null}}

                filterRenderer={CustomRenderers.FilterRenderer}
                
                detailRenderer={CustomRenderers.DetailRenderer}
                confirmDeletionRenderer={CustomRenderers.ConfirmDeletionRenderer}
              />
            </Col>
          </Row>
        </Container>
      )
    }
}
