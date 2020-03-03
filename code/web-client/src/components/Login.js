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
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');

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
            const response = await AuthenticationAPI.Login(username, password);
            if (response.status === 200) {
                await setIsAuthenticatedCallback(true);
                // FIXME: problem with access_token cookie remaining undefined / not being updated in time
                // user needs to refresh home page for the request with the updated authorization to be sent
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

    // TODO: form validation
    return (
        <Container className='p-5 col-md-4'>
                <Form className='d-flex flex-column justify-content-center'>
                    <Form.Group controlId='formGroupUsername' className='mb-4'>
                        <Form.Label className='text-white'>Username</Form.Label>
                        <Form.Control 
                            type='text' 
                            placeholder='Username' 
                            value={username}
                            onChange={e => {
                                setUsername(e.target.value)
                            }}
                        />
                    </Form.Group>
                    <Form.Group controlId='formGroupPassword' className='mb-4'>
                        <Form.Label className='text-white'>Password</Form.Label>
                        <Form.Control 
                            type='password' 
                            placeholder='Password' 
                            value={password}
                            onChange={e => {
                                setPassword(e.target.value)
                            }}
                        />
                    </Form.Group>
                    <Button variant='primary' type='submit' onClick={postLogin} >Connect</Button>
                </Form>
        </Container>
    );
}
