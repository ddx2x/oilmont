package kes

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ddx2x/oilmont/pkg/k8s"
	"github.com/ddx2x/oilmont/pkg/log"
	"github.com/igm/sockjs-go/v3/sockjs"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// EndOfTransmission terminal end of
const EndOfTransmission = "\u0004"

// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
type OP uint8

const (
	// BIND    fe->be     sessionID      Id sent back from TerminalResponse
	BIND = iota // 0
	// STDIN   fe->be     Data           Keystrokes/paste buffer
	STDIN // 1
	// STDOUT  be->fe     Data           Output from the process
	STDOUT // 2
	// RESIZE  fe->be     Rows, Cols     New terminal size
	RESIZE // 3
	// TOAST   be->fe     Data           OOB message to be shown to the user
	TOAST // 4
	// INEXIT
	INEXIT // 5
	// OUTEXIT
	OUTEXIT // 6
	// ping
	PING
)

type Type = string

const (
	DebugShell  Type = "debug"
	CommonShell Type = "common"
)

var shardingManager *manager

func createGlobalSessionManager(multiCluster *k8s.MultiCluster) {
	if shardingManager == nil {
		shardingManager = &manager{
			multiCluster: multiCluster,
			channels:     make(map[string]*sessionChannel),
			lock:         sync.RWMutex{},
		}
	}
}

type manager struct {
	multiCluster *k8s.MultiCluster
	channels     map[string]*sessionChannel
	lock         sync.RWMutex
}

func (m *manager) get(id string) (*sessionChannel, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	v, exists := m.channels[id]
	return v, exists

}

func (m *manager) set(id string, channel *sessionChannel) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.channels[id] = channel
}

// close shuts down the SockJS connection and sends the status code and reason to the clientv2
// Can happen if the process exits or if there is an error starting up the process
// For now the status code is unused and reason is shown to the user (unless "")
func (m *manager) close(sessionID string, status uint32, reason string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	err := m.channels[sessionID].session.Close(status, reason)
	if err != nil {
		log.G(context.TODO()).Warnf("send closed to client error: %s", err)
		return
	}
	delete(m.channels, sessionID)
}

// PtyHandler is what remotecommand expects from a pty
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
}

// process executed cmd in the container specified in request and connects it up with the  sessionChannel (a manager)
func (m *manager) process(request *attachPodRequest, cmd []string, pty PtyHandler) error {
	clusterName := "default"
	if request.Cluster != "" {
		clusterName = request.Cluster
	}
	cli := m.multiCluster.Get(clusterName)
	if cli == nil {
		return fmt.Errorf("not found cluster: %s", clusterName)
	}

	cmds := []string{"/bin/sh", "-c"}
	cmds = append(cmds, cmd...)
	tty := true
	pod, err := cli.Clientset.CoreV1().Pods(request.Namespace).Get(context.TODO(), request.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("not found namespace %s pod %s", request.Namespace, request.Name)
	}
	if _, exist := pod.GetLabels()["cloud.ddx2x.nip/type"]; exist {
		cmds = []string{}
		tty = false
	}
	options := &v1.PodExecOptions{
		Container: request.Container,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}
	req := cli.Clientset.CoreV1().RESTClient().
		Post().
		Namespace(request.Namespace).
		Resource("pods").
		Name(request.Name).
		SubResource("exec").
		Timeout(0).
		VersionedParams(options, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cli.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("could not make remote command: %v", err)
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             pty,
		Stdout:            pty,
		Stderr:            pty,
		Tty:               true,
		TerminalSizeQueue: pty,
	})
}

// sessionChannel a http connect
// upgrade to websocket session bind a manager channel to backend kubernetes API server with SPDY
type sessionChannel struct {
	id       string
	bound    chan struct{}
	session  sockjs.Session
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
	data     chan []byte
}

// message is the messaging protocol between ShellController and TerminalSession.
type message struct {
	Data, SessionID string
	Rows, Cols      uint16
	Width, Height   uint16
	Op              OP
}

// Next impl sizeChan remote command.TerminalSize
func (s *sessionChannel) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.sizeChan:
		return &size
	case <-s.doneChan:
		return nil
	}
}

// Write impl io.Writer
func (s *sessionChannel) Write(p []byte) (int, error) {
	msg, err := json.Marshal(message{Op: STDOUT, Data: string(p)})
	if err != nil {
		return 0, err
	}
	if err = s.session.Send(string(msg)); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Toast can be used to send the user any OOB messages
// hterm puts these in the center of the terminal
func (s *sessionChannel) Toast(p string) error {
	msg, err := json.Marshal(message{Op: TOAST, Data: p})
	if err != nil {
		return err
	}
	if err = s.session.Send(string(msg)); err != nil {
		return err
	}
	return nil
}

func (s *sessionChannel) Read(p []byte) (n int, err error) {
	buf, err := s.session.Recv()
	if err != nil {
		return 0, err
	}
	msg := &message{}
	if buf == "PING" {
		msg.Op = PING
		goto HANDLE
	}
	err = json.Unmarshal([]byte(buf), msg)
	if err != nil {
		return copy(p, EndOfTransmission), err
	}
HANDLE:
	switch msg.Op {
	case STDIN:
		return copy(p, msg.Data), nil
	case INEXIT: // exit from client v2 event
		return 0, fmt.Errorf("client v2 exit")
	case RESIZE:
		s.sizeChan <- remotecommand.TerminalSize{Width: msg.Width, Height: msg.Height}
		return 0, nil
	case PING:
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unknown message type '%d'", msg.Op)
	}
}

// generateTerminalSessionId generates a random manager ID string. The format is not really interesting.
// This ID is used to identify the manager when the clientv2 opens the SockJS connection.
// Not the same as the SockJS manager id! We can't use that as that is generated
// on the client v2 side and we don't have it yet at this point.
func generateTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

// isValidShell checks if the shell is an allowed one
func isValidShell(validShells []string, shell string) bool {
	for _, validShell := range validShells {
		if validShell == shell {
			return true
		}
	}
	return false
}

func (m *manager) remoteExecute(cfg *rest.Config, url *url.URL, pty PtyHandler, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", url)
	if err != nil {
		return err
	}
	return exec.Stream(
		remotecommand.StreamOptions{
			Stdin:             pty,
			Stdout:            pty,
			Stderr:            pty,
			Tty:               tty,
			TerminalSizeQueue: pty,
		},
	)
}

func (m *manager) getContainerIDByName(pod *v1.Pod, containerName string) (string, error) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name != containerName {
			continue
		}
		// #52 if a pod is running but not ready(because of readiness probe), we can connect
		if containerStatus.State.Running == nil {
			return "", fmt.Errorf("container %s not running", containerName)
		}
		return containerStatus.ContainerID, nil
	}

	// #14 otherwise we should search for running init containers
	for _, initContainerStatus := range pod.Status.InitContainerStatuses {
		if initContainerStatus.Name != containerName {
			continue
		}
		if initContainerStatus.State.Running == nil {
			return "", fmt.Errorf("init container %s is not running", containerName)
		}

		return initContainerStatus.ContainerID, nil
	}

	return "", fmt.Errorf("cannot find specified container %s", containerName)
}

func (m *manager) launchDebugPod(request *attachPodRequest, pty PtyHandler) error {
	clusterName := "default"
	if request.Cluster != "" {
		clusterName = request.Cluster
	}
	cli := m.multiCluster.Get(clusterName)
	if cli == nil {
		return fmt.Errorf("not found cluster: %s", clusterName)
	}
	ctx := context.Background()

	pod := v1.Pod{}
	err := cli.RESTClient().
		Get().
		Resource("pods").
		Namespace(request.Namespace).
		Name(request.Name).
		Do(ctx).
		Into(&pod)
	if err != nil {
		return err
	}
	containerID, err := m.getContainerIDByName(&pod, request.Container)
	if err != nil {
		return err
	}
	uri, err := url.Parse(fmt.Sprintf("http://%s:%d", pod.Status.HostIP, 10027))
	if err != nil {
		return err
	}
	uri.Path = fmt.Sprintf("/api/v1/debug")
	params := url.Values{}

	if request.Image == "" {
		request.Image = "nicolaka/netshoot:latest"
	}
	params.Add("image", request.Image)
	params.Add("container", containerID)
	params.Add("verbosity", fmt.Sprintf("%s", "0"))
	hstNm, _ := os.Hostname()
	params.Add("hostname", hstNm)
	params.Add("username", "")
	//TODO: should be set false
	params.Add("lxcfsEnabled", "true")
	params.Add("registrySkipTLS", "false")
	params.Add("authStr", "")

	//TODO: support private registry pull image,just like harbor.
	//var authStr string
	//registrySecret, err := o.CoreClient.Secrets(o.RegistrySecretNamespace).Get(o.RegistrySecretName, v1.GetOptions{})
	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		authStr = ""
	//	} else {
	//		return err
	//	}
	//} else {
	//	authStr = string(registrySecret.Data["authStr"])
	//}

	cmd := []string{"/bin/bash"}
	commandBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	params.Add("command", string(commandBytes))
	uri.RawQuery = params.Encode()
	return m.remoteExecute(cli.Config, uri, pty, true)
}

func (m *manager) launchCommonPod(request *attachPodRequest, session *sessionChannel) error {
	var errs error
	var err error
	validShells := []string{"bash", "sh", "csh"}
	if isValidShell(validShells, request.Shell) {
		cmd := []string{request.Shell}
		errs = shardingManager.process(request, cmd, session)
	} else {
		// No shell given or it was not valid: try some shells until one succeeds or all fail
		for _, testShell := range validShells {
			cmd := []string{testShell}
			if err = shardingManager.process(request, cmd, session); err == nil {
				break
			} else {
				errs = multierr.Append(errs, err)
			}
		}
	}
	return errs
}

// waitForTerminal is called from pod attach api as a goroutine
// Waits for the SockJS connection to be opened by the client v2 the manager to be bound in handleTerminalSession
func waitForTerminal(request *attachPodRequest, sessId string) {
	sc, exist := shardingManager.get(sessId)
	if !exist {
		return
	}
	<-sc.bound
	defer close(sc.bound)
	var err error

	switch request.ShellType {
	case DebugShell:
		err = shardingManager.launchDebugPod(request, sc)
	default:
		err = shardingManager.launchCommonPod(request, sc)
	}
	if err != nil {
		shardingManager.close(sessId, 2, err.Error())
		return
	}
	shardingManager.close(sessId, 1, "process exited")
}

func wrapSockjsHandle(ctx context.Context) func(session sockjs.Session) {
	return func(session sockjs.Session) {
		flog := log.G(ctx).WithField("sessionId", session.ID())
		buf, err := session.Recv()
		if err != nil {
			flog.Warnf("sockjs session recv resize data %s error: %s", buf, err)
			return
		}
		msg := &message{}
		if err = json.Unmarshal([]byte(buf), &msg); err != nil {
			flog.Warnf("resize data %s unmarshal error: %s", buf, err)
			return
		}
		resize := remotecommand.TerminalSize{Width: msg.Width, Height: msg.Height}

		buf, err = session.Recv()
		if err != nil {
			flog.Warnf("sockjs session recv bind data %s error: %s", buf, err)
			return
		}
		msg = &message{}
		if err = json.Unmarshal([]byte(buf), &msg); err != nil {
			flog.Warnf("bind data %s unmarshal error: %s", buf, err)
			return
		}

		if msg.Op != BIND {
			if err := session.Send(fmt.Sprintf(`{"Op":%d,"Data":%s}`, OUTEXIT, "command exception")); err != nil {
				flog.Warnf("resp data error:%s", err)
			}
			return
		}

		count := 0
	RETRY:
		sessionChannel, exist := shardingManager.get(msg.SessionID)
		if !exist {
			time.Sleep(1 * time.Second)
			count++
			if count >= 3 {
				bs, _ := json.Marshal(message{Op: OUTEXIT, Data: "connection manager expired, please close and reconnect"})
				if err := session.Send(string(bs)); err != nil {
					flog.Warnf("send manager message to client error: %s", err)
				}
				return
			}
			goto RETRY
		}

		if sessionChannel.id == "" {
			fmt.Printf("handleTerminalSession: can't find manager '%s'", msg.SessionID)
			return
		}
		sessionChannel.session = session
		shardingManager.set(msg.SessionID, sessionChannel)
		sessionChannel.bound <- struct{}{}
		sessionChannel.sizeChan <- resize
	}
}
