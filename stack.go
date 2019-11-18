package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jinzhu/copier"
	"github.com/v-braun/go-must"
)

type stackRunStatus int

const (
	unknownStatus stackRunStatus = 0
	stoppedStatus stackRunStatus = 1
	warningStatus stackRunStatus = 2
	runningStatus stackRunStatus = 3
)

type stackStatus struct {
	Title         string
	Path          string
	StatusMessage string

	AllServices     []string
	RunningServices []string

	Status stackRunStatus
}

type stack struct {
	status   *stackStatus
	onUpdate func()
	onLog    func(msg string)
	path     string
	title    string

	pollIntervall      time.Duration
	failedCommandSleep time.Duration

	changeFlag bool

	muted bool

	lock sync.Mutex
}

func newStack(title string, path string) *stack {
	result := new(stack)
	result.path = path
	result.title = title
	result.status = new(stackStatus)
	result.status.Title = title
	result.status.Path = path
	result.status.StatusMessage = "run 'docker-compose ps'"
	result.status.Status = unknownStatus
	result.failedCommandSleep = time.Second * 10
	result.pollIntervall = time.Second * 10
	result.changeFlag = false

	result.onUpdate = func() {}
	result.onLog = func(msg string) {}
	result.lock = sync.Mutex{}

	go func() {
		result.monitorLoop()
	}()
	go func() {
		result.logsLoop()
	}()

	return result
}

func (s *stack) updateStatus(handler func(status *stackStatus)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	handler(s.status)

	go s.onUpdate()
}

func (s *stack) GetStatus() *stackStatus {
	s.lock.Lock()
	defer s.lock.Unlock()
	result := new(stackStatus)
	err := copier.Copy(&result, &s.status)
	must.NoError(err, "unexpected error during copy")

	return result
}

func (s *stack) toggleMute() {
	// TODO
	s.muted = !s.muted
}

func (s *stack) toggleRunning() {
	stat := s.GetStatus()
	s.changeFlag = true
	s.updateStatus(func(stat *stackStatus) {
		stat.Status = warningStatus
	})

	go func() {
		// set it again if another routine updated it
		s.changeFlag = true
		s.updateStatus(func(stat *stackStatus) {
			stat.Status = warningStatus
		})

		if stat.Status == runningStatus || stat.Status == warningStatus {
			logger.Printf("%s: begin down", s.title)
			s.execComposeCommand("down").Wait()
			logger.Printf("%s: end down", s.title)

			s.updateStatus(func(stat *stackStatus) {
				stat.Status = stoppedStatus
			})

		} else {
			logger.Printf("%s: begin up", s.title)
			s.execComposeCommand("up", "-d").Wait()
			logger.Printf("%s: end up", s.title)
			s.updateStatus(func(stat *stackStatus) {
				stat.Status = warningStatus
			})
		}

		s.changeFlag = false
	}()
}

func (s *stack) logsLoop() {
	for {

		cmd := s.execComposeCommand("logs", "-f", "--no-color")

		output, _ := cmd.StdoutPipe()
		reader := bufio.NewReader(output)
		cmd.Start()
		for {
			line, _, err := reader.ReadLine()

			if err != nil {
				if err == io.EOF || errors.Is(err, os.ErrClosed) {
					break
				}

				logger.Printf("err in compose-logs: %v", err)
				break
			}
			if bytes.Equal(line, []byte("Attaching to ")) {
				s.updateStatus(func(stat *stackStatus) {
					if stat.Status == runningStatus {
						stat.Status = warningStatus
					} else {
						stat.Status = stoppedStatus
						stat.StatusMessage = "no running container"
					}
				})
				break
			}

			go s.onLog(string(line))
			logger.Printf("got msg: [%s]", string(line))
		}

		// s.updateStatus(func(stat *stackStatus) {
		// 	stat.Status = warningStatus
		// })

		output.Close()
		endOrKill(cmd)
		time.Sleep(s.failedCommandSleep)
	}
}

func (s *stack) monitorLoop() {
	for {

		running, all, runStat, err := s.getContainerStates()
		if err != nil {
			logger.Printf("err in get container stat: %v", err)
			s.updateStatus(func(stat *stackStatus) {
				stat.Status = runStat
				stat.StatusMessage = fmt.Sprintf("failed list running container: %s", err.Error())
			})
			time.Sleep(s.failedCommandSleep)
			continue
		}

		logger.Printf("got state: %v | running: %v | all: %v", runStat, running, all)
		if !s.changeFlag {
			s.updateStatus(func(stat *stackStatus) {
				stat.Status = runStat
				stat.RunningServices = running
				stat.AllServices = all
				stat.StatusMessage = ""
				if runStat == stoppedStatus {
					stat.StatusMessage = "no running container"
				} else if runStat == warningStatus {
					stat.StatusMessage = "not all container running"
				}
			})
		}

		time.Sleep(s.pollIntervall)
	}
}

func (s *stack) getContainerStates() (running []string, all []string, status stackRunStatus, err error) {
	all, err = s.execComposePS("")
	if err != nil {
		return nil, nil, stoppedStatus, err
	}

	running, err = s.execComposePS("--filter=status=running")
	if err != nil {
		return nil, nil, stoppedStatus, err
	}

	stat := stoppedStatus
	if len(running) == len(all) {
		stat = runningStatus
	} else if len(running) == 0 {
		stat = stoppedStatus
	} else {
		stat = warningStatus
	}

	return running, all, stat, err
}

func (s *stack) execComposePS(filter string) ([]string, error) {
	args := []string{"ps", "--services"}
	if filter != "" {
		args = append(args, filter)
	}

	logger.Printf("%s: begin ps %v", s.title, args)
	out, err := s.execComposeCommand(args...).CombinedOutput()
	logger.Printf("%s: end ps %v | result: %s", s.title, args, string(out))
	if err != nil {
		return []string{}, fmt.Errorf("failed exec compose command: %s, out: %s", err.Error(), string(out))
	}

	result := []string{}
	list := strings.Split(string(out), "\n")
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result, nil
}

func endOrKill(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	cmd.Process.Signal(syscall.SIGINT)

	done := make(chan error, 1)
	go func() {
		cmd.Wait()
		close(done)
	}()
	select {
	case <-time.After(3 * time.Second):
		cmd.Process.Kill()
		break
	case <-done:
		break
	}

	<-done
}

func (s *stack) execComposeCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = s.path
	return cmd
}

func createStacks(conf *conf) []*stack {
	result := []*stack{}
	for _, confS := range conf.Stacks {
		s := newStack(confS.Title, confS.Path)
		result = append(result, s)
	}

	return result
}
