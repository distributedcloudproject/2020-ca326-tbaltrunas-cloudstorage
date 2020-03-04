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
            <Row className='border border-primary m-2 p-2 d-flex justify-content-around rounded'>
                <CloudInfo />
                <Logout />
            </Row>
            <Container className='d-flex flex-wrap border border-primary rounded'>
                <Col className='m-1 p-1'>
                    <FileExplorer />
                </Col>
            </Container>
        </Container>
    );
}
