import React from 'react'
import PropTypes from 'prop-types'
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';
import ListGroup from 'react-bootstrap/ListGroup'

const ActionItem = ListGroup.Item;
const ActionButton = Button;
const ActionContainer = Container;
const ActionListContainer = ListGroup;

const ActionItemClass = 'border-0 bg-transparent p-1';

function ActionRenderer(props) {
  const {
    selectedItems,
    isFolder,
    icons,
    nameFilter,

    canCreateFolder,
    onCreateFolder,

    canRenameFile,
    onRenameFile,

    canRenameFolder,
    onRenameFolder,

    canDeleteFile,
    onDeleteFile,

    canDeleteFolder,
    onDeleteFolder,

    canDownloadFile,
    onDownloadFile,

  } = props

  let actions = []

  // Nothing selected: We're in the 'root' folder. Only allowed action is adding a folder.
  // And uploading a file.
  if (!selectedItems.length) {
    // Creating folders
    if (canCreateFolder && !nameFilter) {
      actions.push(
        <ActionItem key="action-add-folder" className={ActionItemClass} /* className="list-group-item" */ >
          <ActionButton
            onClick={onCreateFolder}
            href="#"
            role="button"
          >
            {icons.Folder}
            &nbsp;Add Folder
          </ActionButton>
        </ActionItem>
      )
      return (<ActionListContainer horizontal={'md'} className="item-actions bg-transparent">{actions}</ActionListContainer>)
    }
    return (<ActionContainer className="item-actions">&nbsp;</ActionContainer>)
  }

  // Something is selected. Build custom actions depending on what it is.

  // Selected item has an active action against it. Disable all other actions.
  let selectedItemsAction = selectedItems.filter(item => item.action)
  if (selectedItemsAction.length === selectedItems.length && [... new Set(selectedItemsAction)].length === 1) {
    let actionText
    switch (selectedItemsAction[0].action) {
      case 'delete':
        actionText = 'Deleting ...'
        break

      case 'rename':
        actionText = 'Renaming ...'
        break

      default:
        actionText = 'Moving ...'
        break
    }

    return (
      // TODO: Enable plugging in custom spinner.
      <ActionContainer className="item-actions">
        {icons.Loading} {actionText}
      </ActionContainer>
    )
  }

  // Downloading
  if (!isFolder && canDownloadFile) {
    // canDownloadFile is true if the file has more than 0 bytes in size.
    actions.push(
      <ActionItem key="action-download" className={ActionItemClass}>
        <ActionButton
          onClick={onDownloadFile}
          href="#"
          role="button"
        >
          {icons.Download}
          &nbsp;Download
        </ActionButton>
      </ActionItem>
    )
  }

  // Renaming
  let itemsWithoutKeyDerived = selectedItems.find(item => !item.keyDerived)
  if (!itemsWithoutKeyDerived && !isFolder && canRenameFile && selectedItems.length === 1) {
    // File rename
    actions.push(
      <ActionItem key="action-rename" className={ActionItemClass}>
        <ActionButton
          onClick={onRenameFile}
          href="#"
          role="button"
        >
          {icons.Rename}
          &nbsp;Rename
        </ActionButton>
      </ActionItem>
    )
  } else if (!itemsWithoutKeyDerived && isFolder && canRenameFolder) {
    //Folder rename
    actions.push(
      <ActionItem key="action-rename" className={ActionItemClass}>
        <ActionButton
          onClick={onRenameFolder}
          href="#"
          role="button"
        >
          {icons.Rename}
          &nbsp;Rename
        </ActionButton>
      </ActionItem>
    )
  }

  // Deleting
  if (!itemsWithoutKeyDerived && !isFolder && canDeleteFile) {
    // File delete
    actions.push(
      <ActionItem key="action-delete" className={ActionItemClass}>
        <ActionButton
          onClick={onDeleteFile}
          href="#"
          role="button"
        >
          {icons.Delete}
          &nbsp;Delete
        </ActionButton>
      </ActionItem>
    )
  } else if (!itemsWithoutKeyDerived && isFolder && canDeleteFolder) {
    // Folder delete
    actions.push(
      <ActionItem key="action-delete" className={ActionItemClass}>
        <ActionButton
          onClick={onDeleteFolder}
          href="#"
          role="button"
        >
          {icons.Delete}
          &nbsp;Delete
        </ActionButton>
      </ActionItem>
    )
  }

  // Creating folders
  if (isFolder && canCreateFolder && !nameFilter) {
    actions.push(
      <ActionItem key="action-add-folder" className={ActionItemClass}>
        <ActionButton
          onClick={onCreateFolder}
          href="#"
          role="button"
        >
          {icons.Folder}
          &nbsp;Add Subfolder
        </ActionButton>
      </ActionItem>
    )
  }

  if (!actions.length) {
    return (<ActionContainer className="item-actions">&nbsp;</ActionContainer>)
  }
  return (<ActionListContainer horizontal={'md'} className="item-actions bg-transparent">{actions}</ActionListContainer>)
}

ActionRenderer.propTypes = {
  selectedItems: PropTypes.arrayOf(PropTypes.object),
  isFolder: PropTypes.bool,
  icons: PropTypes.object,
  nameFilter: PropTypes.string,

  canCreateFolder: PropTypes.bool,
  onCreateFolder: PropTypes.func,

  canRenameFile: PropTypes.bool,
  onRenameFile: PropTypes.func,

  canRenameFolder: PropTypes.bool,
  onRenameFolder: PropTypes.func,

  canDeleteFile: PropTypes.bool,
  onDeleteFile: PropTypes.func,

  canDeleteFolder: PropTypes.bool,
  onDeleteFolder: PropTypes.func,

  canDownloadFile: PropTypes.bool,
  onDownloadFile: PropTypes.func,
}

ActionRenderer.defaultProps = {
  selectedItems: [],
  isFolder: false,
  icons: {},
  nameFilter: '',

  canCreateFolder: false,
  onCreateFolder: null,

  canRenameFile: false,
  onRenameFile: null,

  canRenameFolder: false,
  onRenameFolder: null,

  canDeleteFile: false,
  onDeleteFile: null,

  canDeleteFolder: false,
  onDeleteFolder: null,

  canDownloadFile: false,
  onDownloadFile: null,
}

export default ActionRenderer
