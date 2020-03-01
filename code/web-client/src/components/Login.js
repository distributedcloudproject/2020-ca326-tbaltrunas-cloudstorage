import React, { useState } from 'react';
import { Redirect } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import { useAuthContext} from '../context/Authentication';
import { AuthenticationAPI } from '../api';

export default function Login(props) {
    // Add state to our Login component.
    const [isLoggedIn, setIsLoggedIn] = useState(false);
    const [isError, setIsError] = useState(false);
    const [networkAddr, setNetworkAddr] = useState('');

    // Get callback variable from context.
    const { setIsAuthenticatedCallback } = useAuthContext();

    // Get referer if available (to go back to the page that the user wanted to access).
    // const referer = props.location.state.referer || '/';
    const referer = '/';

    // Methods (closures)
    async function postLogin(e) {
        console.log('postLogin called')
        e.preventDefault();

        try {
            const response = await AuthenticationAPI.Login();
            if (response.status === 200) {
                setIsAuthenticatedCallback(true);
                setIsLoggedIn(true);
            } else {
                console.log(response.status);
                setIsError(true);
            }
        } catch (error) {
            console.log(error);
            setIsError(true);
        }
    }

    console.log("Error: " + isError)

    if (isLoggedIn) {
        return <Redirect to={referer} />
    }

    return (
        <Container className='p-5 col-md-4'>
                <Form className='d-flex flex-column justify-content-center'>
                    <Form.Group controlId='formGroupNetworkAddress' className='mb-4'>
                        <Form.Label className='text-white'>Network Address</Form.Label>
                        <Form.Control 
                            type='text' 
                            placeholder='Network Address' 
                            value={networkAddr}
                            onChange={e => {
                                setNetworkAddr(e.target.value)
                            }}
                        />
                    </Form.Group>
                    <Button variant='primary' type='submit' onClick={postLogin} >Connect</Button>
                </Form>
        </Container>
    );
}
