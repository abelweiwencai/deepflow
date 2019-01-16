package profiler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"gitlab.x.lan/yunshan/droplet-libs/debug"

	"gitlab.x.lan/yunshan/droplet/dropletctl"
)

type ProfilerServer struct {
	http.Server
	port int

	running bool
}

const (
	CONFIG_CMD_PROFILER_ON = iota
	CONFIG_CMD_PROFILER_OFF
	CONFIG_CMD_PROFILER_STATUS
)

var log = logging.MustGetLogger("profile")

func NewProfiler(port int) *ProfilerServer {
	server := &ProfilerServer{http.Server{Addr: "0.0.0.0:" + strconv.Itoa(port)}, port, false}
	debug.Register(dropletctl.DROPLETCTL_CONFIG, server)
	return server
}

func (s *ProfilerServer) RecvCommand(conn *net.UDPConn, remote *net.UDPAddr, operate uint16, args *bytes.Buffer) {
	switch operate {
	case CONFIG_CMD_PROFILER_ON:
		s.recvProfilerOn(conn, remote, args)
	case CONFIG_CMD_PROFILER_OFF:
		s.recvProfilerOff(conn, remote, args)
	case CONFIG_CMD_PROFILER_STATUS:
		s.recvProfilerRunningStatus(conn, remote, args)
	}
}

func sendToDropletCtl(conn *net.UDPConn, remote *net.UDPAddr, info interface{}) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(info); err != nil {
		log.Errorf("encoder.Encode: %s", err)
		debug.SendToClient(conn, remote, 1, nil)
		return
	}
	debug.SendToClient(conn, remote, 0, &buffer)
}

func (s *ProfilerServer) recvProfilerOn(conn *net.UDPConn, remote *net.UDPAddr, args *bytes.Buffer) {
	s.CheckAndStart()

	if s.isRunning() {
		sendToDropletCtl(conn, remote, "Success")
	} else {
		sendToDropletCtl(conn, remote, "Fail")
	}
}

func (s *ProfilerServer) recvProfilerOff(conn *net.UDPConn, remote *net.UDPAddr, args *bytes.Buffer) {
	s.Stop()

	if !s.isRunning() {
		sendToDropletCtl(conn, remote, "Success")
	} else {
		sendToDropletCtl(conn, remote, "Fail")
	}
}

func (s *ProfilerServer) recvProfilerRunningStatus(conn *net.UDPConn, remote *net.UDPAddr, args *bytes.Buffer) {
	if s.isRunning() {
		sendToDropletCtl(conn, remote, "Running")
	} else {
		sendToDropletCtl(conn, remote, "Stopped")
	}
}

func (s *ProfilerServer) start() {
	s.running = true
	s.Server = http.Server{Addr: "0.0.0.0:" + strconv.Itoa(s.port)}
	log.Infof("Start profiler on http 0.0.0.0:%d", s.port)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Warning(err)
			}

			s.running = false
		}
	}()
}

func (s *ProfilerServer) Stop() {
	if s.isRunning() {
		if err := s.Shutdown(nil); err != nil {
			log.Infof("Close profiler failed as %v", err)
		} else {
			log.Infof("Close profiler on http 0.0.0.0:%d", s.port)
		}

		s.running = false
	}
}

func (s *ProfilerServer) Start() {
	s.start()
}

func (s *ProfilerServer) CheckAndStart() {
	if !s.isRunning() {
		s.start()
	}
}

func (s *ProfilerServer) isRunning() bool {
	return s.running
}

func sendProfilerOn(args []string) {
	_, messageBuffer, err := debug.SendToServer(dropletctl.DROPLETCTL_CONFIG, CONFIG_CMD_PROFILER_ON, nil)
	if err != nil {
		fmt.Println(err)
	}

	message := strings.TrimSpace(messageBuffer.String())
	fmt.Println(message)
}

func sendProfilerOff(args []string) {
	_, messageBuffer, err := debug.SendToServer(dropletctl.DROPLETCTL_CONFIG, CONFIG_CMD_PROFILER_OFF, nil)
	if err != nil {
		fmt.Println(err)
	}

	message := strings.TrimSpace(messageBuffer.String())
	fmt.Println(message)
}

func sendProfilerStatus(args []string) {
	_, messageBuffer, err := debug.SendToServer(dropletctl.DROPLETCTL_CONFIG, CONFIG_CMD_PROFILER_STATUS, nil)
	if err != nil {
		fmt.Println(err)
	}

	message := strings.TrimSpace(messageBuffer.String())
	fmt.Println(message)
}

func RegisterProfilerCommand() *cobra.Command {
	profiler := &cobra.Command{
		Use:   "profiler",
		Short: "enable/disable droplet profiler option",
	}

	profilerOn := &cobra.Command{
		Use:   "on",
		Short: "enable droplet profiler option",
		Run: func(cmd *cobra.Command, args []string) {
			sendProfilerOn(args)
		},
	}

	profilerOff := &cobra.Command{
		Use:   "off",
		Short: "disable droplet profiler option",
		Run: func(cmd *cobra.Command, args []string) {
			sendProfilerOff(args)
		},
	}

	profilerStatus := &cobra.Command{
		Use:   "status",
		Short: "show droplet profiler status",
		Run: func(cmd *cobra.Command, args []string) {
			sendProfilerStatus(args)
		},
	}

	profiler.AddCommand(profilerOn)
	profiler.AddCommand(profilerOff)
	profiler.AddCommand(profilerStatus)
	return profiler
}
