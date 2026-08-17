package httptransport

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry043/internal/application"
	"github.com/wyw14/cry043/internal/domain"
)

type Services struct {
	Compliance  *application.ComplianceService
	Inspections *application.InspectionService
	Risks       *application.RiskService
	Readiness   func() error
}

type Handler struct {
	services Services
	validate *validator.Validate
}

func New(services Services, middleware ...gin.HandlerFunc) *gin.Engine {
	h := &Handler{services: services, validate: validator.New()}
	engine := gin.New()
	engine.Use(middleware...)
	h.registerProbes(engine)
	h.registerAPI(engine.Group("/api/v1"))
	return engine
}

func (h *Handler) registerProbes(engine *gin.Engine) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "component": "compliance-control"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if err := h.services.Readiness(); err != nil {
			writeFailure(c, http.StatusServiceUnavailable, "DEPENDENCY_NOT_READY", err, nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}

func (h *Handler) registerAPI(api *gin.RouterGroup) {
	specifications := api.Group("/specifications")
	specifications.POST("/:id/simulations", h.allow("author", "approver"), h.simulate)
	specifications.POST("/:id/activations", h.allow("approver"), h.activate)
	specifications.POST("/:id/acknowledgements", h.acknowledge)
	specifications.POST("/:id/compliance-checks", h.evaluate)
	specifications.GET("/:id/activation-readiness", h.allow("author", "approver", "supervisor"), h.activationReadiness)

	assurance := api.Group("/field-assurance")
	assurance.POST("/inspections", h.allow("inspector"), h.recordInspection)
	assurance.POST("/remediations/:id/closures", h.allow("reviewer"), h.closeRemediation)

	controlRoom := api.Group("/control-room")
	controlRoom.Use(h.allow("supervisor", "approver", "inspector"))
	controlRoom.GET("/risks", h.riskBoard)
}

func (h *Handler) allow(roles ...string) gin.HandlerFunc {
	accepted := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		accepted[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := accepted[c.GetHeader("X-Actor-Role")]; !ok {
			writeFailure(c, http.StatusForbidden, "ROLE_NOT_ALLOWED", errors.New("当前角色无权访问该合规工作区"), nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) simulate(c *gin.Context) {
	result, err := h.services.Compliance.Simulate(c, c.Param("id"), actorID(c))
	writeResult(c, http.StatusOK, result, err)
}

func (h *Handler) activate(c *gin.Context) {
	err := h.services.Compliance.Activate(c, c.Param("id"), actorID(c))
	writeResult(c, http.StatusOK, gin.H{"activated": err == nil}, err)
}

func (h *Handler) acknowledge(c *gin.Context) {
	version, err := strconv.Atoi(c.Query("version"))
	if err != nil || version < 1 {
		writeFailure(c, http.StatusUnprocessableEntity, "VERSION_REQUIRED", errors.New("version 必须是正整数"), map[string]string{"version": "请提供当前生效版本"})
		return
	}
	ack, err := h.services.Compliance.Acknowledge(c, c.Param("id"), actorID(c), c.GetHeader("X-Team-ID"), version)
	writeResult(c, http.StatusCreated, ack, err)
}

func (h *Handler) evaluate(c *gin.Context) {
	var input domain.ComplianceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeFailure(c, http.StatusBadRequest, "MALFORMED_COMPLIANCE_INPUT", err, nil)
		return
	}
	violations, err := h.services.Compliance.Evaluate(c, c.Param("id"), input)
	writeResult(c, http.StatusOK, gin.H{"violations": violations, "compliant": err == nil && len(violations) == 0}, err)
}

func (h *Handler) activationReadiness(c *gin.Context) {
	result, err := h.services.Risks.Readiness(c, c.Param("id"))
	writeResult(c, http.StatusOK, result, err)
}

func (h *Handler) recordInspection(c *gin.Context) {
	var input domain.Inspection
	if err := c.ShouldBindJSON(&input); err != nil {
		writeFailure(c, http.StatusBadRequest, "MALFORMED_INSPECTION", err, nil)
		return
	}
	remediations, err := h.services.Inspections.Record(c, input, actorID(c))
	writeResult(c, http.StatusCreated, gin.H{"remediations": remediations}, err)
}

func (h *Handler) closeRemediation(c *gin.Context) {
	err := h.services.Inspections.Close(c, c.Param("id"))
	writeResult(c, http.StatusOK, gin.H{"closed": err == nil}, err)
}

func (h *Handler) riskBoard(c *gin.Context) {
	days := 14
	if raw := c.Query("horizon_days"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 90 {
			writeFailure(c, http.StatusUnprocessableEntity, "INVALID_RISK_WINDOW", errors.New("horizon_days 必须在 1 到 90 之间"), map[string]string{"horizon_days": "允许 1..90"})
			return
		}
		days = parsed
	}
	result, err := h.services.Risks.Board(c, strings.TrimSpace(c.Query("area_id")), strings.TrimSpace(c.Query("process_id")), time.Duration(days)*24*time.Hour)
	writeResult(c, http.StatusOK, result, err)
}

func actorID(c *gin.Context) string {
	if actor := strings.TrimSpace(c.GetHeader("X-Actor-ID")); actor != "" {
		return actor
	}
	return "demo-worker"
}

func writeResult(c *gin.Context, successStatus int, value any, err error) {
	if err == nil {
		c.JSON(successStatus, gin.H{"data": value, "request_id": c.GetString("request_id")})
		return
	}
	code := "WORKFLOW_CONFLICT"
	if errors.Is(err, domain.ErrSafetyException) {
		code = "MANDATORY_SAFETY_EXCEPTION_DENIED"
	}
	writeFailure(c, http.StatusConflict, code, err, nil)
}

func writeFailure(c *gin.Context, status int, code string, err error, fields map[string]string) {
	if fields == nil {
		fields = map[string]string{}
	}
	c.JSON(status, gin.H{
		"code": code, "message": err.Error(), "field_errors": fields,
		"request_id": c.GetString("request_id"), "occurred_at": time.Now().UTC(),
	})
}
