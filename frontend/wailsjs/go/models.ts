export namespace app {
	
	export class AppInfo {
	    Version: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Version = source["Version"];
	        this.Name = source["Name"];
	    }
	}

}

export namespace config {
	
	export class AppConfig {
	    name: string;
	    title: string;
	    defaultTimeoutSeconds: number;
	    maxResponseBodyBytes: number;
	    allowInsecureTLS: boolean;
	    persistRequestHistory: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.defaultTimeoutSeconds = source["defaultTimeoutSeconds"];
	        this.maxResponseBodyBytes = source["maxResponseBodyBytes"];
	        this.allowInsecureTLS = source["allowInsecureTLS"];
	        this.persistRequestHistory = source["persistRequestHistory"];
	    }
	}
	export class AuthConfig {
	    type: string;
	    profileId?: string;
	    allowAuthorizationHeaderOverride: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AuthConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.profileId = source["profileId"];
	        this.allowAuthorizationHeaderOverride = source["allowAuthorizationHeaderOverride"];
	    }
	}
	export class BodyConfig {
	    type: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new BodyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	    }
	}
	export class EndpointView {
	    id: string;
	    name: string;
	    description?: string;
	    groupId?: string;
	    enabled: boolean;
	    method: string;
	    url: string;
	    timeoutSeconds?: number;
	    auth: AuthConfig;
	    headers?: model.KeyValue[];
	    queryParameters?: model.KeyValue[];
	    body: BodyConfig;
	    hasOAuth: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EndpointView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.groupId = source["groupId"];
	        this.enabled = source["enabled"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.auth = this.convertValues(source["auth"], AuthConfig);
	        this.headers = this.convertValues(source["headers"], model.KeyValue);
	        this.queryParameters = this.convertValues(source["queryParameters"], model.KeyValue);
	        this.body = this.convertValues(source["body"], BodyConfig);
	        this.hasOAuth = source["hasOAuth"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EndpointGroup {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new EndpointGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class OAuthProfileView {
	    id: string;
	    name: string;
	    type: string;
	    org_id_uuid: string;
	    client_id: string;
	    clientSecretMasked: string;
	    scope: string;
	    token_url?: string;
	    refreshBeforeExpirySeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new OAuthProfileView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.org_id_uuid = source["org_id_uuid"];
	        this.client_id = source["client_id"];
	        this.clientSecretMasked = source["clientSecretMasked"];
	        this.scope = source["scope"];
	        this.token_url = source["token_url"];
	        this.refreshBeforeExpirySeconds = source["refreshBeforeExpirySeconds"];
	    }
	}
	export class Variable {
	    id: string;
	    base_url: string;
	    environment: string;
	
	    static createFrom(source: any = {}) {
	        return new Variable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.base_url = source["base_url"];
	        this.environment = source["environment"];
	    }
	}
	export class ConfigView {
	    app: AppConfig;
	    variables: Variable[];
	    oauthProfiles: OAuthProfileView[];
	    endpointGroups: EndpointGroup[];
	    endpoints: EndpointView[];
	    configPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app = this.convertValues(source["app"], AppConfig);
	        this.variables = this.convertValues(source["variables"], Variable);
	        this.oauthProfiles = this.convertValues(source["oauthProfiles"], OAuthProfileView);
	        this.endpointGroups = this.convertValues(source["endpointGroups"], EndpointGroup);
	        this.endpoints = this.convertValues(source["endpoints"], EndpointView);
	        this.configPath = source["configPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class OAuthProfile {
	    id: string;
	    name: string;
	    type: string;
	    org_id_uuid: string;
	    client_id: string;
	    client_secret: string;
	    scope: string;
	    token_url?: string;
	    refreshBeforeExpirySeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new OAuthProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.org_id_uuid = source["org_id_uuid"];
	        this.client_id = source["client_id"];
	        this.client_secret = source["client_secret"];
	        this.scope = source["scope"];
	        this.token_url = source["token_url"];
	        this.refreshBeforeExpirySeconds = source["refreshBeforeExpirySeconds"];
	    }
	}
	
	export class ValidationResult {
	    valid: boolean;
	    errors?: string[];
	    warnings?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}

}

export namespace logbus {
	
	export class Entry {
	    time: string;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.level = source["level"];
	        this.message = source["message"];
	    }
	}

}

export namespace model {
	
	export class CertificateView {
	    position: string;
	    subject: string;
	    issuer: string;
	    serialNumber: string;
	    version: number;
	    signatureAlgorithm: string;
	    notBefore: string;
	    notAfter: string;
	    isExpired: boolean;
	    isNotYetValid: boolean;
	    daysToExpiry: number;
	    subjectKeyId: string;
	    authorityKeyId: string;
	    sans: string[];
	    keyAlgorithm: string;
	    keySize: number;
	    publicKeyPem: string;
	    fingerprintSha1: string;
	    fingerprintSha256: string;
	    isCa: boolean;
	    maxPathLength: number;
	    keyUsage: string[];
	    extendedKeyUsage: string[];
	    crlDistributionPoints?: string[];
	    policies?: string[];
	    rawDer: string;
	    pem: string;
	    signatureBytes: string;
	
	    static createFrom(source: any = {}) {
	        return new CertificateView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.subject = source["subject"];
	        this.issuer = source["issuer"];
	        this.serialNumber = source["serialNumber"];
	        this.version = source["version"];
	        this.signatureAlgorithm = source["signatureAlgorithm"];
	        this.notBefore = source["notBefore"];
	        this.notAfter = source["notAfter"];
	        this.isExpired = source["isExpired"];
	        this.isNotYetValid = source["isNotYetValid"];
	        this.daysToExpiry = source["daysToExpiry"];
	        this.subjectKeyId = source["subjectKeyId"];
	        this.authorityKeyId = source["authorityKeyId"];
	        this.sans = source["sans"];
	        this.keyAlgorithm = source["keyAlgorithm"];
	        this.keySize = source["keySize"];
	        this.publicKeyPem = source["publicKeyPem"];
	        this.fingerprintSha1 = source["fingerprintSha1"];
	        this.fingerprintSha256 = source["fingerprintSha256"];
	        this.isCa = source["isCa"];
	        this.maxPathLength = source["maxPathLength"];
	        this.keyUsage = source["keyUsage"];
	        this.extendedKeyUsage = source["extendedKeyUsage"];
	        this.crlDistributionPoints = source["crlDistributionPoints"];
	        this.policies = source["policies"];
	        this.rawDer = source["rawDer"];
	        this.pem = source["pem"];
	        this.signatureBytes = source["signatureBytes"];
	    }
	}
	export class KeyValue {
	    key: string;
	    value: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KeyValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.enabled = source["enabled"];
	    }
	}
	export class RequestInput {
	    requestId: string;
	    endpointId: string;
	    method: string;
	    url: string;
	    headers: KeyValue[];
	    queryParams: KeyValue[];
	    bodyType: string;
	    body: string;
	    timeoutSeconds: number;
	    useOAuth: boolean;
	    oauthProfileId: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.endpointId = source["endpointId"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.headers = this.convertValues(source["headers"], KeyValue);
	        this.queryParams = this.convertValues(source["queryParams"], KeyValue);
	        this.bodyType = source["bodyType"];
	        this.body = source["body"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.useOAuth = source["useOAuth"];
	        this.oauthProfileId = source["oauthProfileId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TLSConnectionView {
	    version: string;
	    cipherSuite: string;
	    cipherSuiteName: string;
	    negotiatedProtocol?: string;
	    serverName: string;
	    resumed: boolean;
	    scts?: string[];
	    ocspStapled: boolean;
	    peerCertificates: number;
	
	    static createFrom(source: any = {}) {
	        return new TLSConnectionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.cipherSuite = source["cipherSuite"];
	        this.cipherSuiteName = source["cipherSuiteName"];
	        this.negotiatedProtocol = source["negotiatedProtocol"];
	        this.serverName = source["serverName"];
	        this.resumed = source["resumed"];
	        this.scts = source["scts"];
	        this.ocspStapled = source["ocspStapled"];
	        this.peerCertificates = source["peerCertificates"];
	    }
	}
	export class TLSInfo {
	    status: string;
	    error?: string;
	    targetHost: string;
	    attemptedServerName?: string;
	    connection?: TLSConnectionView;
	    certificates: CertificateView[];
	    validationSkipped?: boolean;
	    originalError?: string;
	
	    static createFrom(source: any = {}) {
	        return new TLSInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.error = source["error"];
	        this.targetHost = source["targetHost"];
	        this.attemptedServerName = source["attemptedServerName"];
	        this.connection = this.convertValues(source["connection"], TLSConnectionView);
	        this.certificates = this.convertValues(source["certificates"], CertificateView);
	        this.validationSkipped = source["validationSkipped"];
	        this.originalError = source["originalError"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ResponseOutput {
	    requestId: string;
	    statusCode: number;
	    status: string;
	    headers: Record<string, Array<string>>;
	    body: string;
	    bodyTruncated: boolean;
	    durationMs: number;
	    sizeBytes: number;
	    contentType: string;
	    usedOAuth: boolean;
	    tokenFromCache: boolean;
	    tls?: TLSInfo;
	    errorCode: string;
	    errorMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new ResponseOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requestId = source["requestId"];
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.body = source["body"];
	        this.bodyTruncated = source["bodyTruncated"];
	        this.durationMs = source["durationMs"];
	        this.sizeBytes = source["sizeBytes"];
	        this.contentType = source["contentType"];
	        this.usedOAuth = source["usedOAuth"];
	        this.tokenFromCache = source["tokenFromCache"];
	        this.tls = this.convertValues(source["tls"], TLSInfo);
	        this.errorCode = source["errorCode"];
	        this.errorMessage = source["errorMessage"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace oauth {
	
	export class TokenStatus {
	    profileId: string;
	    hasToken: boolean;
	    fromCache: boolean;
	    // Go type: time
	    expiresAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new TokenStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profileId = source["profileId"];
	        this.hasToken = source["hasToken"];
	        this.fromCache = source["fromCache"];
	        this.expiresAt = this.convertValues(source["expiresAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

