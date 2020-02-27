import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Logout from './Logout';
import FileExplorer from './FileExplorer';

export default function Home(props) {
    return (
        <Container className='d-flex flex-column p-2'>
            <Row className='border border-primary m-2 p-2 d-flex justify-content-around rounded'>
                <p>Network Name</p>
                <Logout />
            </Row>
            <Container className='d-flex flex-wrap border border-primary rounded'>
                <Col className='m-1 p-1'>
                    <p>Controls</p>
                </Col>
                <Col className='m-1 p-1' xs={10}>
                    <FileExplorer />
                </Col>
            </Container>
        </Container>
    );
}
