package gateway

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/core"
	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/ddx2x/oilmont/pkg/utils/uri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	STREAM_END   string = "STREAM_END"
	STREAM_ERROR string = "STREAM_ERROR"
	PING         string = "ping"
	USER_CONFIG  string = "USER_CONFIG"
)

type watcherEvent struct {
	Type       core.EventType `json:"type"`
	Object     interface{}    `json:"object"`
	UserConfig interface{}    `json:"userConfig"`
	URL        string         `json:"url"`
	Status     int            `json:"status"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (gw *Gateway) watch(g *gin.Context) {
	watcherEventChan := make(chan watcherEvent, 32)
	errorCh := make(chan error)
	fullURL := g.Request.URL
	ctx, cancel := context.WithCancel(context.Background())

	path := g.Request.URL.Path
	start := time.Now()
	stop := time.Since(start)
	latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
	statusCode := g.Writer.Status()
	clientIP := g.ClientIP()
	clientUserAgent := g.Request.UserAgent()
	referer := g.Request.Referer()

	flog := log.G(ctx).WithFields(map[string]interface{}{
		"statusCode": statusCode,
		"latency":    latency, // time to process
		"clientIP":   clientIP,
		"method":     g.Request.Method,
		"path":       path,
		"referer":    referer,
		"userAgent":  clientUserAgent,
	})

	go gw.asyncWatch(flog, ctx, fullURL, watcherEventChan, g, errorCh)

	endEvent := watcherEvent{
		Type:   STREAM_END,
		URL:    fullURL.String(),
		Status: 410,
	}

	ticker := time.NewTicker(10 * time.Second)

	clientUniques := RandStringRunes(10)

	defer func() {
		flog.Warnf("-----END----- \r\n close long connection: %s \n  id: %s \r\n",
			fullURL.String(), clientUniques,
		)
	}()

	flog.Warnf("-----BEGIN----- \r\n watch start long connection: %s \r\n id: %s \r\n",
		fullURL.String(), clientUniques,
	)

	g.Stream(func(w io.Writer) bool {
		select {
		case <-g.Writer.CloseNotify(): //client close
			cancel()
			return false

		case err := <-errorCh: // watch process error
			if err == nil {
				return false
			}
			endEvent.Object = err
			g.SSEvent("", endEvent)
			return false

		case event, ok := <-watcherEventChan: // event send
			if !ok {
				g.SSEvent("", endEvent)
				return false
			}
			flog.Warnf(
				"-----PROCESS----- \r\n send event to id: %s {type:%s,object:%s } \r\n",
				clientUniques, event.Type, event.Object,
			)
			g.SSEvent("", event)

		case <-ticker.C: // ticker check
			//userName := g.GetHeader(common.HttpRequestUserHeaderKey)
			//cfg, err := gw.userCfg(userName, "", true)
			//if err == nil {
			//	g.SSEvent("", watcherEvent{
			//		Type:       USER_CONFIG,
			//		URL:        clientUniques,
			//		UserConfig: cfg,
			//		Status:     200,
			//	})
			//}
			g.SSEvent("",
				watcherEvent{
					Type:   PING,
					URL:    clientUniques,
					Status: 200,
				})
		}

		return true
	})
}

func (gw *Gateway) isKubeResource(uri *uri.URI) bool {
	_, err := k8s.ShardingResourceRegistry.GetGVR(uri.Resource)
	if err == nil {
		return true
	}
	return false
}

func (gw *Gateway) watchMgo(ctx context.Context, uri *uri.URI, writer chan<- watcherEvent, stopC <-chan struct{}) error {
	db, table := common.GetResourceTable(ctx, gw.stage, uri.Tenant, uri.Resource)
	var filter datasource.Filter
	if uri.Namespace != "" {
		filter = datasource.Filter{
			Key:   "metadata.workspace",
			Value: uri.Namespace,
		}
	}
	cloudChs, err := gw.stage.WatchEvent(ctx, db, table, uri.ResourceVersion, filter)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-stopC:
				return
			case event, ok := <-cloudChs:
				if !ok {
					return
				}
				writer <- watcherEvent{
					Type:   event.Type,
					Object: event.Object,
				}
			}
		}
	}()

	return nil
}

func (gw *Gateway) watchK8s(ctx context.Context, cluster string, uri *uri.URI, writer chan<- watcherEvent, stopC <-chan struct{}) error {
	gvr, err := k8s.ShardingResourceRegistry.GetGVR(uri.Resource)
	if err != nil {
		return err
	}
	cli := gw.mc.Get(cluster)
	if cli == nil {
		return fmt.Errorf("multi cluster not found cluster %s", cluster)
	}
	listOptions := metav1.ListOptions{
		ResourceVersion: uri.ResourceVersion,
	}
	watchInterface, err := cli.Interface.Resource(gvr).Namespace(uri.Namespace).Watch(ctx, listOptions)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-stopC:
				watchInterface.Stop()
				return
			case event, ok := <-watchInterface.ResultChan():
				if !ok {
					close(writer)
					return
				}
				Type := core.ADDED
				switch event.Type {
				case watch.Modified:
					Type = core.MODIFIED
				case watch.Deleted:
					Type = core.DELETED
				}
				writer <- watcherEvent{
					Type:   Type,
					Object: event.Object,
				}
			}
		}
	}()

	return nil
}

func (gw *Gateway) asyncWatch(
	flog log.Logger,
	ctx context.Context,
	url *url.URL,
	writeEventChan chan<- watcherEvent,
	g *gin.Context,
	errC chan<- error,
) {
	flog = flog.WithField("function", "asyncWatch")
	uris, err := gw.parser.ParseWatch(url)
	if err != nil {
		errC <- err
	}
	userName := g.GetHeader(common.HttpRequestUserHeaderKey)
	if userName == "" {
		errC <- fmt.Errorf("NotAuthorized Account")
	}
	tenant := g.GetHeader(common.HttpRequestUserHeaderTENANT)

	uris, err = gw.perm.checkURIsAndReorganize(tenant, userName, uris)
	if err != nil {
		flog.Warnf("check uris and reorganize uri error: %s", err)
	}

	closeChs := make([]chan struct{}, 0)
	wg := sync.WaitGroup{}
	wg.Add(len(uris))

	writer := make(chan watcherEvent)
	for _, uri := range uris {
		closeCh := make(chan struct{})
		if gw.isKubeResource(uri) {
			flog.Infof("watch k8s resource %s from tenant %s", uri.Resource, uri.Tenant)
			if err := gw.watchK8s(ctx, uri.Cluster, uri, writer, closeCh); err != nil {
				writeEventChan <- watcherEvent{
					Type:   STREAM_ERROR,
					Status: 500,
				}
				errC <- err
				flog.Warnf("collect cluster %s cloud resource %s event chan error: %s", uri.Cluster, uri.Resource, err)
				continue
			}
		} else {
			flog.Infof("watch cloud platform resource %s from tenant %s", uri.Resource, uri.Tenant)
			if err := gw.watchMgo(ctx, uri, writer, closeCh); err != nil {
				writeEventChan <- watcherEvent{
					Type:   STREAM_ERROR,
					Status: 500,
				}
				errC <- err
				flog.Warnf("collect hybrid cloud resource %s event chan error: %s", uri.Resource, err)
				continue
			}
		}

		go func() {
			defer wg.Done()
			for {
				select {
				case event, ok := <-writer:
					if !ok {
						return
					}
					evt := watcherEvent{
						Type:   event.Type,
						Object: event.Object,
					}
					writeEventChan <- evt
				}
			}

		}()
		closeChs = append(closeChs, closeCh)
	}

	go func() {
		<-ctx.Done()
		for _, c := range closeChs {
			c <- struct{}{}
		}
	}()

	wg.Wait()
}
