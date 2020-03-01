// We use a proxy for the backend and absolute URL's in code
let base = '/'

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
