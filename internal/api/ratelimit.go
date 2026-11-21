package api

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/v3/pkg/common/apiresp"
	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/log"
	"server/v3/pkg/common/stability/ratelimit"
	"server/v3/pkg/common/stability/ratelimit/bbr"
)

type RateLimiter struct {
	Enable       bool          `yaml:"enable"`
	Window       time.Duration `yaml:"window"`       // time duration per window
	Bucket       int           `yaml:"bucket"`       // bucket number for each window
	CPUThreshold int64         `yaml:"cpuThreshold"` // CPU threshold; valid range 0–1000 (1000 = 100%)
}

func RateLimitMiddleware(config *RateLimiter) gin.HandlerFunc {
	if !config.Enable {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := bbr.NewBBRLimiter(
		bbr.WithWindow(config.Window),
		bbr.WithBucket(config.Bucket),
		bbr.WithCPUThreshold(config.CPUThreshold),
	)

	return func(c *gin.Context) {
		status := limiter.Stat()

		c.Header("X-BBR-CPU", strconv.FormatInt(status.CPU, 10))
		c.Header("X-BBR-MinRT", strconv.FormatInt(status.MinRt, 10))
		c.Header("X-BBR-MaxPass", strconv.FormatInt(status.MaxPass, 10))
		c.Header("X-BBR-MaxInFlight", strconv.FormatInt(status.MaxInFlight, 10))
		c.Header("X-BBR-InFlight", strconv.FormatInt(status.InFlight, 10))

		done, err := limiter.Allow()
		if err != nil {
			c.Header("X-RateLimit-Policy", "BBR")
			c.Header("Retry-After", calculateBBRRetryAfter(status, config.CPUThreshold))
			c.Header("X-RateLimit-Limit", strconv.FormatInt(status.MaxInFlight, 10))
			c.Header("X-RateLimit-Remaining", "0") // There is no concept of remaining quota in BBR.

			log.ZWarn(c, "rate limited", err, "path", c.Request.URL.Path)
			c.AbortWithStatus(http.StatusTooManyRequests)
			apiresp.GinError(c, errs.NewCodeError(http.StatusTooManyRequests, "too many requests, please try again later"))
			return
		}

		c.Next()
		done(ratelimit.DoneInfo{})
	}
}

func calculateBBRRetryAfter(status bbr.Stat, cpuThreshold int64) string {
	denom := float64(cpuThreshold)
	if denom <= 0 {
		denom = 1000.0
	}
	loadRatio := float64(status.CPU) / denom
	if math.IsNaN(loadRatio) || math.IsInf(loadRatio, 0) || loadRatio < 0.8 {
		return "1"
	}
	if loadRatio < 0.95 {
		return "2"
	}

	backoff := 1 + int64(math.Pow(loadRatio-0.95, 2)*50)
	if backoff < 1 {
		backoff = 1
	}
	if backoff > 5 {
		backoff = 5
	}
	return strconv.FormatInt(backoff, 10)
}
