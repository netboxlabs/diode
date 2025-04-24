package telemetry

const (
	// AttributeHostname is the hostname of the machine that the request is coming from
	AttributeHostname = "hostname"
	// AttributeSDKName is the name of the SDK that is being used
	AttributeSDKName = "sdk_name"
	// AttributeSDKVersion is the version of the SDK that is being used
	AttributeSDKVersion = "sdk_version"
	// AttributeProducerAppName is the name of the producer application
	AttributeProducerAppName = "producer_app_name"
	// AttributeProducerAppVersion is the version of the producer application
	AttributeProducerAppVersion = "producer_app_version"
	// AttributeSuccess is a boolean attribute that indicates if the request was successful
	AttributeSuccess = "success"
	// AttributeStream is the stream that the request is coming from
	AttributeStream = "stream"
	// AttributeState is the state of the request
	AttributeState = "state"
)
