import React from 'react';
import { CloudAPI } from '../api';

// CloudInfo displays simple information about the cloud network we are connected to.
export default class CloudInfo extends React.Component {
    state = {
        NetworkName: 'Network Name',
    }

    async componentDidMount() {
        try {
            const cloudInfo = await CloudAPI.CloudInfo();
            console.log('Got cloud info: ' + cloudInfo)
            this.setState({ NetworkName: cloudInfo.networkname });
        } catch (error) {
            console.error(error)
        }
    }

    render() {
        return (
            <p>{this.state.NetworkName}</p>
        )
    }
}
