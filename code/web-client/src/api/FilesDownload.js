import axios from 'axios';
import urljoin from 'url-join';
import * as Constants from './Constants';

// GetFileDownloadLink returns a temporary URL to download the file.
// The link points to an endpoint that returns the contents with suitable download headers.
export async function GetFileDownloadLink(fileID) {
    // FIXME: authorization not being sent
    console.log("Called API method: GetFileDownloadLink")
    // TODO: actual call
    const url = urljoin(Constants.GetBase(), `/downloadlink/?fileKey=${fileID}`)
    try {
        const response = await axios.get(url)
        if (response.status !== 200) {
            console.error(response.status);
        }
        const endpointURL = response.data; // represents file download link without the backend address
        return urljoin(Constants.GetBackendAddress(), Constants.GetBase(), endpointURL);    
    } catch (error) {
        console.error(error)
    }
    
}
