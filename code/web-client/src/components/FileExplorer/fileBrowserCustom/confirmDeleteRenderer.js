import React from 'react'
import PropTypes from 'prop-types'
import Container from 'react-bootstrap/Container'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'

function ConfirmDeletionRenderer(props) {
  const {
    children,
    handleDeleteSubmit,
    handleFileClick,
    url,
  } = props

  return (
    <Form className="deleting" onSubmit={handleDeleteSubmit}>
      <a
        href={url}
        download="download"
        onClick={handleFileClick}
      >
        {children}
      </a>
      <Container className='m-2'>
        <Button type="submit">
          Confirm Deletion
        </Button>
      </Container>
    </Form>
  )
}

ConfirmDeletionRenderer.propTypes = {
  children: PropTypes.node,
  handleDeleteSubmit: PropTypes.func,
  handleFileClick: PropTypes.func,
  url: PropTypes.string,
}

ConfirmDeletionRenderer.defaultProps = {
  url: '#',
}

export default ConfirmDeletionRenderer
