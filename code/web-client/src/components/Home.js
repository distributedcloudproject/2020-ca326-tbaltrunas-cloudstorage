import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Logout from './Logout';

export default function Home(props) {
    return (
        <Container className='d-flex flex-column border p-1'>
            <Row className='border m-2 p-2'>
                <Col className='border' large={10}>
                    <p>Network Name</p>
                </Col>
                <Col className='border' large={10}>
                    <Logout />
                </Col>
            </Row>
            <Container className='d-flex flex-wrap border'>
                <Col className='border m-1 p-1'>
                    <p>Controls</p>
                </Col>
                <Col className='border m-1 p-1' xs={10}>
                    <p>File browser</p>
                    <p>File</p>
                    <p>File</p>
                    <p>File</p>
                    <p>File</p>
                    <p>File</p>
                    <p>File</p>
                </Col>
            </Container>
        </Container>
    );
}
