import React from 'react';
import Navbar from 'react-bootstrap/Navbar'

const title = 'Distributed Cloud Storage';

export default function Header(props) {
    return (
        <Navbar bg='light' className='navbar-static-top d-flex justify-content-start'>
            <Navbar.Brand 
                className='d-flex align-items-center justify-content-between'
                href="/" >
                <img 
                    src={process.env.PUBLIC_URL + 'dcloud.svg'} 
                    alt='icon distributed cloud storage' 
                    width='35'
                    height='35'
                    className='d-inline-block align-top icon-button' />
                <span>{title}</span>
            </Navbar.Brand>
        </Navbar>
    );
}
