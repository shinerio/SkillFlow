export namespace backup {
	
	export class RemoteFile {
	    path: string;
	    size: number;
	    isDir: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RemoteFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	    }
	}

}

export namespace config {
	
	export class CloudConfig {
	    provider: string;
	    enabled: boolean;
	    bucketName: string;
	    remotePath: string;
	    credentials: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new CloudConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.enabled = source["enabled"];
	        this.bucketName = source["bucketName"];
	        this.remotePath = source["remotePath"];
	        this.credentials = source["credentials"];
	    }
	}
	export class ToolConfig {
	    name: string;
	    skillsDir: string;
	    enabled: boolean;
	    custom: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.skillsDir = source["skillsDir"];
	        this.enabled = source["enabled"];
	        this.custom = source["custom"];
	    }
	}
	export class AppConfig {
	    skillsStorageDir: string;
	    defaultCategory: string;
	    tools: ToolConfig[];
	    cloud: CloudConfig;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillsStorageDir = source["skillsStorageDir"];
	        this.defaultCategory = source["defaultCategory"];
	        this.tools = this.convertValues(source["tools"], ToolConfig);
	        this.cloud = this.convertValues(source["cloud"], CloudConfig);
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

export namespace install {
	
	export class SkillCandidate {
	    Name: string;
	    Path: string;
	    Installed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SkillCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.Installed = source["Installed"];
	    }
	}

}


	export class ProxyConfig {
	    Mode: string;
	    URL: string;

	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Mode = source["Mode"];
	        this.URL = source["URL"];
	    }
	}

export namespace skill {
	
	export class Skill {
	    ID: string;
	    Name: string;
	    Path: string;
	    Category: string;
	    Source: string;
	    SourceURL: string;
	    SourceSubPath: string;
	    SourceSHA: string;
	    LatestSHA: string;
	    // Go type: time
	    InstalledAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: time
	    LastCheckedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.Category = source["Category"];
	        this.Source = source["Source"];
	        this.SourceURL = source["SourceURL"];
	        this.SourceSubPath = source["SourceSubPath"];
	        this.SourceSHA = source["SourceSHA"];
	        this.LatestSHA = source["LatestSHA"];
	        this.InstalledAt = this.convertValues(source["InstalledAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.LastCheckedAt = this.convertValues(source["LastCheckedAt"], null);
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

	export class SkillMeta {
	    Name: string;
	    Description: string;
	    ArgumentHint: string;
	    AllowedTools: string;
	    Context: string;
	    DisableModelInvocation: boolean;

	    static createFrom(source: any = {}) {
	        return new SkillMeta(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.ArgumentHint = source["ArgumentHint"];
	        this.AllowedTools = source["AllowedTools"];
	        this.Context = source["Context"];
	        this.DisableModelInvocation = source["DisableModelInvocation"];
	    }
	}


}

