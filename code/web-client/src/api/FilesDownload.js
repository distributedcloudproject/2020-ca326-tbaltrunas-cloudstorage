import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// GetFileDownloadLink returns a temporary URL to download the file.
// The link points to an endpoint that returns the contents with suitable download headers.
export function GetFileDownloadLink(file) {
    // FIXME: authorization not being sent
    console.log("Called API method: GetFileDownloadLink")
    // TODO: actual call
    const url = urljoin(Constants.GetBase(), `/downloadlink/${file.key}`)
    return axios.get(url)
}
