import PropTypes from 'prop-types'
import React from 'react'
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
      <Form>
        <Form.Group>
            <Form.Control
                ref="filter"
                ref={el => { this.filterRef = el }}
                as="input"
                type="search"
                placeholder="Filter files"
                value={this.props.value}
                onChange={this.handleFilterChange}
            />
        </Form.Group>
      </Form>
    )
  }
}

export default FilterRenderer
