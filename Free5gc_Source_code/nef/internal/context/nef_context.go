package context

import (
	"context"
	"fmt"
	"sync"

	"github.com/free5gc/nef/internal/logger"
	// "github.com/free5gc/nef/internal/sbi/processor"
	"github.com/free5gc/nef/pkg/factory"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/oauth"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	// "github.com/free5gc/util/httpwrapper"
	 "time"
)

type nef interface {
	Config() *factory.Config
}

type NefContext struct {
	nef
NfService       map[ServiceName]models.NfService
	nfInstID       string // NF Instance ID
	pcfPaUri       string
	udrDrUri       string
	NrfCertPem      string
	numCorreID     uint64
	OAuth2Required bool
	afs            map[string]*AfData
	mu             sync.RWMutex
}
var nefContext =NefContext{}
type Nfcontext interface {
	AuthorizationCheck(token string) error
	}

var _ Nfcontext = &NefContext{}
func NewContext(nef nef) (*NefContext, error) {
	c := &NefContext{
		nef:      nef,
		nfInstID: uuid.New().String(),
	}
	c.afs = make(map[string]*AfData)
	logger.CtxLog.Infof("New nfInstID: [%s]", c.nfInstID)
	return c, nil
}

func (c *NefContext) NfInstID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nfInstID
}

func (c *NefContext) SetNfInstID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nfInstID = id
	logger.CtxLog.Infof("Set nfInstID: [%s]", c.nfInstID)
}

func (c *NefContext) PcfPaUri() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pcfPaUri
}

func (c *NefContext) SetPcfPaUri(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pcfPaUri = uri
	logger.CtxLog.Infof("Set pcfPaUri: [%s]", c.pcfPaUri)
}

func (c *NefContext) UdrDrUri() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.udrDrUri
}

func (c *NefContext) SetUdrDrUri(uri string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.udrDrUri = uri
	logger.CtxLog.Infof("Set udrDrUri: [%s]", c.udrDrUri)
}
func (c *NefContext) NewAf(afID string) *AfData {
	af := &AfData{
		AfID:     afID,
		Subs:     make(map[string]*AfSubscription),
		PfdTrans: make(map[string]*AfPfdTransaction),
		Log:      logger.CtxLog.WithField(logger.FieldAFID, fmt.Sprintf("AF:%s", afID)),
	}
	return af
}
func (c *NefContext) NewQosAf(scsAsId string) *AfQosData {
	afQos := &AfQosData{
		scsAsID:  scsAsId,
		Subs:     make(map[string]*AfSubscription),
		PfdTrans: make(map[string]*AfPfdTransaction),
		Log:      logger.CtxLog.WithField(logger.FieldAFID, fmt.Sprintf("AF:%s", scsAsId)),
	}
	return afQos
}
func (c *NefContext) AddAf(af *AfData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afs[af.AfID] = af
	af.Log.Infoln("AF is added")
}

func (c *NefContext) GetAf(afID string) *AfData {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.afs[afID]
}

func (c *NefContext) DeleteAf(afID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.afs, afID)
	logger.CtxLog.Infof("AF[%s] is deleted", afID)
}

func (c *NefContext) NewCorreID() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.numCorreID++
	return c.numCorreID
}

func (c *NefContext) ResetCorreID() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.numCorreID = 0
}

func (c *NefContext) IsAppIDExisted(appID string) (string, string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, af := range c.afs {
		af.Mu.RLock()
		if transID, ok := af.IsAppIDExisted(appID); ok {
			defer af.Mu.RUnlock()
			return af.AfID, transID, true
		}
		af.Mu.RUnlock()
	}
	return "", "", false
}

func (c *NefContext) FindAfSub(CorrID string) (*AfData, *AfSubscription) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, af := range c.afs {
		af.Mu.RLock()
		for _, sub := range af.Subs {
			if sub.NotifCorreID == CorrID {
				defer af.Mu.RUnlock()
				return af, sub
			}
		}
		af.Mu.RUnlock()
	}
	return nil, nil
}

func (c *NefContext) GetTokenCtx(serviceName models.ServiceName, targetNF models.NfType) (
	context.Context, *models.ProblemDetails, error,
) {
	if !c.OAuth2Required {
		return context.TODO(), nil, nil
	}
	return oauth.GetTokenCtx(models.NfType_NEF, targetNF,
		c.nfInstID, c.Config().NrfUri(), string(serviceName))
}


var jwtSigningKey = []byte("super-secret-key") // MUST match the one used during token generation
//  Main token validation logic
func (c *NefContext) AuthorizationCheck(token string) error {

	if !c.OAuth2Required {
		logger.UtilLog.Debugf("AuthorizationCheck: OAuth2 not required")
		return nil
	}

	logger.UtilLog.Debugf("AuthorizationCheck: token [%s]", token)

	if token == "" {
		logger.UtilLog.Errorf("AuthorizationCheck: token is empty")
		return fmt.Errorf("token is empty")
	}

	// Strip "Bearer " prefix if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
		token = strings.TrimSpace(token)
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSigningKey, nil
	})

	if err != nil {
		logger.UtilLog.Errorf("AuthorizationCheck: token parse error: %v", err)
		return fmt.Errorf("invalid token")
	}

	if !parsedToken.Valid {
		logger.UtilLog.Errorf("AuthorizationCheck: token is not valid")
		return fmt.Errorf("invalid token")
	}

	// Inspect claims (optional)
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
			return fmt.Errorf("token expired")
		}
		logger.UtilLog.Debugf("AuthorizationCheck: Token validated. Claims: %v", claims)
	}

	return nil
}
