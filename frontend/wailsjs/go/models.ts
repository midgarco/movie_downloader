export namespace moviedownloader {
	
	export class Movie {
	    id?: string;
	    filename?: string;
	    codec?: string;
	    runtime?: string;
	    extension?: string;
	    resolution?: string;
	    size?: string;
	    post_date?: string;
	    subject?: string;
	    group?: string;
	    audio_languages?: string[];
	    full_resolution?: string;
	    height?: string;
	    width?: string;
	    bps?: number;
	    sample_rate?: number;
	    fps?: number;
	    audio_codec?: string;
	    poster?: string;
	    primary_url?: string;
	    fallback_url?: string;
	    virus?: boolean;
	    type?: string;
	    ts?: number;
	    sub_languages?: string[];
	    raw_size?: number;
	
	    static createFrom(source: any = {}) {
	        return new Movie(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filename = source["filename"];
	        this.codec = source["codec"];
	        this.runtime = source["runtime"];
	        this.extension = source["extension"];
	        this.resolution = source["resolution"];
	        this.size = source["size"];
	        this.post_date = source["post_date"];
	        this.subject = source["subject"];
	        this.group = source["group"];
	        this.audio_languages = source["audio_languages"];
	        this.full_resolution = source["full_resolution"];
	        this.height = source["height"];
	        this.width = source["width"];
	        this.bps = source["bps"];
	        this.sample_rate = source["sample_rate"];
	        this.fps = source["fps"];
	        this.audio_codec = source["audio_codec"];
	        this.poster = source["poster"];
	        this.primary_url = source["primary_url"];
	        this.fallback_url = source["fallback_url"];
	        this.virus = source["virus"];
	        this.type = source["type"];
	        this.ts = source["ts"];
	        this.sub_languages = source["sub_languages"];
	        this.raw_size = source["raw_size"];
	    }
	}
	export class SearchResults {
	    movies?: Movie[];
	    num_pages?: number;
	    page?: number;
	    per_page?: string;
	    count?: number;
	    returned?: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.movies = this.convertValues(source["movies"], Movie);
	        this.num_pages = source["num_pages"];
	        this.page = source["page"];
	        this.per_page = source["per_page"];
	        this.count = source["count"];
	        this.returned = source["returned"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class SearchResponse {
	    results?: SearchResults;
	
	    static createFrom(source: any = {}) {
	        return new SearchResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], SearchResults);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

