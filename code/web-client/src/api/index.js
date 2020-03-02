import * as APIConstants from './Constants';
import * as FilesAPI from './Files';
import * as FilesDownloadAPI from './FilesDownload';
import * as AuthenticationAPI from './Authentication';
import Cookies from 'js-cookie';
import axios from 'axios';

// FIXME: may not need this for all requests.
axios.defaults.headers.common['Authorization'] = `Bearer ${Cookies.get('access_token')}` 

// axios.interceptors.request.use((config) => {
//     const token = Cookies.get('access_token');
//     if (token) {
//         config.headers.Authorization = `Bearer ${token}`;
//     }
//     return config;
// })

export {
    APIConstants,
    FilesAPI,
    FilesDownloadAPI,
    AuthenticationAPI,
}
