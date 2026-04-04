package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kbfood/internal/infra/platform"
	"kbfood/internal/interface/http/dto"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// SyncHandler handles manual sync operations
type SyncHandler struct {
	syncJob   SyncJobRunner
	tttClient *platform.TanTanTangClient
}

// SyncJobRunner interface for jobs that can be manually triggered
type SyncJobRunner interface {
	Run(ctx context.Context) error
	Name() string
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncJob SyncJobRunner, tttClient *platform.TanTanTangClient) *SyncHandler {
	return &SyncHandler{
		syncJob:   syncJob,
		tttClient: tttClient,
	}
}

// TriggerSync handles POST /api/admin/sync
func (h *SyncHandler) TriggerSync(c echo.Context) error {
	ctx := c.Request().Context()

	log.Info().Msg("Manual sync triggered")

	if err := h.syncJob.Run(ctx); err != nil {
		log.Error().Err(err).Msg("Manual sync failed")
		return c.JSON(http.StatusInternalServerError, dto.Error(500, "sync operation failed"))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]string{
		"message": "Sync completed successfully",
		"job":     h.syncJob.Name(),
	}))
}

// TestAPI handles GET /api/admin/test-api - test TanTanTang API directly
func (h *SyncHandler) TestAPI(c echo.Context) error {
	ctx := c.Request().Context()
	region := c.QueryParam("region")
	if region == "" {
		region = "长沙"
	}

	// Get region config for coords
	cityCfg := platform.GetRegionConfig(region)

	// Test API call
	products, err := h.tttClient.FetchProducts(ctx, region)
	if err != nil {
		log.Error().Err(err).Str("region", region).Msg("API test failed")
		return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"region":      region,
			"success":     false,
			"error":       err.Error(),
			"count":       0,
			"products":    []interface{}{},
			"serverTime":  time.Now().Format("2006-01-02 15:04:05"),
		}))
	}

	return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"region":      region,
		"city":        cityCfg.CityName,
		"success":     true,
		"count":       len(products),
		"sampleCount": min(len(products), 3),
		"products":    products,
		"serverTime":  time.Now().Format("2006-01-02 15:04:05"),
	}))
}

// TestNotification handles POST /api/admin/test-notification
// Sends a test notification via ntfy to verify the topic is valid
func (h *SyncHandler) TestNotification(c echo.Context) error {
	var params struct {
		NtfyTopic string `json:"ntfyTopic"`
		BarkKey   string `json:"barkKey"` // backward compatibility
	}

	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, dto.Error(400, "Invalid request format"))
	}

	topicInput := strings.TrimSpace(params.NtfyTopic)
	if topicInput == "" {
		topicInput = strings.TrimSpace(params.BarkKey)
	}

	if topicInput == "" {
		return c.JSON(http.StatusBadRequest, dto.Error(400, "ntfy topic is required"))
	}

	// User can provide either full topic URL or topic name.
	topic := normalizeAdminNtfyTopic(topicInput)
	if topic == "" {
		return c.JSON(http.StatusBadRequest, dto.Error(400, "invalid ntfy topic"))
	}

	baseURL := "https://ntfy.sh"
	if strings.HasPrefix(strings.ToLower(topicInput), "http") {
		if parsed, err := url.Parse(topicInput); err == nil {
			baseURL = parsed.Scheme + "://" + parsed.Host
		}
	}

	fullURL := strings.TrimRight(baseURL, "/") + "/" + url.PathEscape(topic)
	body := "测试通知：美食监控配置成功！您可以收到价格提醒了。"

	log.Info().Str("ntfyURL", fullURL).Msg("Sending test notification")

	// Send test notification
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBufferString(body))
	if err != nil {
		return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"success":  false,
			"error":    err.Error(),
			"testedAt": time.Now().Format("2006-01-02 15:04:05"),
		}))
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Title", "测试通知")
	req.Header.Set("Priority", "5")

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Test notification request failed")
		return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
			"success":  false,
			"error":    err.Error(),
			"testedAt": time.Now().Format("2006-01-02 15:04:05"),
		}))
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	log.Info().Bool("success", success).Int("statusCode", resp.StatusCode).Msg("Test notification completed")

	return c.JSON(http.StatusOK, dto.Success(map[string]interface{}{
		"success":    success,
		"statusCode": resp.StatusCode,
		"testedAt":   time.Now().Format("2006-01-02 15:04:05"),
	}))
}

func normalizeAdminNtfyTopic(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(trimmed), "http") {
		if parsed, err := url.Parse(trimmed); err == nil {
			path := strings.Trim(parsed.Path, "/")
			if path == "" {
				return ""
			}
			parts := strings.Split(path, "/")
			return parts[len(parts)-1]
		}
	}

	return trimmed
}
