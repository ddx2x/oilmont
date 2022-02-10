package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/resource/event"
	"github.com/ddx2x/oilmont/pkg/service"
	"math"
	"net/http"
	"os"
	"time"
)

type IAPIServer interface {
	Handler() http.Handler
	Server() *gin.Engine
	RecordEvent(source, action, operator string, obj core.IObject, status event.OperatorStatusType)
}

var _ IAPIServer = &BaseAPIServer{}

type BaseAPIServer struct {
	*gin.Engine
	service.IService
}

func (b BaseAPIServer) Handler() http.Handler { return b }

func (b BaseAPIServer) Server() *gin.Engine { return b.Engine }

func (b BaseAPIServer) RecordEvent(source, action, operator string, obj core.IObject, status event.OperatorStatusType) {
	cloudEvent := &event.CloudEvent{
		Metadata: core.Metadata{
			Name:      obj.GetName(),
			Workspace: obj.GetWorkspace(),
		},
		Spec: event.CloudEventSpec{
			Source:          source,
			Name:            obj.GetName(),
			Action:          action,
			Operator:        operator,
			SourceWorkspace: obj.GetWorkspace(),
			OperatorStatus:  status,
		},
	}
	cloudEvent.GenerateVersion()
	b.Add(cloudEvent)
}

func NewBaseAPIServer(p service.IService) *BaseAPIServer {
	engine := gin.New()
	engine.Use([]gin.HandlerFunc{LoggerMiddleware(log.L), gin.Recovery()}...)
	return &BaseAPIServer{
		engine,
		p,
	}
}

const timeFormat = "02/Jan/2006:15:04:05 -0700"

func LoggerMiddleware(logger log.Logger, notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if _, ok := skip[path]; ok {
			return
		}

		entry := logger.WithFields(
			map[string]interface{}{
				"hostname":   hostname,
				"statusCode": statusCode,
				"latency":    latency, // time to process
				"clientIP":   clientIP,
				"method":     c.Request.Method,
				"path":       path,
				"referer":    referer,
				"dataLength": dataLength,
				"userAgent":  clientUserAgent,
			})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(timeFormat), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
