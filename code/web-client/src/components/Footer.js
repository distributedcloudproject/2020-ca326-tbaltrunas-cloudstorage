import React from 'react';
import Navbar from 'react-bootstrap/Navbar'
import Nav from 'react-bootstrap/Nav';

const copyrightYear = 2020;
const copyrightNotice = `Copyright © ${copyrightYear} 
Distributed Cloud Storage, 
Tomas Baltrunas & Bartosz Śwituszak. 
All rights reserved`;

export default function Footer(props) {
    return (
        <Navbar bg='light' fixed='bottom' className='d-flex'>
            <Navbar.Brand href='/about' >
                <span className='text-wrap h6'>{copyrightNotice}</span>
            </Navbar.Brand>
            <Nav className=''>
                <Nav.Link href='/'>Home</Nav.Link>
                <Nav.Link href='/about'>About</Nav.Link>
            </Nav>
        </Navbar>
    );
}
