import url from 'url';

// TODO: environment variables?
let base = url.format({
    protocol: 'http',
    hostname: '127.0.0.1',
    port: '9001',
});

let endpointVersion = 'api/v1';

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
