// UI handlers adapted from https://github.com/uptick/react-keyed-file-browser.

export function UpdateUICreateFiles(state, files, prefix) {
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
}

export function UpdateUIRenameFile(state, oldKey, newKey) {
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
}

export function UpdateUIDeleteFile(state, fileKey) {
    console.log(state)
    console.log(fileKey)
    const newFiles = []
    state.files.map((file) => {
      console.log(file.key === fileKey)
      if (file.key !== fileKey) {
        newFiles.push(file)
      }
    })
    state.files = newFiles
    console.log(state)
    return state
}

export function UpdateUICreateFolder(state, key) {
  state.files = state.files.concat([{
    key: key,
  }])
  return state
}

export function UpdateUIRenameFolder(state, oldKey, newKey) {
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
}

export function UpdateUIDeleteFolder(state, folderKey) {
  const newFiles = []
  state.files.map((file) => {
    if (file.key.substr(0, folderKey.length) !== folderKey) {
      newFiles.push(file)
    }
  })
  state.files = newFiles
  return state
}
