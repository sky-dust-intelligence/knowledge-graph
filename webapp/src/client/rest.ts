import {ClientResponse, Options} from "../types/client";
import {ServerError} from "../types/errors";


const HEADER_AUTH = 'Authorization';
const HEADER_BEARER = 'BEARER';
const HEADER_CONTENT_TYPE = 'Content-Type';
const HEADER_REQUESTED_WITH = 'X-Requested-With';
const HEADER_X_CSRF_TOKEN = 'X-CSRF-Token';


export class Rest {
    apiVersion = '/api/v1';
    url = '';
    token = '';
    csrf = '';

    getBaseRoute() {
        return `${this.url}${this.apiVersion}`;
    }

    setToken(token: string) {
        this.token = token;
    }

    doFetch = async <ClientDataResponse>(url: string, options: Options): Promise<ClientDataResponse> => {
        const {data} = await this.doFetchWithResponse<ClientDataResponse>(url, options);

        return data;
    };

    doFetchWithResponse = async <ClientDataResponse>(url: string, options: Options): Promise<ClientResponse<ClientDataResponse>> => {
        const response = await fetch(url, this.getOptions(options));
        const headers = parseAndMergeNestedHeaders(response.headers);

        let data;
        try {
            data = await response.json();
        } catch (err) {
            throw new ClientError({
                message: 'Received invalid response from the server.',
                url,
            });
        }

        if (response.ok) {
            return {
                response,
                headers,
                data,
            };
        }

        const msg = data.msg || '';

        throw new ClientError({
            message: msg,
            server_error_id: data.id,
            status_code: data.status_code,
            url,
        });
    };

    getOptions(options: Options) {
        const newOptions: Options = {...options};

        const headers: {[x: string]: string} = {
            [HEADER_REQUESTED_WITH]: 'XMLHttpRequest',
        };

        if (this.token) {
            headers[HEADER_AUTH] = `${HEADER_BEARER} ${this.token}`;
        }

        const csrfToken = this.csrf || getCSRFFromCookie();
        if (options.method && options.method.toLowerCase() !== 'get' && csrfToken) {
            headers[HEADER_X_CSRF_TOKEN] = csrfToken;
        }

        newOptions.credentials = 'include';

        if (!headers[HEADER_CONTENT_TYPE] && options.body) {
            // when the body is an instance of FormData we let browser set the Content-Type header generated by FormData interface with correct boundary
            if (!(options.body instanceof FormData)) {
                headers[HEADER_CONTENT_TYPE] = 'application/json';
            }
        }

        if (newOptions.headers) {
            Object.assign(headers, newOptions.headers);
        }

        return {
            ...newOptions,
            headers,
        };
    }
}

function getCSRFFromCookie() {
    if (typeof document !== 'undefined' && typeof document.cookie !== 'undefined') {
        const cookies = document.cookie.split(';');
        for (let i = 0; i < cookies.length; i++) {
            const cookie = cookies[i].trim();
            if (cookie.startsWith('KGCSRF=')) {
                return cookie.replace('KGCSRF=', '');
            }
        }
    }
    return '';
}

function parseAndMergeNestedHeaders(originalHeaders: any) {
    const headers = new Map();
    let nestedHeaders = new Map();
    originalHeaders.forEach((val: string, key: string) => {
        const capitalizedKey = key.replace(/\b[a-z]/g, (l) => l.toUpperCase());
        let realVal = val;
        if (val && val.match(/\n\S+:\s\S+/)) {
            const nestedHeaderStrings = val.split('\n');
            realVal = nestedHeaderStrings.shift() as string;
            const moreNestedHeaders = new Map(
                nestedHeaderStrings.map((h: any) => h.split(/:\s/)),
            );
            nestedHeaders = new Map([...nestedHeaders, ...moreNestedHeaders]);
        }
        headers.set(capitalizedKey, realVal);
    });
    return new Map([...headers, ...nestedHeaders]);
}

export class ClientError extends Error implements ServerError {
    url?: string;
    serverErrorID?: string;
    statusCode?: number;

    constructor(data: ServerError) {
        super(data.message + ': ' + data.url || '');

        this.message = data.message;
        this.url = data.url;
        this.serverErrorID = data.server_error_id;
        this.statusCode = data.status_code;

        // Ensure message is treated as a property of this class when object spreading. Without this,
        // copying the object by using `{...error}` would not include the message.
        Object.defineProperty(this, 'message', {enumerable: true});
    }
}