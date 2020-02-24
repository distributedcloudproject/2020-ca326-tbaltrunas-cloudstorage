import React, { useState } from 'react';
import { Redirect } from 'react-router-dom';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import { useAuthContext} from '../context/Authentication';

export default function Login(props) {
    // Add state to our Login component.
    const [isLoggedIn, setLoggedIn] = useState(false);
    const [isError, setIsError] = useState(false);
    const [networkAddr, setNetworkAddr] = useState("");

    // Get callback variable from context.
    const { setAuthTokensCallback } = useAuthContext();

    // Get referer if available (to go back to the page that the user wanted to access).
    const referer = props.location.state.referer || '/';

    // Methods (closures)
    function postLogin() {
        // POST request
        // if result successful
        const token = { accessToken: "hello" }
        setAuthTokensCallback(token)
        setLoggedIn(true);
    }

    if (isLoggedIn) {
        return <Redirect to={referer} />
    }
    // not logged in, return a login page
    return (
        <Container>
                <Form>
                    <Form.Group controlId="formGroupNetworkAddress">
                        <Form.Label>Network Address</Form.Label>
                        <Form.Control 
                            type="text" 
                            placeholder="Network Address" 
                            value={networkAddr}
                            onChange={e => {
                                setNetworkAddr(e.target.value)
                            }}
                        />
                    </Form.Group>
                    <Form.Group controlID="formGroupSubmit">
                        <Button variant="primary" onClick={postLogin}>Connect</Button>
                    </Form.Group>
                </Form>
        </Container>
    );
}
