import * as FilesAPI from './Files';
import * as AuthenticationAPI from './Authentication';
import Cookies from 'js-cookie';
import axios from 'axios';

// FIXME: may not need this for all requests.
axios.interceptors.request.use((config) => {
    const token = Cookies.get('access_token');
    if (token) {
        config.headers.Authorization = `Bearer: ${token}`;
    }
    return config;
})

export {
    FilesAPI,
    AuthenticationAPI,
}
