export namespace main {
	
	export class ComboPage {
	    total: number;
	    combos: store.ComboResult[];
	
	    static createFrom(source: any = {}) {
	        return new ComboPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.combos = this.convertValues(source["combos"], store.ComboResult);
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
	export class ConfigDetail {
	    id: number;
	    name: string;
	    shapeId: number;
	    cartridgeId: number;
	    hash: string;
	    grid: string;
	    sortOrder: number;
	    total: number;
	    cached: boolean;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.shapeId = source["shapeId"];
	        this.cartridgeId = source["cartridgeId"];
	        this.hash = source["hash"];
	        this.grid = source["grid"];
	        this.sortOrder = source["sortOrder"];
	        this.total = source["total"];
	        this.cached = source["cached"];
	        this.warnings = source["warnings"];
	    }
	}
	export class CreateResult {
	    config: store.Config;
	    dupOf: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], store.Config);
	        this.dupOf = source["dupOf"];
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
	export class PieceInfo {
	    name: string;
	    shape: string;
	
	    static createFrom(source: any = {}) {
	        return new PieceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.shape = source["shape"];
	    }
	}

}

export namespace store {
	
	export class Cartridge {
	    id: number;
	    name: string;
	    pieces: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new Cartridge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pieces = source["pieces"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class ComboResult {
	    combo: string;
	    grid: string;
	
	    static createFrom(source: any = {}) {
	        return new ComboResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.combo = source["combo"];
	        this.grid = source["grid"];
	    }
	}
	export class Config {
	    id: number;
	    name: string;
	    shapeId: number;
	    cartridgeId: number;
	    hash: string;
	    grid: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.shapeId = source["shapeId"];
	        this.cartridgeId = source["cartridgeId"];
	        this.hash = source["hash"];
	        this.grid = source["grid"];
	        this.sortOrder = source["sortOrder"];
	    }
	}

}

