export namespace engine {
	
	export class Options {
	    Input: string;
	    Output: string;
	    Files: string;
	    OutputDir: string;
	    StartTime: string;
	    EndTime: string;
	    ExperiredText: string;
	    UnsupportedText: string;
	    PwdEnabled: boolean;
	    UserPassword: string;
	    OwnerPassword: string;
	    WatermarkEnabled: boolean;
	    WatermarkText: string;
	    WatermarkDesc: string;
	    AllowedPrint: boolean;
	    AllowedCopy: boolean;
	    AllowedEdit: boolean;
	    AllowedConvert: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Input = source["Input"];
	        this.Output = source["Output"];
	        this.Files = source["Files"];
	        this.OutputDir = source["OutputDir"];
	        this.StartTime = source["StartTime"];
	        this.EndTime = source["EndTime"];
	        this.ExperiredText = source["ExperiredText"];
	        this.UnsupportedText = source["UnsupportedText"];
	        this.PwdEnabled = source["PwdEnabled"];
	        this.UserPassword = source["UserPassword"];
	        this.OwnerPassword = source["OwnerPassword"];
	        this.WatermarkEnabled = source["WatermarkEnabled"];
	        this.WatermarkText = source["WatermarkText"];
	        this.WatermarkDesc = source["WatermarkDesc"];
	        this.AllowedPrint = source["AllowedPrint"];
	        this.AllowedCopy = source["AllowedCopy"];
	        this.AllowedEdit = source["AllowedEdit"];
	        this.AllowedConvert = source["AllowedConvert"];
	    }
	}

}

