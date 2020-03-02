import urljoin from 'url-join';
import * as Constants from './Constants';

// GetFileContentsLink returns a URL for the file's contents.
// The endpoint returns the contents with headers suitable for browser download.
export function GetFileContentsLink(file) {
    // FIXME: authorization not being sent
    console.log("Called API method: GetFileContentsLink")
    const url = urljoin(Constants.GetBackendAddress(), Constants.GetBase(), `/files/${file.key}`, '?filter=contents')
    console.log('Computed URL: ' + url)
    return url
}
