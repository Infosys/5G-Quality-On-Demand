package processor

import (
	"fmt"

	"encoding/json"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/pfcp"
	"github.com/free5gc/pfcp/pfcpType"
	"github.com/free5gc/pfcp/pfcpUdp"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/logger"
	pfcp_message "github.com/free5gc/smf/internal/pfcp/message"
)

func logAsJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		logger.PduSessLog.Errorf("Failed to marshal %s: %v", label, err)
		return
	}
	logger.PduSessLog.Infof("%s: %s", label, string(data))
}

type PFCPState struct {
	upf     *smf_context.UPF
	pdrList []*smf_context.PDR
	farList []*smf_context.FAR
	barList []*smf_context.BAR
	qerList []*smf_context.QER
	urrList []*smf_context.URR
}

type SendPfcpResult struct {
	Status smf_context.PFCPSessionResponseStatus
	RcvMsg *pfcpUdp.Message
	Err    error
}

// ActivateUPFSession send all datapaths to UPFs and send result to UE
// It returns after all PFCP response have been returned or timed out,
// and before sending N1N2MessageTransfer request if it is needed.

func (p *Processor) ActivateUPFSession(
	smContext *smf_context.SMContext,
	notifyUeHander func(*smf_context.SMContext, bool),
) {
	smContext.Log.Infof("In ActivateUPFSession")
	modifySessionInvoked := false // Flag to track if modifyExistingPfcpSession is invoked
	if smContext == nil {
		fmt.Println("Error: smContext is nil")
	} else if smContext.Tunnel == nil {
		fmt.Println("Error: smContext.Tunnel is nil")
	} else if len(smContext.Tunnel.DataPathPool) == 0 {
		fmt.Println("Error: smContext.Tunnel.DataPathPool is empty")
	} else {
		for i, dataPath := range smContext.Tunnel.DataPathPool {
			if dataPath == nil {
				fmt.Printf("Error: DataPath at index %d is nil\n", i)
				continue
			}
			if dataPath.FirstDPNode == nil {
				fmt.Printf("Error: FirstDPNode in DataPath at index %d is nil\n", i)
				continue
			}

			// // --- Printing DownLinkTunnel PDR and QER ---
			// if dataPath.FirstDPNode.DownLinkTunnel == nil {
			// 	fmt.Printf("Error: DownLinkTunnel in FirstDPNode at index %d is nil\n", i)
			// } else if dataPath.FirstDPNode.DownLinkTunnel.PDR == nil {
			// 	fmt.Printf("Error: DownLinkTunnel.PDR in FirstDPNode at index %d is nil\n", i)
			// } else {
			// 	pdr := dataPath.FirstDPNode.DownLinkTunnel.PDR
			// 	if jsonData, err := json.MarshalIndent(pdr, "", "  "); err != nil {
			// 		fmt.Printf("Failed to marshal DownLink PDR: %v\n", err)
			// 	} else {
			// 		fmt.Printf("DownLink PDR: %s\n", string(jsonData))
			// 	}
			// 	// Print QERs from DownLinkTunnel PDR
			// 	if pdr.QER == nil || len(pdr.QER) == 0 {
			// 		fmt.Printf("No QERs found in DownLinkTunnel.PDR at index %d\n", i)
			// 	} else {
			// 		for j, qer := range pdr.QER {
			// 			if qer == nil {
			// 				fmt.Printf("QER at index %d in DownLinkTunnel.PDR is nil\n", j)
			// 				continue
			// 			}
			// 			if jsonData, err := json.MarshalIndent(qer, "", "  "); err != nil {
			// 				fmt.Printf("Failed to marshal QER at index %d in DownLinkTunnel.PDR: %v\n", j, err)
			// 			} else {
			// 				fmt.Printf("DownLink QER[%d]: %s\n", j, string(jsonData))
			// 			}
			// 		}
			// 	}
			// }

			// --- Printing UpLinkTunnel PDR and QER ---
			if dataPath.FirstDPNode.UpLinkTunnel == nil {
				fmt.Printf("Error: UpLinkTunnel in FirstDPNode at index %d is nil\n", i)
			} else if dataPath.FirstDPNode.UpLinkTunnel.PDR == nil {
				fmt.Printf("Error: UpLinkTunnel.PDR in FirstDPNode at index %d is nil\n", i)
			} else {
				pdr := dataPath.FirstDPNode.UpLinkTunnel.PDR
				if jsonData, err := json.MarshalIndent(pdr, "", "  "); err != nil {
					fmt.Printf("Failed to marshal UpLink PDR: %v\n", err)
				} else {
					fmt.Printf("UpLink PDR: %s\n", string(jsonData))
				}
				// Print QERs from UpLinkTunnel PDR
				if pdr.QER == nil || len(pdr.QER) == 0 {
					fmt.Printf("No QERs found in UpLinkTunnel.PDR at index %d\n", i)
				} else {
					for j, qer := range pdr.QER {
						if qer == nil {
							fmt.Printf("QER at index %d in UpLinkTunnel.PDR is nil\n", j)
							continue
						}
						if jsonData, err := json.MarshalIndent(qer, "", "  "); err != nil {
							fmt.Printf("Failed to marshal QER at index %d in UpLinkTunnel.PDR: %v\n", j, err)
						} else {
							fmt.Printf("UpLink QER[%d]: %s\n", j, string(jsonData))
						}
					}
				}
			}
		}
	}

	// if smContext == nil {
	// 	fmt.Println("Error: smContext is nil")
	// } else if smContext.Tunnel == nil {
	// 	fmt.Println("Error: smContext.Tunnel is nil")
	// } else if len(smContext.Tunnel.DataPathPool) == 0 {
	// 	fmt.Println("Error: smContext.Tunnel.DataPathPool is empty")
	// } else {
	// 	for i, dataPath := range smContext.Tunnel.DataPathPool {
	// 		if dataPath == nil {
	// 			fmt.Printf("Error: DataPath at index %d is nil\n", i)
	// 			continue
	// 		}
	// 		if dataPath.FirstDPNode == nil {
	// 			fmt.Printf("Error: FirstDPNode in DataPath at index %d is nil\n", i)
	// 			continue
	// 		}

	// 		// --- Printing DownLinkTunnel PDR ---
	// 		if dataPath.FirstDPNode.DownLinkTunnel == nil {
	// 			fmt.Printf("Error: DownLinkTunnel in FirstDPNode at index %d is nil\n", i)
	// 		} else if dataPath.FirstDPNode.DownLinkTunnel.PDR == nil {
	// 			fmt.Printf("Error: DownLinkTunnel.PDR in FirstDPNode at index %d is nil\n", i)
	// 		} else {
	// 			pdr := dataPath.FirstDPNode.DownLinkTunnel.PDR
	// 			if jsonData, err := json.MarshalIndent(pdr, "", "  "); err != nil {
	// 				fmt.Printf("Failed to marshal DownLink PDR: %v\n", err)
	// 			} else {
	// 				fmt.Printf("DownLink PDR: %s\n", string(jsonData))
	// 			}
	// 		}

	// 		// --- Printing UpLinkTunnel PDR ---
	// 		if dataPath.FirstDPNode.UpLinkTunnel == nil {
	// 			fmt.Printf("Error: UpLinkTunnel in FirstDPNode at index %d is nil\n", i)
	// 		} else if dataPath.FirstDPNode.UpLinkTunnel.PDR == nil {
	// 			fmt.Printf("Error: UpLinkTunnel.PDR in FirstDPNode at index %d is nil\n", i)
	// 		} else {
	// 			pdr := dataPath.FirstDPNode.UpLinkTunnel.PDR
	// 			if jsonData, err := json.MarshalIndent(pdr, "", "  "); err != nil {
	// 				fmt.Printf("Failed to marshal UpLink PDR: %v\n", err)
	// 			} else {
	// 				fmt.Printf("UpLink PDR: %s\n", string(jsonData))
	// 			}
	// 		}
	// 	}
	// }

	// Create a copy to prevent modifying the original smContext
	// smContextCopy := smContext
	// // // Nullify the fields that cause the circular reference.
	// if smContextCopy.Tunnel != nil {
	// 	for _, dp := range smContextCopy.Tunnel.DataPathPool {
	// 		if dp != nil && dp.FirstDPNode != nil {
	// 			nodeCopy := *dp.FirstDPNode
	// 			nodeCopy.UpLinkTunnel = nil
	// 			nodeCopy.DownLinkTunnel = nil
	// 			dp.FirstDPNode = &nodeCopy
	// 		}
	// 	}
	// }
	// logAsJSON("This is the smContext", smContextCopy)
	//logAsJSON("This is the smContext", smContext)
	// // Marshal and print the smContext object for debugging
	// if jsonData, err := json.MarshalIndent(smContext, "", "  "); err != nil {
	// 	logger.CtxLog.Errorf("Failed to marshal smContext: %v", err)
	// } else {
	// 	logger.CtxLog.Infof("5g team This is the marshalled smContext: %s", string(jsonData))
	// }

	pfcpPool := make(map[string]*PFCPState)
	counter := 0 // Initialize a counter

	for _, dataPath := range smContext.Tunnel.DataPathPool {
		if !dataPath.Activated {
			continue
		}
		for node := dataPath.FirstDPNode; node != nil; node = node.Next() {
			counter++ // Increment the counter for each processed node
			smContext.Log.Infof("Processing node #%d: %+v", counter, node)

			pdrList := make([]*smf_context.PDR, 0, 2)
			farList := make([]*smf_context.FAR, 0, 2)
			qerList := make([]*smf_context.QER, 0, 2)
			logAsJSON("QER list created ", qerList)

			urrList := make([]*smf_context.URR, 0, 2)

			if node.UpLinkTunnel != nil && node.UpLinkTunnel.PDR != nil {
				pdrList = append(pdrList, node.UpLinkTunnel.PDR)
				farList = append(farList, node.UpLinkTunnel.PDR.FAR)
				if node.UpLinkTunnel.PDR.QER != nil {
					logAsJSON("retrieved QER from UpLink Tunnel", node.UpLinkTunnel.PDR.QER)
					qerList = append(qerList, node.UpLinkTunnel.PDR.QER...)
					logAsJSON("Added QER from UpLink Tunnel", qerList)

				}
				if node.UpLinkTunnel.PDR.URR != nil {
					urrList = append(urrList, node.UpLinkTunnel.PDR.URR...)
				}
			}
			if node.DownLinkTunnel != nil && node.DownLinkTunnel.PDR != nil {
				pdrList = append(pdrList, node.DownLinkTunnel.PDR)
				farList = append(farList, node.DownLinkTunnel.PDR.FAR)
				if node.DownLinkTunnel.PDR.URR != nil {
					urrList = append(urrList, node.DownLinkTunnel.PDR.URR...)
				}
				// skip send QER because uplink and downlink shared one QER
			}

			pfcpState := pfcpPool[node.GetNodeIP()]
			smContext.Log.Infof("Existing PFCPState found for node: %+v", node.GetNodeIP())

			if pfcpState == nil {
				pfcpPool[node.GetNodeIP()] = &PFCPState{
					upf:     node.UPF,
					pdrList: pdrList,
					farList: farList,
					qerList: qerList,
					urrList: urrList,
				}
			} else {
				pfcpState.pdrList = append(pfcpState.pdrList, pdrList...)
				pfcpState.farList = append(pfcpState.farList, farList...)
				pfcpState.qerList = append(pfcpState.qerList, qerList...)
				pfcpState.urrList = append(pfcpState.urrList, urrList...)
			}
		}
	}

	resChan := make(chan SendPfcpResult)

	for ip, pfcp := range pfcpPool {
		sessionContext, exist := smContext.PFCPContext[ip]
		if !exist || sessionContext.RemoteSEID == 0 {
			go establishPfcpSession(smContext, pfcp, resChan)
		} else {
			modifySessionInvoked = true // Set flag when modifyExistingPfcpSession is invoked
			go modifyExistingPfcpSession(smContext, pfcp, resChan, "")
		}
	}

	waitAllPfcpRsp(smContext, len(pfcpPool), resChan, notifyUeHander)
	close(resChan)
	// Execute the code only if modifyExistingPfcpSession was invoked
	if modifySessionInvoked {
		// Build and send NAS and N2 messages using N1N2MessageTransfer
		nasBuf, err := smf_context.BuildGSMPDUSessionModificationCommand(smContext)
		if err != nil {
			smContext.Log.Errorf("Failed to build GSM PDUSessionModificationCommand: %+v", err)
			return
		} else {
			smContext.Log.Infof("Successfully built GSM PDUSessionModificationCommand")
		}

		n2Buf, err := smf_context.BuildPDUSessionResourceModifyRequestTransfer(smContext)
		if err != nil {
			smContext.Log.Errorf("Failed to build PDUSessionResourceModifyRequestTransfer: %+v", err)
			return
		} else {
			smContext.Log.Infof("Successfully built PDUSessionResourceModifyRequestTransfer")
		}

		n1n2Request := models.N1N2MessageTransferRequest{
			BinaryDataN1Message:     nasBuf,
			BinaryDataN2Information: n2Buf,
			JsonData: &models.N1N2MessageTransferReqData{
				PduSessionId: smContext.PDUSessionID,
				N1MessageContainer: &models.N1MessageContainer{
					N1MessageClass:   "SM",
					N1MessageContent: &models.RefToBinaryData{ContentId: "GSM_NAS"},
				},
				N2InfoContainer: &models.N2InfoContainer{
					N2InformationClass: models.N2InformationClass_SM,
					SmInfo: &models.N2SmInformation{
						PduSessionId: smContext.PDUSessionID,
						N2InfoContent: &models.N2InfoContent{
							NgapIeType: models.AmfCommunicationNgapIeType_PDU_RES_MOD_REQ,
							NgapData: &models.RefToBinaryData{
								ContentId: "PDU_RES_MOD",
							},
						},
						SNssai: smContext.SNssai,
					},
				},
			},
		}

		ctx, _, err := smf_context.GetSelf().GetTokenCtx(models.ServiceName_NAMF_COMM, models.NrfNfManagementNfType_AMF)
		if err != nil {
			smContext.Log.Warnf("Get namf-comm token failed: %+v", err)
			return
		} else {
			smContext.Log.Infof("Successfully got NAMF_COMM token for N1N2MessageTransfer")
		}

		rspData, errMsgTransfer := p.Consumer().N1N2MessageTransfer(ctx, smContext.Supi, n1n2Request, smContext.CommunicationClientApiPrefix)
		if errMsgTransfer != nil {
			smContext.Log.Errorf("N1N2MessageTransfer failed: %+v", errMsgTransfer)
			return
		} else {
			smContext.Log.Infof("Successfully sent N1N2MessageTransfer for PDU session modification")
		}

		if rspData.Cause == models.N1N2MessageTransferCause_N1_MSG_NOT_TRANSFERRED {
			smContext.Log.Warnf("N1 message not transferred: %v", rspData.Cause)
		} else {
			smContext.Log.Infof("N1 message transferred successfully")
		}
	}

}

func QueryReport(smContext *smf_context.SMContext, upf *smf_context.UPF,
	urrs []*smf_context.URR, reportResaon models.ChfConvergedChargingTriggerType,
) {
	for _, urr := range urrs {
		urr.State = smf_context.RULE_QUERY
	}

	pfcpState := &PFCPState{
		upf:     upf,
		urrList: urrs,
	}

	resChan := make(chan SendPfcpResult)
	go modifyExistingPfcpSession(smContext, pfcpState, resChan, reportResaon)
	pfcpResult := <-resChan

	if pfcpResult.Err != nil {
		logger.PduSessLog.Errorf("Query URR Report by PFCP Session Mod Request fail: %v", pfcpResult.Err)
		return
	}
}

func establishPfcpSession(smContext *smf_context.SMContext,
	state *PFCPState,
	resCh chan<- SendPfcpResult,
) {
	logger.PduSessLog.Infoln("Sending PFCP Session Establishment Request")

	rcvMsg, err := pfcp_message.SendPfcpSessionEstablishmentRequest(
		state.upf, smContext, state.pdrList, state.farList, state.barList, state.qerList, state.urrList)
	if err != nil {
		logger.PduSessLog.Warnf("Sending PFCP Session Establishment Request error: %+v", err)
		resCh <- SendPfcpResult{
			Status: smf_context.SessionEstablishFailed,
			Err:    err,
		}
		return
	}

	rsp := rcvMsg.PfcpMessage.Body.(pfcp.PFCPSessionEstablishmentResponse)
	if rsp.UPFSEID != nil {
		NodeIDtoIP := rsp.NodeID.ResolveNodeIdToIp().String()
		pfcpSessionCtx := smContext.PFCPContext[NodeIDtoIP]
		pfcpSessionCtx.RemoteSEID = rsp.UPFSEID.Seid
	}

	if rsp.Cause != nil && rsp.Cause.CauseValue == pfcpType.CauseRequestAccepted {
		logger.PduSessLog.Infoln("Received PFCP Session Establishment Accepted Response")
		resCh <- SendPfcpResult{
			Status: smf_context.SessionEstablishSuccess,
			RcvMsg: rcvMsg,
		}
	} else {
		logger.PduSessLog.Infoln("Received PFCP Session Establishment Not Accepted Response")
		resCh <- SendPfcpResult{
			Status: smf_context.SessionEstablishFailed,
			Err:    fmt.Errorf("cause[%d] if not request accepted", rsp.Cause.CauseValue),
		}
	}
}

// func modifyExistingPfcpSession(
// 	smContext *smf_context.SMContext,
// 	state *PFCPState,
// 	resCh chan<- SendPfcpResult,
// 	reportResaon models.ChfConvergedChargingTriggerType,
// ) {
// 	logger.PduSessLog.Infoln("Sending PFCP Session Modification Request")

// 	rcvMsg, err := pfcp_message.SendPfcpSessionModificationRequest(
// 		state.upf, smContext, state.pdrList, state.farList, state.barList, state.qerList, state.urrList)
// 	if err != nil {
// 		logger.PduSessLog.Warnf("Sending PFCP Session Modification Request error: %+v", err)
// 		resCh <- SendPfcpResult{
// 			Status: smf_context.SessionUpdateFailed,
// 			Err:    err,
// 		}
// 		return
// 	}

// 	logger.PduSessLog.Infoln("Received PFCP Session Modification Response")

// 	rsp := rcvMsg.PfcpMessage.Body.(pfcp.PFCPSessionModificationResponse)
// 	if rsp.Cause != nil && rsp.Cause.CauseValue == pfcpType.CauseRequestAccepted {
// 		resCh <- SendPfcpResult{
// 			Status: smf_context.SessionUpdateSuccess,
// 			RcvMsg: rcvMsg,
// 		}
// 		if rsp.UsageReport != nil {
// 			SEID := rcvMsg.PfcpMessage.Header.SEID
// 			upfNodeID := smContext.GetNodeIDByLocalSEID(SEID)
// 			smContext.HandleReports(nil, rsp.UsageReport, nil, upfNodeID, reportResaon)
// 		}
// 	} else {
// 		resCh <- SendPfcpResult{
// 			Status: smf_context.SessionUpdateFailed,
// 			Err:    fmt.Errorf("cause[%d] if not request accepted", rsp.Cause.CauseValue),
// 		}
// 	}
// }

func modifyExistingPfcpSession(
	smContext *smf_context.SMContext,
	state *PFCPState,
	resCh chan<- SendPfcpResult,
	reportResaon models.ChfConvergedChargingTriggerType,
) {
	// Log the variables being sent
	logAsJSON("UPF", state.upf)

	// for i, pdr := range state.pdrList {
	// 	logAsJSON(fmt.Sprintf("PDR[%d]", i), pdr)
	// }
	// for i, far := range state.farList {
	// logAsJSON(fmt.Sprintf("FAR[%d]", i), far)
	// }
	// for i, bar := range state.barList {
	// 	logAsJSON(fmt.Sprintf("BAR[%d]", i), bar)
	// }
	for i, qer := range state.qerList {
		logAsJSON(fmt.Sprintf("QER[%d]", i), qer)
	}

	rcvMsg, err := pfcp_message.SendPfcpSessionModificationRequest(
		state.upf, smContext, state.pdrList, state.farList, state.barList, state.qerList, state.urrList)
	if err != nil {
		logger.PduSessLog.Warnf("Sending PFCP Session Modification Request error: %+v", err)
		resCh <- SendPfcpResult{
			Status: smf_context.SessionUpdateFailed,
			Err:    err,
		}
		return
	}

	// Log the received message
	logger.PduSessLog.Infof("Received PFCP Session Modification Response: %+v", rcvMsg)

	rsp := rcvMsg.PfcpMessage.Body.(pfcp.PFCPSessionModificationResponse)
	if rsp.Cause != nil && rsp.Cause.CauseValue == pfcpType.CauseRequestAccepted {
		logger.PduSessLog.Infof("PFCP Session Modification Response Cause: %+v", rsp.Cause)
		resCh <- SendPfcpResult{
			Status: smf_context.SessionUpdateSuccess,
			RcvMsg: rcvMsg,
		}
		if rsp.UsageReport != nil {
			logger.PduSessLog.Infof("Usage Report: %+v", rsp.UsageReport)
			SEID := rcvMsg.PfcpMessage.Header.SEID
			upfNodeID := smContext.GetNodeIDByLocalSEID(SEID)
			logger.PduSessLog.Infof("SEID: %+v, UPF Node ID: %+v", SEID, upfNodeID)
			smContext.HandleReports(nil, rsp.UsageReport, nil, upfNodeID, reportResaon)
		}
	} else {
		logger.PduSessLog.Warnf("PFCP Session Modification Response Cause: %+v", rsp.Cause)
		resCh <- SendPfcpResult{
			Status: smf_context.SessionUpdateFailed,
			Err:    fmt.Errorf("cause[%d] if not request accepted", rsp.Cause.CauseValue),
		}
	}
}

func waitAllPfcpRsp(
	smContext *smf_context.SMContext,
	pfcpPoolLen int,
	resChan <-chan SendPfcpResult,
	notifyUeHander func(*smf_context.SMContext, bool),
) {
	success := true
	for i := 0; i < pfcpPoolLen; i++ {
		res := <-resChan
		if notifyUeHander == nil {
			continue
		}

		if res.Status == smf_context.SessionEstablishFailed ||
			res.Status == smf_context.SessionUpdateFailed {
			success = false
		}
	}
	if notifyUeHander != nil {
		notifyUeHander(smContext, success)
	}
}

func (p *Processor) EstHandler(isDone <-chan struct{},
	smContext *smf_context.SMContext, success bool,
) {
	// Waiting for Create SMContext Request completed
	if isDone != nil {
		<-isDone
	}
	if success {
		p.sendPDUSessionEstablishmentAccept(smContext)
	} else {
		// TODO: set appropriate 5GSM cause according to PFCP cause value
		p.sendPDUSessionEstablishmentReject(smContext, nasMessage.Cause5GSMNetworkFailure)
	}
}

func ModHandler(smContext *smf_context.SMContext, success bool) {
}

func (p *Processor) sendPDUSessionEstablishmentReject(
	smContext *smf_context.SMContext,
	nasErrorCause uint8,
) {
	smNasBuf, err := smf_context.BuildGSMPDUSessionEstablishmentReject(
		smContext, nasMessage.Cause5GSMNetworkFailure)
	if err != nil {
		logger.PduSessLog.Errorf("Build GSM PDUSessionEstablishmentReject failed: %s", err)
		return
	}

	n1n2Request := models.N1N2MessageTransferRequest{
		BinaryDataN1Message: smNasBuf,
		JsonData: &models.N1N2MessageTransferReqData{
			PduSessionId: smContext.PDUSessionID,
			N1MessageContainer: &models.N1MessageContainer{
				N1MessageClass:   "SM",
				N1MessageContent: &models.RefToBinaryData{ContentId: "GSM_NAS"},
			},
		},
	}

	smContext.SetState(smf_context.InActive)

	ctx, _, errToken := smf_context.GetSelf().GetTokenCtx(models.ServiceName_NAMF_COMM, models.NrfNfManagementNfType_AMF)
	if errToken != nil {
		logger.PduSessLog.Warnf("Get NAMF_COMM context failed: %s", errToken)
		return
	}
	rspData, err := p.Consumer().
		N1N2MessageTransfer(ctx, smContext.Supi, n1n2Request, smContext.CommunicationClientApiPrefix)
	if err != nil || rspData == nil {
		logger.ConsumerLog.Warnf("N1N2MessageTransfer for SendPDUSessionEstablishmentReject failed: %+v", err)
		return
	}

	if rspData.Cause == models.N1N2MessageTransferCause_N1_MSG_NOT_TRANSFERRED {
		logger.PduSessLog.Warnf("%v", rspData.Cause)
	}
	p.RemoveSMContextFromAllNF(smContext, true)
}

func (p *Processor) sendPDUSessionEstablishmentAccept(
	smContext *smf_context.SMContext,
) {
	smNasBuf, err := smf_context.BuildGSMPDUSessionEstablishmentAccept(smContext)
	if err != nil {
		logger.PduSessLog.Errorf("Build GSM PDUSessionEstablishmentAccept failed: %s", err)
		return
	}

	n2Pdu, err := smf_context.BuildPDUSessionResourceSetupRequestTransfer(smContext)
	if err != nil {
		logger.PduSessLog.Errorf("Build PDUSessionResourceSetupRequestTransfer failed: %s", err)
		return
	}

	n1n2Request := models.N1N2MessageTransferRequest{
		BinaryDataN1Message:     smNasBuf,
		BinaryDataN2Information: n2Pdu,
		JsonData: &models.N1N2MessageTransferReqData{
			PduSessionId: smContext.PDUSessionID,
			N1MessageContainer: &models.N1MessageContainer{
				N1MessageClass:   "SM",
				N1MessageContent: &models.RefToBinaryData{ContentId: "GSM_NAS"},
			},
			N2InfoContainer: &models.N2InfoContainer{
				N2InformationClass: models.N2InformationClass_SM,
				SmInfo: &models.N2SmInformation{
					PduSessionId: smContext.PDUSessionID,
					N2InfoContent: &models.N2InfoContent{
						NgapIeType: models.AmfCommunicationNgapIeType_PDU_RES_SETUP_REQ,
						NgapData: &models.RefToBinaryData{
							ContentId: "N2SmInformation",
						},
					},
					SNssai: smContext.SNssai,
				},
			},
		},
	}

	ctx, _, err := smf_context.GetSelf().GetTokenCtx(models.ServiceName_NAMF_COMM, models.NrfNfManagementNfType_AMF)
	if err != nil {
		logger.PduSessLog.Warnf("Get NAMF_COMM context failed: %s", err)
		return
	}

	rspData, err := p.Consumer().
		N1N2MessageTransfer(ctx, smContext.Supi, n1n2Request, smContext.CommunicationClientApiPrefix)
	if err != nil || rspData == nil {
		logger.ConsumerLog.Warnf("N1N2MessageTransfer for sendPDUSessionEstablishmentAccept failed: %+v", err)
		return
	}

	smContext.SetState(smf_context.Active)

	if rspData.Cause == models.N1N2MessageTransferCause_N1_MSG_NOT_TRANSFERRED {
		logger.PduSessLog.Warnf("%v", rspData.Cause)
	}
}

func (p *Processor) updateAnUpfPfcpSession(
	smContext *smf_context.SMContext,
	pdrList []*smf_context.PDR,
	farList []*smf_context.FAR,
	barList []*smf_context.BAR,
	qerList []*smf_context.QER,
	urrList []*smf_context.URR,
) smf_context.PFCPSessionResponseStatus {
	defaultPath := smContext.Tunnel.DataPathPool.GetDefaultPath()
	ANUPF := defaultPath.FirstDPNode
	rcvMsg, err := pfcp_message.SendPfcpSessionModificationRequest(
		ANUPF.UPF, smContext, pdrList, farList, barList, qerList, urrList)
	if err != nil {
		logger.PduSessLog.Warnf("Sending PFCP Session Modification Request to AN UPF error: %+v", err)
		return smf_context.SessionUpdateFailed
	}

	rsp := rcvMsg.PfcpMessage.Body.(pfcp.PFCPSessionModificationResponse)
	if rsp.Cause == nil || rsp.Cause.CauseValue != pfcpType.CauseRequestAccepted {
		logger.PduSessLog.Warn("Received PFCP Session Modification Not Accepted Response from AN UPF")
		return smf_context.SessionUpdateFailed
	}

	logger.PduSessLog.Info("Received PFCP Session Modification Accepted Response from AN UPF")

	if smf_context.GetSelf().ULCLSupport && smContext.BPManager != nil {
		if smContext.BPManager.BPStatus == smf_context.UnInitialized {
			logger.PfcpLog.Infoln("Add PSAAndULCL")
			if err = p.AddPDUSessionAnchorAndULCL(smContext); err != nil {
				logger.PfcpLog.Error(err)
				return smf_context.SessionUpdateFailed
			}
			smContext.BPManager.BPStatus = smf_context.AddingPSA
		}
	}

	return smf_context.SessionUpdateSuccess
}

func ReleaseTunnel(smContext *smf_context.SMContext) []SendPfcpResult {
	resChan := make(chan SendPfcpResult)

	deletedPFCPNode := make(map[string]bool)
	for _, dataPath := range smContext.Tunnel.DataPathPool {
		var targetNodes []*smf_context.DataPathNode
		for node := dataPath.FirstDPNode; node != nil; node = node.Next() {
			targetNodes = append(targetNodes, node)
		}
		dataPath.DeactivateTunnelAndPDR(smContext)
		for _, node := range targetNodes {
			curUPFID, err := node.GetUPFID()
			if err != nil {
				logger.PduSessLog.Error(err)
				continue
			}
			if _, exist := deletedPFCPNode[curUPFID]; !exist {
				go deletePfcpSession(node.UPF, smContext, resChan)
				deletedPFCPNode[curUPFID] = true
			}
		}
	}

	// collect all responses
	resList := make([]SendPfcpResult, 0, len(deletedPFCPNode))
	for i := 0; i < len(deletedPFCPNode); i++ {
		resList = append(resList, <-resChan)
	}

	return resList
}

func deletePfcpSession(upf *smf_context.UPF, ctx *smf_context.SMContext, resCh chan<- SendPfcpResult) {
	logger.PduSessLog.Infoln("Sending PFCP Session Deletion Request")

	rcvMsg, err := pfcp_message.SendPfcpSessionDeletionRequest(upf, ctx)
	if err != nil {
		logger.PduSessLog.Warnf("Sending PFCP Session Deletion Request error: %+v", err)
		resCh <- SendPfcpResult{
			Status: smf_context.SessionReleaseFailed,
			Err:    err,
		}
		return
	}

	rsp := rcvMsg.PfcpMessage.Body.(pfcp.PFCPSessionDeletionResponse)
	if rsp.Cause != nil && rsp.Cause.CauseValue == pfcpType.CauseRequestAccepted {
		logger.PduSessLog.Info("Received PFCP Session Deletion Accepted Response")
		resCh <- SendPfcpResult{
			Status: smf_context.SessionReleaseSuccess,
		}
		if rsp.UsageReport != nil {
			SEID := rcvMsg.PfcpMessage.Header.SEID
			upfNodeID := ctx.GetNodeIDByLocalSEID(SEID)
			ctx.HandleReports(nil, nil, rsp.UsageReport, upfNodeID, "")
		}
	} else {
		logger.PduSessLog.Warn("Received PFCP Session Deletion Not Accepted Response")
		resCh <- SendPfcpResult{
			Status: smf_context.SessionReleaseFailed,
			Err:    fmt.Errorf("cause[%d] if not request accepted", rsp.Cause.CauseValue),
		}
	}
}
