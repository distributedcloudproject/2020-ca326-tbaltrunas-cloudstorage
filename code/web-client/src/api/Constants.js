import urljoin from 'url-join';

// We use a proxy for the backend and absolute URL's in code

// this shoud not be needed in most code.
let backendAddress = 'https://localhost:9443'
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

function GetBackendAddress() {
    return backendAddress;
}

export {
    GetBase,
    SetBase,
    GetEndpointVersion,
    SetEndpointVersion,
    GetBackendAddress,
}
