import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

export function CloudInfo() {
    console.log('Called API method: CloudInfo');
    const url = urljoin(Constants.GetBase(), '/cloudinfo');
    return axios.get(url);
}
