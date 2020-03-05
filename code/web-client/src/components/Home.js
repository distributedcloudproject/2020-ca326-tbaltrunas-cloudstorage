import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Logout from './Logout';
import CloudInfo from './CloudInfo';
import FileExplorer from './FileExplorer';

export default function Home(props) {
    return (
        <Container className='d-flex flex-column p-2'>
            {/* TODO: reuse container "styling" */}
            <Container className='border rounded border-primary shadow-sm m-2 p-2 d-flex justify-content-around'>
                <Col className='col-md-10 d-flex justify-content-center'>
                    <CloudInfo />
                </Col>
                <Col className='d-flex justify-content-center'>
                    <Logout />
                </Col>
            </Container>
            <Container className='border border-primary rounded shadow-sm m-2 p-2 d-flex flex-wrap'>
                <Col className='m-1 p-1'>
                    <FileExplorer />
                </Col>
            </Container>
        </Container>
    );
}
