package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/service"
	ps "github.com/mitchellh/go-ps"
)

const serviceName = "RDP"
const serviceDescription = "Webrtc based remote desktop program"

type program struct{}

func (p program) Start(s service.Service) error {
	fmt.Println("Service started")
	go p.run()
	return nil
}

func (p program) Stop(s service.Service) error {
	fmt.Println(s.String() + " stopped")
	return nil
}

func (p program) run() {

	for {

		time.Sleep(1 * time.Second)

		processList, err := ps.Processes()
		if err != nil {
			log.Println("ps.Processes() Failed")
			return
		}

		isFound := false
		for x := range processList {
			process := processList[x]

			onlyName := strings.TrimSuffix(process.Executable(), filepath.Ext(process.Executable()))

			if onlyName == "agent" {
				isFound = true
			}
		}

		if !isFound {
			agentCmd := exec.Command(path.Join(".", "agent.exe"))

			if err != nil {
				log.Println(err)
			} else {
				agentpOut, _ := agentCmd.StdoutPipe()
				agentCmd.Start()
				outBytes, _ := io.ReadAll(agentpOut)
				agentCmd.Wait()
				fmt.Println(string(outBytes))
				time.Sleep(3 * time.Second)
			}
		}

	}
}

func main() {
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		fmt.Println("Cannot create the service: " + err.Error())
	}
	err = s.Run()
	if err != nil {
		fmt.Println("Cannot start the service: " + err.Error())
	}
}
