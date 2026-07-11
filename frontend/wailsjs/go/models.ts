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

export namespace model {
	
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

}

export namespace oauth {
	
	export class TokenStatus {
	    ProfileID: string;
	    HasToken: boolean;
	    // Go type: time
	    ExpiresAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new TokenStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ProfileID = source["ProfileID"];
	        this.HasToken = source["HasToken"];
	        this.ExpiresAt = this.convertValues(source["ExpiresAt"], null);
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

