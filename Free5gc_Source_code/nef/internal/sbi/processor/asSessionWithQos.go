package processor

import (
	"fmt"
	"net/http"
	// "strconv"
	"encoding/json"

	qos_models "github.com/free5gc/nef/internal/context" // Ensure this path is correct and the package exists
	"github.com/free5gc/nef/internal/logger"
	"github.com/free5gc/nef/pkg/factory"
	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
)

func (p *Processor) PostAsSessionWithQoSSubscription(
	scsAsID string,
	qosSub *qos_models.AsSessionWithQoSSubscription,
) *HandlerResponse {
	logger.AssessionwithQosLog.Infof("PostAsSessionWithQoSSubscription - scsAsID[%s]", scsAsID)
	fmt.Printf("PostAsSessionWithQoSSubscription - scsAsID[%s], qosSub: %+v\n", scsAsID, qosSub)
	rsp := validateAsSessionWithQoSData(qosSub)
	if rsp != nil {
		return rsp
	}

	nefCtx := p.Context()
	af := nefCtx.GetAf(scsAsID)
	if af == nil {
		af = nefCtx.NewAf(scsAsID)
		if af == nil {
			pd := openapi.ProblemDetailsSystemFailure("No resource can be allocated")
			return &HandlerResponse{int(pd.Status), nil, pd}
		}
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	correID := nefCtx.NewCorreID()
	afSub := af.NewSubQOS(correID, qosSub)
	if afSub == nil {
		pd := openapi.ProblemDetailsSystemFailure("No resource can be allocated")
		return &HandlerResponse{int(pd.Status), nil, pd}
	}

	if len(qosSub.Gpsi) > 0 || len(qosSub.UeIpv4Addr) > 0 {
		asc := p.convertAsSessionWithQoSSubToAppSessionContext(qosSub)
		fmt.Printf("convertAsSessionWithQoSSubToAppSessionContext: %+v\n", asc)
		rspStatus, rspBody, appSessID := p.Consumer().PostAppSessions(asc)
		if rspStatus != http.StatusCreated {
			return &HandlerResponse{rspStatus, nil, rspBody}
		}
		fmt.Printf("PostAppSessions response: appSessID=%s\n", appSessID)
		afSub.AppSessID = appSessID
	} else {
		// Invalid case. Return Error
		pd := openapi.ProblemDetailsMalformedReqSyntax("Not individual UE case")
		return &HandlerResponse{int(pd.Status), nil, pd}
	}

	af.Subs[afSub.SubID] = afSub
	af.Log.Infoln("Subscription is added")

	nefCtx.AddAf(af)

	// Create Location URI
	qosSub.Self = p.genAsSessionWithQoSSubURI(scsAsID, afSub.SubID)
	headers := map[string][]string{
		"Location": {qosSub.Self},
	}
	return &HandlerResponse{http.StatusCreated, headers, qosSub}
}


func (p *Processor) PatchIndividualAsSessionWithQoSSubscription(
	scsAsID, subscriptionID string,
	qosSubPatch *qos_models.AsSessionWithQoSSubscriptionPatch,
) *HandlerResponse {
	logger.AssessionwithQosLog.Infof("PatchIndividualAsSessionWithQoSSubscription - scsAsID[%s], subscriptionID[%s]", scsAsID, subscriptionID)

	af := p.Context().GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF is not found")
		return &HandlerResponse{http.StatusNotFound, nil, pd}
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	afSub, ok := af.Subs[subscriptionID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Subscription is not found")
		return &HandlerResponse{http.StatusNotFound, nil, pd}
	}

	if afSub.AppSessID != "" {
		ascUpdateData := convertAsSessionwithQosSubPatchToAppSessionContextUpdateData(qosSubPatch)
		rspStatus, rspBody := p.Consumer().PatchAppSession(afSub.AppSessID, ascUpdateData)
		if rspStatus != http.StatusOK &&
			rspStatus != http.StatusNoContent {
			return &HandlerResponse{rspStatus, nil, rspBody}
		}
	
	} else {
		pd := openapi.ProblemDetailsDataNotFound("No AppSessID ")
		return &HandlerResponse{int(pd.Status), nil, pd}
	}

	afSub.PatchQosSubData(qosSubPatch)
	return &HandlerResponse{http.StatusOK, nil, afSub.QosSub}
}

func (p *Processor) DeleteIndividualAsSessionWithQoSSubscription(
	scsAsID, subscriptionID string,
) *HandlerResponse {
	logger.AssessionwithQosLog.Infof("DeleteIndividualAsSessionWithQoSSubscription - scsAsID[%s], subscriptionID[%s]", scsAsID, subscriptionID)

	af := p.Context().GetAf(scsAsID)
	if af == nil {
		pd := openapi.ProblemDetailsDataNotFound("AF is not found")
		return &HandlerResponse{http.StatusNotFound, nil, pd}
	}

	af.Mu.Lock()
	defer af.Mu.Unlock()

	sub, ok := af.Subs[subscriptionID]
	if !ok {
		pd := openapi.ProblemDetailsDataNotFound("Subscription is not found")
		return &HandlerResponse{http.StatusNotFound, nil, pd}
	}

	if sub.AppSessID != "" {
		rspStatus, rspBody := p.Consumer().DeleteAppSession(sub.AppSessID)
		if rspStatus != http.StatusOK &&
			rspStatus != http.StatusNoContent {
			return &HandlerResponse{rspStatus, nil, rspBody}
		}
	} 
	
	delete(af.Subs, subscriptionID)
	return &HandlerResponse{http.StatusNoContent, nil, nil}
}

func validateAsSessionWithQoSData(
	qosSub *qos_models.AsSessionWithQoSSubscription,
) *HandlerResponse {

	// TS29.522: One of individual UE identifier
	// (i.e. "gpsi", “macAddr”, "ipv4Addr" or "ipv6Addr"),
	if qosSub.Gpsi == "" &&
		qosSub.UeIpv4Addr == "" {
		pd := openapi.
			ProblemDetailsMalformedReqSyntax(
				"Missing one of Gpsi, Ipv4Addr,ExterAppId")
		return &HandlerResponse{int(pd.Status), nil, pd}
	}
	return nil
}

func (p *Processor) genAsSessionWithQoSSubURI(
	scsAsID string, subscriptionId string,
) string {
	// E.g. https://localhost:29505/3gpp-traffic-Influence/v1/{afId}/subscriptions/{subscriptionId}
	return p.Config().ServiceUri(factory.ServiceAsSessionWithQoS) + "/" + scsAsID + "/subscriptions/" + subscriptionId
}

//	func (p *Processor) genNotificationUri() string {
//		return p.Config().ServiceUri(factory.ServiceNefCallback) + "/notification/smf"
//	}
func (p *Processor) convertAsSessionWithQoSSubToAppSessionContext(
	qosSub *qos_models.AsSessionWithQoSSubscription,
) *models.AppSessionContext {

	medComponents := make(map[string]models.MediaComponent)

	for medCompId, comp := range qosSub.MultiModDatFlows {
		// Sub-component map per media component
		medSubComps := make(map[string]models.MediaSubComponent)

		// FNum starts at 1 and increments for each flowInfo
		for idx, flowInfo := range comp.FlowInfos {
			subCompId := fmt.Sprintf("%d", idx+1)

			medSubComps[subCompId] = models.MediaSubComponent{
				FNum:      int32(idx + 1),
				FDescs:    flowInfo.FlowDescriptions,
				FStatus:   "ENABLED",
				FlowUsage: "AF_SIGNALLING",
			}
		}

		medComponents[medCompId] = models.MediaComponent{
			MedCompN:    comp.MedCompN,
			MedType:     comp.MedType,
			MirBwDl:     comp.MirBwDl,
			MirBwUl:     comp.MirBwUl,
			MarBwDl:     comp.MarBwDl,
			MarBwUl:     comp.MarBwUl,
			MedSubComps: medSubComps,
		}
	}

	asc := &models.AppSessionContext{
		AscReqData: &models.AppSessionContextReqData{
			Dnn:           qosSub.Dnn,
			SuppFeat:      qosSub.SupportedFeatures,
			UeIpv4:        qosSub.UeIpv4Addr,
			NotifUri:      qosSub.NotificationDestination,
			Gpsi:          qosSub.Gpsi,
			MedComponents: medComponents,
		},
	}

	// Optional: pretty print for debugging
	if prettyJson, err := json.MarshalIndent(asc, "", "  "); err == nil {
		logger.CtxLog.Infof("Generated AppSessionContext:\n%s", string(prettyJson))
	} else {
		logger.CtxLog.Errorln("Failed to marshal AppSessionContext:", err)
	}

	return asc
}

func convertAsSessionwithQosSubPatchToAppSessionContextUpdateData(
	qosSubPatch *qos_models.AsSessionWithQoSSubscriptionPatch,
) *models.AppSessionContextUpdateData {
	ascUpdate := &models.AppSessionContextUpdateData{}

	// Only add MedComponents if present
	if len(qosSubPatch.MultiModDatFlows) > 0 {
		medComponents := make(map[string]models.MediaComponentRm)

		for medCompId, comp := range qosSubPatch.MultiModDatFlows {
			subComps := make(map[string]models.MediaSubComponentRm)

			for idx, flowInfo := range comp.FlowInfos {
				subComps[fmt.Sprintf("%d", idx+1)] = models.MediaSubComponentRm{
					FNum:      int32(idx + 1),
					FDescs:    flowInfo.FlowDescriptions,
					FStatus:   "ENABLED",
					FlowUsage: "AF_SIGNALLING",
				}
			}

			medComponents[medCompId] = models.MediaComponentRm{
				MedCompN:    comp.MedCompN,
				MedType:     comp.MedType,
				MirBwDl:     comp.MirBwDl,
				MarBwDl:     comp.MarBwDl,
				MirBwUl:     comp.MirBwUl,
				MarBwUl:     comp.MarBwUl,
				AfRoutReq: &models.AfRoutingRequirementRm{
					SpVal: &models.SpatialValidityRm{},
				},
				MedSubComps: subComps,
			}
		}

		ascUpdate.MedComponents = medComponents
	}

	// Optional: pretty print for debugging
	if prettyJson, err := json.MarshalIndent(ascUpdate, "", "  "); err == nil {
		logger.CtxLog.Infof("Generated AppSessionContext:\n%s", string(prettyJson))
	} else {
		logger.CtxLog.Errorln("Failed to marshal AppSessionContext:", err)
	}

	return ascUpdate
}






// func (p *Processor) convertAsSessionWithQoSSubToAppSessionContext(

// 	qosSub *qos_models.AsSessionWithQoSSubscription,
// ) *models.AppSessionContext {

// 	fmt.Printf("convertAsSessionWithQoSSubToAppSessionContext qosSub: %+v\n", qosSub)
// 	asc := &models.AppSessionContext{
// 		AscReqData: &models.AppSessionContextReqData{
// 			MedComponents: map[string]models.MediaComponent{
// 				"1": {
// 					MarBwDl:  qosSub.AsSessionMediaComponent.MarBwDl,
// 					MarBwUl:  qosSub.AsSessionMediaComponent.MarBwUl,
// 					MirBwDl:  qosSub.AsSessionMediaComponent.MirBwDl,
// 					MirBwUl:  qosSub.AsSessionMediaComponent.MirBwUl,
// 					MedCompN: qosSub.AsSessionMediaComponent.MedCompN,
// 					MedType:  qosSub.AsSessionMediaComponent.MedType,
// 					MedSubComps: map[string]models.MediaSubComponent{
// 						"1": {
// 							FNum:      1,
// 							FDescs:    qosSub.FlowInfo.FlowDescriptions,
// 							FlowUsage: "AF_SIGNALLING",
// 						},
// 					},
// 				},
// 			},
// 			UeIpv4:   qosSub.UeIpv4Addr,
// 			NotifUri: qosSub.NotificationDestination,
// 			SuppFeat: qosSub.SupportedFeatures,
// 			Dnn:      qosSub.Dnn,
// 			Gpsi:     qosSub.Gpsi,
// 			// SliceInfo: qosSub.Snssai,
// 			// Supi: qosSub.Supi,
// 		},
// 	}
// prettyJson, err := json.MarshalIndent(asc, "", "  ")
// if err != nil {
// 	logger.CtxLog.Errorln("Failed to marshal appSessionreqData:", err)
// } else {
// 	logger.CtxLog.Infof("Deserialized AppSessionContextUpdateData:\n%s", string(prettyJson))
// }
// 	return asc
// }

// func (p *Processor) convertTrafficInfluSubPatchToAppSessionContextUpdateData(
// 	tiSubPatch *models_nef.TrafficInfluSubPatch,
// ) *models.AppSessionContextUpdateData {
// 	ascUpdate := &models.AppSessionContextUpdateData{
// 		AfRoutReq: &models.AfRoutingRequirementRm{
// 			AppReloc:    tiSubPatch.AppReloInd,
// 			RouteToLocs: tiSubPatch.TrafficRoutes,
// 			TempVals:    tiSubPatch.TempValidities,
// 		},
// 	}
// 	return ascUpdate
// }

// func (p *Processor) PutIndividualAsSessionWithQoSSubscription(
// 	scsAsID, subscriptionId string,
// 	QosSub *qos_models.AsSessionWithQoSSubscription,
// ) *HandlerResponse {
// 	logger.TrafInfluLog.Infof("PutIndividualAsSessionWithQoSSubscription - scsAsID[%s], subscriptionId[%s]", scsAsID, subscriptionId)

// 	rsp := validateAsSessionWithQoSData(QosSubSub)
// 	if rsp != nil {
// 		return rsp
// 	}

// 	af := p.Context().GetAf(scsAsID)
// 	if af == nil {
// 		pd := openapi.ProblemDetailsDataNotFound("AF is not found")
// 		return &HandlerResponse{http.StatusNotFound, nil, pd}
// 	}

// 	af.Mu.Lock()
// 	defer af.Mu.Unlock()

// 	afSub, ok := af.Subs[subID]
// 	if !ok {
// 		pd := openapi.ProblemDetailsDataNotFound("Subscription is not found")
// 		return &HandlerResponse{http.StatusNotFound, nil, pd}
// 	}

// 	afSub.TiSub = tiSub
// 	if afSub.AppSessID != "" {
// 		asc := p.convertTrafficInfluSubToAppSessionContext(tiSub, afSub.NotifCorreID)
// 		rspStatus, rspBody, appSessID := p.Consumer().PostAppSessions(asc)
// 		if rspStatus != http.StatusCreated {
// 			return &HandlerResponse{rspStatus, nil, rspBody}
// 		}
// 		afSub.AppSessID = appSessID
// 	} else if afSub.InfluID != "" {
// 		tiData := p.convertTrafficInfluSubToTrafficInfluData(tiSub, afSub.NotifCorreID)
// 		rspStatus, rspBody := p.Consumer().AppDataInfluenceDataPut(afSub.InfluID, tiData)
// 		if rspStatus != http.StatusOK &&
// 			rspStatus != http.StatusCreated &&
// 			rspStatus != http.StatusNoContent {
// 			return &HandlerResponse{rspStatus, nil, rspBody}
// 		}
// 	} else {
// 		pd := openapi.ProblemDetailsDataNotFound("No AppSessID or InfluID")
// 		return &HandlerResponse{int(pd.Status), nil, pd}
// 	}

// 	return &HandlerResponse{http.StatusOK, nil, afSub.TiSub}
// }
