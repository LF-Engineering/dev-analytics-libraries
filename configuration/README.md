# Description
    configuration is meant to be a single source of configuration values.
    in order to decouple configuration logic from configuration storage tool.  
    configuration package can deal with whatever storage tool, and any application
    wants to get or set configurations must deal with configuration module directly
    instead of dealing with a specific storage tool like (ssm or secret manager), also 
    it can be used for local configurations as seen in examples below.

### local usage example
        `   
        // init internal used storage
        s := configuration.NewLocalConfigStorage()
        // init service
	    configProvider := configuration.NewProvider(s)
	    
        // set config value
        URL := "http://localhost:9200"
	    err := configProvider.Set(EsUrl, URL)

        // get config value
        esURL, err := configProvider.Get(EsUrl)`

### aws ssm usage example
        ` 
        // init ssm storage
        s := configuration.NewSSMConfigStorage()
        // init service
	    configProvider := configuration.NewProvider(s)
	    
        // set config value
        URL := "http://localhost:9200"
	    err := configProvider.Set(configuration.EsUrl, URL)

        // get config value
        esURL, err := configProvider.Get(configuration.EsUrl)`
