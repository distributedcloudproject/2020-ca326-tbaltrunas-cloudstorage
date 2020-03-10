import PropTypes from 'prop-types'
import React from 'react'
import Container from 'react-bootstrap/Container'
import Form from 'react-bootstrap/Form'

class FilterRenderer extends React.Component {
  static propTypes = {
    value: PropTypes.string.isRequired,
    updateFilter: PropTypes.func,
  }

  handleFilterChange = (event) => {
    const newValue = this.filterRef.value
    this.props.updateFilter(newValue)
  }

  render() {
    return (
        <Container className='col-sm-4'>
                <Form className='d-flex flex-column justify-content-center'>
                    <Form.Group className='mb-4'>
                        <Form.Control 
                            ref="filter"
                            ref={el => { this.filterRef = el }}
                            as="input"
                            type="input" // note that type must be "input" and not "search" for correct layout
                            placeholder="Filter files"
                            value={this.props.value}
                            onChange={this.handleFilterChange}
                        />
                    </Form.Group>
                </Form>
        </Container>
      )
  }
}

export default FilterRenderer
