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
	export class UIState {
	    inMap: boolean;
	    sessionStart: number;
	    sessionEnd: number;
	    totalDrops: number;
	    tally: Record<string, number>;
	    recent: UIEvent[];
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inMap = source["inMap"];
	        this.sessionStart = source["sessionStart"];
	        this.sessionEnd = source["sessionEnd"];
	        this.totalDrops = source["totalDrops"];
	        this.tally = source["tally"];
	        this.recent = this.convertValues(source["recent"], UIEvent);
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

