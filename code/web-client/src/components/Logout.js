import React from 'react';
import Button from 'react-bootstrap/Button';
import { useAuthContext } from '../context/Authentication';

export default function Logout(props) {
    const { setAuthTokensCallback } = useAuthContext();

    function logOut() {
        setAuthTokensCallback()
    }

    return (
        <Button variant='warning' onClick={logOut}>Disconnect</Button>
    )
}
