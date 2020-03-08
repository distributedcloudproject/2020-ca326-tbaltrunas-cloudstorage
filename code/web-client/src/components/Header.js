import React from 'react';
import Navbar from 'react-bootstrap/Navbar'

const title = 'Distributed Cloud Storage';

export default function Header(props) {
    return (
        <Navbar bg='light' className='navbar-static-top d-flex justify-content-start'>
            <Navbar.Brand 
                className='d-flex align-items-end justify-content-between'
                href="/" >
                <img 
                    src={process.env.PUBLIC_URL + 'logo.svg'} 
                    alt='icon distributed cloud storage' 
                    width='35'
                    height='35'
                    className='p-1 d-inline-block align-top' />
                <span className='h5'>{title}</span>
            </Navbar.Brand>
        </Navbar>
    );
}
