import PropTypes from 'prop-types'
import React from 'react'
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';

class DetailRenderer extends React.Component {
  static propTypes = {
    file: PropTypes.shape({
      key: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      extension: PropTypes.string.isRequired,
      url: PropTypes.string,
    }).isRequired,
    close: PropTypes.func,
  }

  handleCloseClick = (event) => {
    if (event) {
      event.preventDefault()
    }
    this.props.close()
  }

  render() {
    let name = this.props.file.key.split('/')
    name = name.length ? name[name.length - 1] : ''

    return (
      <Container className='border rounded border-secondary p-3 pl-3'>
        <h2>Item Detail</h2>
        <dl>
          <dt>Key</dt>
          <dd>{this.props.file.key}</dd>

          <dt>Name</dt>
          <dd>{name}</dd>
        </dl>
        <Button href="#" onClick={this.handleCloseClick}>
            Close
        </Button>
      </Container>
    )
  }
}

export default DetailRenderer
