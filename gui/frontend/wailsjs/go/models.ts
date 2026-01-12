export namespace main {
	
	export class FileSelection {
	    csvPath: string;
	    excelPath: string;
	
	    static createFrom(source: any = {}) {
	        return new FileSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.csvPath = source["csvPath"];
	        this.excelPath = source["excelPath"];
	    }
	}
	export class ValidationResult {
	    valid: boolean;
	    fileName: string;
	    filePath: string;
	    message: string;
	    headers: string[];
	    rowCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.fileName = source["fileName"];
	        this.filePath = source["filePath"];
	        this.message = source["message"];
	        this.headers = source["headers"];
	        this.rowCount = source["rowCount"];
	    }
	}

}

