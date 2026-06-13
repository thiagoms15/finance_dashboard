package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/thiago/finance/backend/internal/auth"
	"github.com/thiago/finance/backend/internal/config"
	"github.com/thiago/finance/backend/internal/domain"
	"github.com/thiago/finance/backend/internal/marketdata"
	"github.com/thiago/finance/backend/internal/repository"
	"github.com/thiago/finance/backend/internal/service"
	"github.com/thiago/finance/backend/internal/transport/http/dto"
	"github.com/thiago/finance/backend/internal/transport/http/middleware"
)

type Server struct {
	cfg          config.Config
	service      *service.AppService
	tokens       *auth.TokenManager
	iconResolver *marketdata.IconResolver
}

func NewRouter(cfg config.Config, service *service.AppService, tokens *auth.TokenManager) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Env == "development" {
		gin.SetMode(gin.DebugMode)
	}

	server := &Server{
		cfg:          cfg,
		service:      service,
		tokens:       tokens,
		iconResolver: marketdata.NewIconResolver(cfg),
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins(),
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if cfg.Env != "production" {
		router.StaticFile("/docs/openapi.yaml", "api/openapi.yaml")
	}

	api := router.Group("/api")
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.RateLimit(5, 10*time.Second))
	{
		authGroup.POST("/register", server.register)
		authGroup.POST("/login", server.login)
		authGroup.POST("/refresh", server.refresh)
		authGroup.POST("/logout", server.logout)
		authGroup.POST("/password-reset/request", server.passwordResetRequest)
		authGroup.POST("/password-reset/confirm", server.passwordResetConfirm)
	}

	protected := api.Group("/")
	protected.Use(middleware.JWTAuth(tokens))
	{
		protected.GET("/assets", server.listAssets)
		protected.POST("/assets", server.createAsset)
		protected.GET("/assets/:id", server.getAsset)
		protected.GET("/assets/:id/icon", server.getAssetIcon)
		protected.GET("/transactions", server.listTransactions)
		protected.POST("/transactions", server.createTransaction)
		protected.PUT("/transactions/:id", server.updateTransaction)
		protected.DELETE("/transactions/:id", server.deleteTransaction)
		protected.GET("/dividends", server.listDividends)
		protected.POST("/dividends", server.createDividend)
		protected.PUT("/dividends/:id", server.updateDividend)
		protected.DELETE("/dividends/:id", server.deleteDividend)
		protected.GET("/portfolio", server.portfolio)
		protected.GET("/portfolio/summary", server.portfolioSummary)
		protected.GET("/portfolio/performance", server.portfolioPerformance)
	}

	return router
}

func (s *Server) register(c *gin.Context) {
	var req dto.RegisterRequest
	if !decodeJSON(c, &req) {
		return
	}

	output, err := s.service.Register(c.Request.Context(), service.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	s.writeAuthResponse(c, http.StatusCreated, output)
}

func (s *Server) login(c *gin.Context) {
	var req dto.LoginRequest
	if !decodeJSON(c, &req) {
		return
	}
	output, err := s.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	s.writeAuthResponse(c, http.StatusOK, output)
}

func (s *Server) logout(c *gin.Context) {
	if err := s.service.Logout(c.Request.Context(), s.refreshTokenFromCookie(c)); err != nil {
		writeError(c, err)
		return
	}
	s.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) refresh(c *gin.Context) {
	output, err := s.service.RefreshSession(c.Request.Context(), s.refreshTokenFromCookie(c))
	if err != nil {
		writeError(c, err)
		return
	}
	s.writeAuthResponse(c, http.StatusOK, output)
}

func (s *Server) passwordResetRequest(c *gin.Context) {
	var req dto.PasswordResetRequest
	if !decodeJSON(c, &req) {
		return
	}

	token, err := s.service.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		writeError(c, err)
		return
	}

	response := gin.H{"ok": true}
	if s.cfg.Env == "development" && token != "" {
		response["resetToken"] = token
	}
	c.JSON(http.StatusAccepted, response)
}

func (s *Server) passwordResetConfirm(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if !decodeJSON(c, &req) {
		return
	}

	if err := s.service.ConfirmPasswordReset(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) listAssets(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	assets, err := s.service.ListAssets(c.Request.Context(), c.Query("search"), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assets})
}

func (s *Server) getAsset(c *gin.Context) {
	assetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid asset id")
		return
	}
	asset, err := s.service.GetAsset(c.Request.Context(), assetID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, asset)
}

func (s *Server) createAsset(c *gin.Context) {
	var req dto.AssetRequest
	if !decodeJSON(c, &req) {
		return
	}

	asset, err := s.service.CreateAsset(c.Request.Context(), service.AssetInput{
		Symbol:   req.Symbol,
		Name:     req.Name,
		Exchange: req.Exchange,
		Currency: req.Currency,
		Sector:   req.Sector,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, asset)
}

func (s *Server) getAssetIcon(c *gin.Context) {
	assetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid asset id")
		return
	}

	asset, err := s.service.GetAsset(c.Request.Context(), assetID)
	if err != nil {
		writeError(c, err)
		return
	}

	icon, err := s.iconResolver.FetchAssetIcon(c.Request.Context(), asset)
	if err != nil {
		if errors.Is(err, marketdata.ErrIconNotFound) {
			writeErrorMessage(c, http.StatusNotFound, "icon_not_found", "icon not available for this asset")
			return
		}
		writeErrorMessage(c, http.StatusBadGateway, "icon_fetch_failed", "unable to fetch asset icon")
		return
	}

	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, icon.ContentType, icon.Body)
}

func (s *Server) listTransactions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	items, err := s.service.ListTransactions(c.Request.Context(), middleware.UserID(c), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (s *Server) createTransaction(c *gin.Context) {
	input, ok := transactionInputFromRequest(c)
	if !ok {
		return
	}
	item, err := s.service.CreateTransaction(c.Request.Context(), middleware.UserID(c), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) updateTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid transaction id")
		return
	}
	input, ok := transactionInputFromRequest(c)
	if !ok {
		return
	}
	item, err := s.service.UpdateTransaction(c.Request.Context(), middleware.UserID(c), transactionID, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) deleteTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid transaction id")
		return
	}
	if err := s.service.DeleteTransaction(c.Request.Context(), middleware.UserID(c), transactionID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) listDividends(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	items, err := s.service.ListDividends(c.Request.Context(), middleware.UserID(c), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (s *Server) createDividend(c *gin.Context) {
	input, ok := dividendInputFromRequest(c)
	if !ok {
		return
	}
	item, err := s.service.CreateDividend(c.Request.Context(), middleware.UserID(c), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *Server) updateDividend(c *gin.Context) {
	dividendID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid dividend id")
		return
	}
	input, ok := dividendInputFromRequest(c)
	if !ok {
		return
	}
	item, err := s.service.UpdateDividend(c.Request.Context(), middleware.UserID(c), dividendID, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) deleteDividend(c *gin.Context) {
	dividendID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_id", "invalid dividend id")
		return
	}
	if err := s.service.DeleteDividend(c.Request.Context(), middleware.UserID(c), dividendID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) portfolio(c *gin.Context) {
	preferred := strings.ToUpper(c.DefaultQuery("currency", "USD"))
	snapshot, err := s.service.PortfolioSnapshot(c.Request.Context(), middleware.UserID(c), preferred)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, snapshot)
}

func (s *Server) portfolioSummary(c *gin.Context) {
	preferred := strings.ToUpper(c.DefaultQuery("currency", "USD"))
	snapshot, err := s.service.PortfolioSnapshot(c.Request.Context(), middleware.UserID(c), preferred)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, snapshot.Summary)
}

func (s *Server) portfolioPerformance(c *gin.Context) {
	preferred := strings.ToUpper(c.DefaultQuery("currency", "USD"))
	points, err := s.service.PortfolioPerformance(c.Request.Context(), middleware.UserID(c), preferred)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": points})
}

func decodeJSON(c *gin.Context, out any) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_json", err.Error())
		return false
	}
	return true
}

func transactionInputFromRequest(c *gin.Context) (service.TransactionInput, bool) {
	var req dto.TransactionRequest
	if !decodeJSON(c, &req) {
		return service.TransactionInput{}, false
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_asset_id", "invalid asset id")
		return service.TransactionInput{}, false
	}
	quantity, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_quantity", "invalid quantity")
		return service.TransactionInput{}, false
	}
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_price", "invalid price")
		return service.TransactionInput{}, false
	}
	fees := decimal.Zero
	if strings.TrimSpace(req.Fees) != "" {
		fees, err = decimal.NewFromString(req.Fees)
		if err != nil {
			writeErrorMessage(c, http.StatusBadRequest, "invalid_fees", "invalid fees")
			return service.TransactionInput{}, false
		}
	}
	transactionDate, err := time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_transaction_date", "transactionDate must be RFC3339")
		return service.TransactionInput{}, false
	}

	return service.TransactionInput{
		AssetID:         assetID,
		Type:            domain.TransactionType(strings.ToUpper(req.Type)),
		Quantity:        quantity,
		Price:           price,
		Fees:            fees,
		Currency:        req.Currency,
		TransactionDate: transactionDate,
		Notes:           req.Notes,
	}, true
}

func dividendInputFromRequest(c *gin.Context) (service.DividendInput, bool) {
	var req dto.DividendRequest
	if !decodeJSON(c, &req) {
		return service.DividendInput{}, false
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_asset_id", "invalid asset id")
		return service.DividendInput{}, false
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_amount", "invalid amount")
		return service.DividendInput{}, false
	}
	paymentDate, err := time.Parse(time.RFC3339, req.PaymentDate)
	if err != nil {
		writeErrorMessage(c, http.StatusBadRequest, "invalid_payment_date", "paymentDate must be RFC3339")
		return service.DividendInput{}, false
	}

	return service.DividendInput{
		AssetID:     assetID,
		Amount:      amount,
		Currency:    req.Currency,
		PaymentDate: paymentDate,
	}, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeErrorMessage(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		writeErrorMessage(c, http.StatusUnauthorized, "invalid_credentials", err.Error())
	case errors.Is(err, service.ErrInvalidSession):
		writeErrorMessage(c, http.StatusUnauthorized, "invalid_session", err.Error())
	case isAccountLocked(err):
		writeErrorMessage(c, http.StatusTooManyRequests, "account_locked", err.Error())
	case errors.Is(err, repository.ErrNotFound):
		writeErrorMessage(c, http.StatusNotFound, "not_found", err.Error())
	default:
		writeErrorMessage(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeErrorMessage(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func (s *Server) writeAuthResponse(c *gin.Context, status int, output service.LoginOutput) {
	s.setRefreshCookie(c, output.RefreshToken, output.RefreshExpiresAt)
	c.JSON(status, output)
}

func (s *Server) refreshTokenFromCookie(c *gin.Context) string {
	token, _ := c.Cookie(s.cfg.RefreshCookieName)
	return token
}

func (s *Server) setRefreshCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(cookieSameSite(s.cfg.RefreshCookieSameSiteMode()))
	c.SetCookie(
		s.cfg.RefreshCookieName,
		token,
		maxAge,
		s.cfg.RefreshCookiePath,
		s.cfg.RefreshCookieDomain,
		s.cfg.RefreshCookieSecure,
		true,
	)
}

func (s *Server) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(cookieSameSite(s.cfg.RefreshCookieSameSiteMode()))
	c.SetCookie(
		s.cfg.RefreshCookieName,
		"",
		-1,
		s.cfg.RefreshCookiePath,
		s.cfg.RefreshCookieDomain,
		s.cfg.RefreshCookieSecure,
		true,
	)
}

func cookieSameSite(mode string) http.SameSite {
	switch strings.ToLower(mode) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func isAccountLocked(err error) bool {
	var lockedErr service.AccountLockedError
	return errors.As(err, &lockedErr)
}
