import React from 'react';
import Container from 'react-bootstrap/Container';
import Logout from './Logout';

export default function Home(props) {
    return (
        <Container className='p-5 col-md-2'>
            <Logout/>
        </Container>
    );
}
