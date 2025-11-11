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
	export class ExportedBoard {
	    id: string;
	    name: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportedBoard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ExportedCard {
	    id: string;
	    column_id: string;
	    title: string;
	    description?: string;
	    attachments?: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportedCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.column_id = source["column_id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.attachments = source["attachments"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ExportedColumn {
	    id: string;
	    board_id: string;
	    name: string;
	    position: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportedColumn(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.board_id = source["board_id"];
	        this.name = source["name"];
	        this.position = source["position"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class ExportedTranscription {
	    id: string;
	    board_id: string;
	    transcription: string;
	    recording_path?: string;
	    intent?: string;
	    assistant_response?: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportedTranscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.board_id = source["board_id"];
	        this.transcription = source["transcription"];
	        this.recording_path = source["recording_path"];
	        this.intent = source["intent"];
	        this.assistant_response = source["assistant_response"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}

}

