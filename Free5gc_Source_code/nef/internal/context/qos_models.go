package context

import (
	"github.com/free5gc/openapi/models"
)

type AsSessionMediaComponent struct {
	FlowInfos []models.FlowInfo 
	MedCompN int32
	MedType  models.MediaType
	MirBwUl  string
	MirBwDl  string
	MarBwUl  string
	MarBwDl  string
}

type AsSessionWithQoSSubscription struct {
	// Resource URI
	Self string `json:"self,omitempty" yaml:"self" bson:"self,omitempty"`

	// Supported features (bitmask string)
	SupportedFeatures string `json:"supportedFeatures,omitempty" yaml:"supportedFeatures" bson:"supportedFeatures,omitempty"`

	// Data Network Name
	Dnn string `json:"dnn,omitempty" yaml:"dnn" bson:"dnn,omitempty"`

	// Slice info
	Snssai *models.Snssai `json:"snssai,omitempty" yaml:"snssai" bson:"snssai,omitempty"`

	// Notification callback URL
	NotificationDestination string `json:"notificationDestination" yaml:"notificationDestination" bson:"notificationDestination,omitempty"`

	// UE IPv4 Address
	UeIpv4Addr string `json:"ueIpv4Addr,omitempty" yaml:"ueIpv4Addr" bson:"ueIpv4Addr,omitempty"`

	// GPSI (e.g., msisdn-14155552671)
	Gpsi string `json:"gpsi,omitempty" yaml:"gpsi" bson:"gpsi,omitempty"`

	// Media components map: mediaComponentId -> MediaComponent
	MultiModDatFlows map[string]AsSessionMediaComponent `json:"asSessionMediaComponent,omitempty" yaml:"asSessionMediaComponent" bson:"asSessionMediaComponent,omitempty"`
}

type AsSessionWithQoSSubscriptionPatch struct {
	
	// Notification callback URL
	NotificationDestination string `json:"notificationDestination" yaml:"notificationDestination" bson:"notificationDestination,omitempty"`
	// Media components map: mediaComponentId -> MediaComponent
	MultiModDatFlows map[string]AsSessionMediaComponent `json:"asSessionMediaComponent,omitempty" yaml:"asSessionMediaComponent" bson:"asSessionMediaComponent,omitempty"`
}
	
	
type ServiceName string
	const(
		ServiceName_3GPP_AS_SESSION_WITH_QOS         ServiceName = "3gpp-as-session-with-qos"
	)


type AuthorizationJSON struct {
		Client_id     string `json:"client_id"`
		Client_secret string `json:"client_secret"`
		Grant_type    string `json:"grant_type"`
	}
