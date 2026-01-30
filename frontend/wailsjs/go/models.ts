export namespace engine {
	
	export class Options {
	    Input: string;
	    Output: string;
	    Files: string;
	    OutputDir: string;
	    // Go type: time
	    StartTime: any;
	    // Go type: time
	    EndTime: any;
	    Watermark: string;
	    ExperiredText: string;
	    UnsupportedText: string;
	    DisablePrint: boolean;
	    DisableCopy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Input = source["Input"];
	        this.Output = source["Output"];
	        this.Files = source["Files"];
	        this.OutputDir = source["OutputDir"];
	        this.StartTime = this.convertValues(source["StartTime"], null);
	        this.EndTime = this.convertValues(source["EndTime"], null);
	        this.Watermark = source["Watermark"];
	        this.ExperiredText = source["ExperiredText"];
	        this.UnsupportedText = source["UnsupportedText"];
	        this.DisablePrint = source["DisablePrint"];
	        this.DisableCopy = source["DisableCopy"];
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

