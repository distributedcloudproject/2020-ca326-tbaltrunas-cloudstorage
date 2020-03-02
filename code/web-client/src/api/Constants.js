import urljoin from 'url-join';

// We use a proxy for the backend and absolute URL's in code

let endpointVersion = 'v1';
let base = urljoin('/api/', endpointVersion)

function GetBase() {
    return base
}

function SetBase(value) {
    base = value
}

function GetEndpointVersion() {
    return endpointVersion
}

function SetEndpointVersion(value) {
    endpointVersion = value
}

export {
    GetBase,
    SetBase,
    GetEndpointVersion,
    SetEndpointVersion,
}
