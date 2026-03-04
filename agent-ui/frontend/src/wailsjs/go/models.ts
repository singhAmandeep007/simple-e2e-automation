export namespace main {
	
	export class AgentConfig {
	    agentId: string;
	    controlPlaneWs: string;
	    agentBinPath: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agentId = source["agentId"];
	        this.controlPlaneWs = source["controlPlaneWs"];
	        this.agentBinPath = source["agentBinPath"];
	    }
	}

}

