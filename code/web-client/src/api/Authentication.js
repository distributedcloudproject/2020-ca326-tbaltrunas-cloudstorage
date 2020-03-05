import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// Login performs a GET request with login details to obtain authorization.
export function Login(username, password) {
    console.log("Called API method: Login");
    const url = urljoin(Constants.GetBase(), '/auth/login');
    return axios.post(url, {
        username: username,
        password: password,
    }, { withCredentials: true } // for cookies
    )
}
