import React from 'react';
import Button from 'react-bootstrap/Button';
import { useAuthContext } from '../context/Authentication';

const logoutButtonText = 'Logout';

export default function Logout(props) {
    const { setIsAuthenticatedCallback } = useAuthContext();

    function logOut() {
        setIsAuthenticatedCallback(false)
    }

    return (
        <Button variant='warning' onClick={logOut}>{logoutButtonText}</Button>
    )
}
