import React, { useState } from 'react';
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
    function postLogin() {
        // POST request
        // if result successful
        const token = { accessToken: 'hello' }
        setAuthTokensCallback(token)
        setIsLoggedIn(true);
    }

    if (isLoggedIn) {
        return <Redirect to={referer} />
    }
    // not logged in, return a login page
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
                    <Button variant='primary' type='submit' onClick={postLogin}>Connect</Button>
                </Form>
        </Container>
    );
}
