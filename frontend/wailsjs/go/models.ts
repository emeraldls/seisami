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
	export class Column {
	    ID: string;
	    BoardID: string;
	    Title: string;
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
	        this.Title = source["Title"];
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
	export class Ticket {
	    ID: string;
	    ColumnID: string;
	    Title: string;
	    Description: sql.NullString;
	    AssigneeID: sql.NullInt64;
	    StoryPoints: sql.NullInt64;
	    PrLink: sql.NullString;
	    TicketType: string;
	    CreatedAt: sql.NullString;
	    UpdatedAt: sql.NullString;
	
	    static createFrom(source: any = {}) {
	        return new Ticket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.ColumnID = source["ColumnID"];
	        this.Title = source["Title"];
	        this.Description = this.convertValues(source["Description"], sql.NullString);
	        this.AssigneeID = this.convertValues(source["AssigneeID"], sql.NullInt64);
	        this.StoryPoints = this.convertValues(source["StoryPoints"], sql.NullInt64);
	        this.PrLink = this.convertValues(source["PrLink"], sql.NullString);
	        this.TicketType = source["TicketType"];
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
	
	export class NullInt64 {
	    Int64: number;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NullInt64(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Int64 = source["Int64"];
	        this.Valid = source["Valid"];
	    }
	}
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

