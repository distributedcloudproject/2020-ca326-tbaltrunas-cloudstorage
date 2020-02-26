import React, { useState } from 'react';
import axios from 'axios';
import { Redirect } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import { useAuthContext} from '../context/Authentication';

export default function Login(props) {
    // Add state to our Login component.
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [isError, setIsError] = useState(false);
    const [networkAddr, setNetworkAddr] = useState('');

    // Get callback variable from context.
    const { setAuthTokensCallback } = useAuthContext();

    // Get referer if available (to go back to the page that the user wanted to access).
    // const referer = props.location.state.referer || '/';
    const referer = '/';

    // Methods (closures)
    function postLogin(e) {
        console.log('Post login called')
        e.preventDefault();

        // POST request
        // if result successful
        const path = 'http://127.0.0.1:9001/auth'
        console.log('Getting: ' + path)
        axios.get(path)
            .then(response => {
                console.log(response)
                console.log(response.status)
                console.log(response.data)
                if (response.status === 200) {
                    setAuthTokensCallback(response.data)
                    setIsLoggedIn(true);            
                } else {
                    console.log(response.status)
                    setIsError(true)
                }
            })
            .catch(error => {
                console.log(error)
                setIsError(true)
            })
    }

    console.log("Error: " + isError)

    if (isLoggedIn) {
        return <Redirect to={referer} />
    }

    return (
        <Container className='p-5 col-md-4'>
                <Form className='d-flex flex-column justify-content-center'>
                    <Form.Group controlId='formGroupNetworkAddress' className='mb-4'>
                        <Form.Label>Network Address</Form.Label>
                        <Form.Control 
                            type='text' 
                            placeholder='Network Address' 
                            value={networkAddr}
                            onChange={e => {
                                setNetworkAddr(e.target.value)
                            }}
                        />
                    </Form.Group>
                    <Button variant='primary' type='submit' onClick={e => {postLogin(e)}} >Connect</Button>
                </Form>
        </Container>
    );
}
