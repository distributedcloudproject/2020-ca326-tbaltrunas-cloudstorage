import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// Login performs a GET request with login details to obtain authorization.
export function Login() {
    console.log("Called API method: Login");
    const url = urljoin(Constants.GetBase(), '/auth');
    // TODO: add username/password query params
    return axios.get(url, {withCredentials: true})
}
