export namespace app {
	
	export class UIEvent {
	    time: number;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new UIEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.kind = source["kind"];
	    }
	}
	export class UIMap {
	    start: number;
	    end: number;
	    durationMs: number;
	    earnings: number;
	
	    static createFrom(source: any = {}) {
	        return new UIMap(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = source["start"];
	        this.end = source["end"];
	        this.durationMs = source["durationMs"];
	        this.earnings = source["earnings"];
	    }
	}
	export class UITallyItem {
	    name: string;
	    type: string;
	    price: number;
	    last_update: number;
	    from: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new UITallyItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.price = source["price"];
	        this.last_update = source["last_update"];
	        this.from = source["from"];
	        this.count = source["count"];
	    }
	}
	export class UIState {
	    inMap: boolean;
	    sessionStart: number;
	    sessionEnd: number;
	    mapStart: number;
	    mapEnd: number;
	    totalDrops: number;
	    tally: Record<string, UITallyItem>;
	    recent: UIEvent[];
	    maps: UIMap[];
	    earningsPerSession: number;
	    earningsPerHour: number;
	    avgMapTimeMs: number;
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inMap = source["inMap"];
	        this.sessionStart = source["sessionStart"];
	        this.sessionEnd = source["sessionEnd"];
	        this.mapStart = source["mapStart"];
	        this.mapEnd = source["mapEnd"];
	        this.totalDrops = source["totalDrops"];
	        this.tally = this.convertValues(source["tally"], UITallyItem, true);
	        this.recent = this.convertValues(source["recent"], UIEvent);
	        this.maps = this.convertValues(source["maps"], UIMap);
	        this.earningsPerSession = source["earningsPerSession"];
	        this.earningsPerHour = source["earningsPerHour"];
	        this.avgMapTimeMs = source["avgMapTimeMs"];
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

