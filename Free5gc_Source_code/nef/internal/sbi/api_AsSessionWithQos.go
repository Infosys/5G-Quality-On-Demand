package sbi

import (
	
	"net/http"
	"encoding/json"
	"github.com/free5gc/nef/internal/logger"

	qos_models "github.com/free5gc/nef/internal/context"
	"github.com/gin-gonic/gin"
)

func (s *Server) getAsSessionWithQoSEndpoints() []Endpoint {
	return []Endpoint{

		{
			Method:  http.MethodPost,
			Pattern: "/:scsAsId/subscriptions",
			APIFunc: s.apiPostAsSessionWithQoSSubscription,
		},
		{
			Method:  http.MethodPatch,
			Pattern: "/:scsAsId/subscriptions/:subscriptionId",
			APIFunc: s.apiPatchIndividualAsSessionWithQoSSubscription,
		},
		{
			Method:  http.MethodDelete,
			Pattern: "/:scsAsId/subscriptions/:subscriptionId",
			APIFunc: s.apiDeleteIndividualAsSessionWithQoSSubscription,
		},
		

	}
}

func (s *Server) apiPostAsSessionWithQoSSubscription(gc *gin.Context) {
	// Check content type
	contentType, err := checkContentTypeIsJSON(gc)
	if err != nil {
		return
	}
	// Deserialize the data
	var qosSub *qos_models.AsSessionWithQoSSubscription
	if err := s.deserializeData(gc, &qosSub, contentType); err != nil {
		return
	}
	prettyJson, err := json.MarshalIndent(qosSub, "", "  ")
	if err != nil {
		logger.CtxLog.Errorln("Failed to marshal qosSub:", err)
	} else {
		logger.CtxLog.Infof("Deserialized AppSessionContextUpdateData:\n%s", string(prettyJson))
	}
	// Post traffic influence subscription
	hdlRsp := s.Processor().PostAsSessionWithQoSSubscription(
		gc.Param("scsAsId"), qosSub)
	// Build and send HTTP response
	s.buildAndSendHttpResponse(gc, hdlRsp, false)
}

func (s *Server) apiPatchIndividualAsSessionWithQoSSubscription(gc *gin.Context) {
	// Check content type
	contentType, err := checkContentTypeIsJSON(gc)
	if err != nil {
		return
	}
	// Deserialize the data
	var qosSubpatch *qos_models.AsSessionWithQoSSubscriptionPatch
	if err := s.deserializeData(gc, &qosSubpatch, contentType); err != nil {
		return
	}
	prettyJson, err := json.MarshalIndent(qosSubpatch, "", "  ")
	if err != nil {
		logger.CtxLog.Errorln("Failed to marshal qosSub:", err)
	} else {
		logger.CtxLog.Infof("Deserialized AppSessionContextUpdateData:\n%s", string(prettyJson))
	}
	// Patch traffic influence subscription
	hdlRsp := s.Processor().PatchIndividualAsSessionWithQoSSubscription(
		gc.Param("scsAsId"), gc.Param("subscriptionId"), qosSubpatch)
	// Build and send HTTP response
	s.buildAndSendHttpResponse(gc, hdlRsp, false)
}

func (s *Server) apiDeleteIndividualAsSessionWithQoSSubscription(gc *gin.Context) {
	hdlRsp := s.Processor().DeleteIndividualAsSessionWithQoSSubscription(
		gc.Param("scsAsId"), gc.Param("subscriptionId"))

	s.buildAndSendHttpResponse(gc, hdlRsp, false)
}
