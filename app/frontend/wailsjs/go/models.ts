export namespace frontend {
	
	export class FileFilter {
	    DisplayName: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DisplayName = source["DisplayName"];
	        this.Pattern = source["Pattern"];
	    }
	}

}

export namespace query {
	
	export class Board {
	    ID: string;
	    Name: string;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Board(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], sql.NullString);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], sql.NullString);
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
	export class Card {
	    ID: string;
	    ColumnID: string;
	    Title: string;
	    Description: sql.NullString;
	    Attachments: sql.NullString;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Card(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.ColumnID = source["ColumnID"];
	        this.Title = source["Title"];
	        this.Description = this.convertValues(source["Description"], sql.NullString);
	        this.Attachments = this.convertValues(source["Attachments"], sql.NullString);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], sql.NullString);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], sql.NullString);
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
	export class Column {
	    ID: string;
	    BoardID: string;
	    Name: string;
	    Position: number;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.BoardID = source["BoardID"];
	        this.Name = source["Name"];
	        this.Position = source["Position"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], sql.NullString);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], sql.NullString);
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
	export class Setting {
	    ID: number;
	    TranscriptionMethod: string;
	    WhisperBinaryPath: sql.NullString;
	    WhisperModelPath: sql.NullString;
	    OpenaiApiKey: sql.NullString;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Setting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.TranscriptionMethod = source["TranscriptionMethod"];
	        this.WhisperBinaryPath = this.convertValues(source["WhisperBinaryPath"], sql.NullString);
	        this.WhisperModelPath = this.convertValues(source["WhisperModelPath"], sql.NullString);
	        this.OpenaiApiKey = this.convertValues(source["OpenaiApiKey"], sql.NullString);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], sql.NullString);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], sql.NullString);
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
	export class Transcription {
	    ID: string;
	    BoardID: string;
	    Transcription: string;
	    RecordingPath: sql.NullString;
	    Intent: sql.NullString;
	    AssistantResponse: sql.NullString;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Transcription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.BoardID = source["BoardID"];
	        this.Transcription = source["Transcription"];
	        this.RecordingPath = this.convertValues(source["RecordingPath"], sql.NullString);
	        this.Intent = this.convertValues(source["Intent"], sql.NullString);
	        this.AssistantResponse = this.convertValues(source["AssistantResponse"], sql.NullString);
	        this.CreatedAt = this.convertValues(source["CreatedAt"], sql.NullString);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], sql.NullString);
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

export namespace sql {
	
	export class NullString {
	    String: string;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.String = source["String"];
	        this.Valid = source["Valid"];
	    }
	}

}

export namespace types {
	
	export class AppVersion {
	    version: string;
	    notes: string;
	    url: string;
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new AppVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.notes = source["notes"];
	        this.url = source["url"];
	        this.sha256 = source["sha256"];
	    }
	}

}

