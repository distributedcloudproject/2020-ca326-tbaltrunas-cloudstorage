import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

export async function CloudInfo() {
    console.log('Called API method: CloudInfo');
    const url = urljoin(Constants.GetBase(), '/cloudinfo');
    try {
        const response = await axios.get(url)
        if (response.status !== 200) {
            console.error(response.status)
        }
        return response.data
    } catch (error) {
        console.error(error)
    }
}
