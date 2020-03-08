import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';

export default function About(props) {
    return (
        <div>
            <Container>
                <Container>
                    <h1 className='text-primary'>Contact Us</h1>
                    <p>Tomas Baltrunas & Bartosz Śwituszak</p>
                </Container>
                <Container>
                    <h1 className='text-primary'>Credits</h1>
                    <h2>Project icon</h2>
                    <p><a>https://www.freepik.com/free-icon/cloud-computing_913701.htm</a></p>
                    <p>Icons made by <a href="https://www.flaticon.com/authors/freepik" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon"> www.flaticon.com</a></p>
                </Container>
            </Container>
        </div>
    );
}



