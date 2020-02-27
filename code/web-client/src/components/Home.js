import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Logout from './Logout';
import FileExplorer from './FileExplorer';

export default function Home(props) {
    return (
        <Container className='d-flex flex-column border p-1'>
            <Row className='border m-2 p-2 d-flex justify-content-around'>
                <p>Network Name</p>
                <Logout />
            </Row>
            <Container className='d-flex flex-wrap border'>
                <Col className='border m-1 p-1'>
                    <p>Controls</p>
                </Col>
                <Col className='border m-1 p-1' xs={10}>
                    <FileExplorer />
                </Col>
            </Container>
        </Container>
    );
}
