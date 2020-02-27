import React from 'react';
import FileBrowser from 'react-keyed-file-browser';
import '../../../node_modules/react-keyed-file-browser/dist/react-keyed-file-browser.css';
import FileExplorerIcons from './Icons';
import './FileExplorer.css';
import { FilesAPI } from '../../api'

export default class FileExplorer extends React.Component {
    state = {
      files: [],
    }

    componentDidMount() {
      const files = FilesAPI.GetFiles((files) => {
        this.setState({ files: files })
      })
    }
  
    // The handlers have been attempted from https://github.com/uptick/react-keyed-file-browser.

    // handleCreateFiles adds new files to existing files.
    handleCreateFiles = (files, prefix) => {
      this.setState(state => {
        const newFiles = files.map((file) => {
          let newKey = prefix
          if (prefix !== '' && prefix.substring(prefix.length - 1, prefix.length) !== '/') {
            newKey += '/'
          }
          newKey += file.name
          return {
            key: newKey,
            size: file.size,
            // modified: +Moment(),
          }
        })
  
        const uniqueNewFiles = []
        newFiles.map((newFile) => {
          let exists = false
          state.files.map((existingFile) => {
            if (existingFile.key === newFile.key) {
              exists = true
            }
          })
          if (!exists) {
            uniqueNewFiles.push(newFile)
          }
        })
        state.files = state.files.concat(uniqueNewFiles)
        return state
      })
    }

    // handleRenameFile renames an existing file.
    handleRenameFile = (oldKey, newKey) => {
      this.setState(state => {
        const newFiles = []
        state.files.map((file) => {
          if (file.key === oldKey) {
            newFiles.push({
              ...file,
              key: newKey,
            //   modified: +Moment(),
            })
          } else {
            newFiles.push(file)
          }
        })
        state.files = newFiles
        return state
      })
    }

    // handleDeleteFile deletes an existing file.
    handleDeleteFile = (fileKey) => {
      this.setState(state => {
        const newFiles = []
        state.files.map((file) => {
          if (file.key !== fileKey) {
            newFiles.push(file)
          }
        })
        state.files = newFiles
        return state
      })
    }

    // handleCreateFolder creates a new folder.
    handleCreateFolder = (key) => {
      console.log("Creating folder: " + key)
      this.setState(state => {
        state.files = state.files.concat([{
          key: key,
        }])
        return state
      })
    }

    // handleRenameFolder renames an existing folder.
    handleRenameFolder = (oldKey, newKey) => {
      this.setState(state => {
        const newFiles = []
        state.files.map((file) => {
          if (file.key.substr(0, oldKey.length) === oldKey) {
            newFiles.push({
              ...file,
              key: file.key.replace(oldKey, newKey),
            //   modified: +Moment(),
            })
          } else {
            newFiles.push(file)
          }
        })
        state.files = newFiles
        return state
      })
    }

    // handleDeleteFolder deletes an existing folder.
    handleDeleteFolder = (folderKey) => {
      this.setState(state => {
        const newFiles = []
        state.files.map((file) => {
          if (file.key.substr(0, folderKey.length) !== folderKey) {
            newFiles.push(file)
          }
        })
        state.files = newFiles
        return state
      })
    }
  
    render() {
      return (
        <FileBrowser
          files={this.state.files}
          icons={FileExplorerIcons}

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
      )
    }
}
