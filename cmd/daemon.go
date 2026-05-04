package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	daemonPort    int
	daemonWorkDir string
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage AI-Git background service",
	Long: `Manage the AI-Git background service for better performance.

The background service keeps the index in memory, providing faster queries
without the need to reinitialize for each command.`,
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background service",
	RunE:  runDaemonStart,
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background service",
	RunE:  runDaemonStop,
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check background service status",
	RunE:  runDaemonStatus,
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the background service",
	RunE:  runDaemonRestart,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRestartCmd)

	daemonStartCmd.Flags().IntVarP(&daemonPort, "port", "p", 8080, "Service port")
	daemonStartCmd.Flags().StringVarP(&daemonWorkDir, "work-dir", "w", "", "Working directory")
}

type DaemonInfo struct {
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	WorkDir    string    `json:"work_dir"`
	StartTime  time.Time `json:"start_time"`
	Status     string    `json:"status"`
	IndexCount int       `json:"index_count"`
}

func getDaemonInfoPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-git", "daemon.json")
}

func getPIDFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-git", "daemon.pid")
}

func isDaemonRunning() bool {
	pidFile := getPIDFilePath()
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return false
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func runDaemonStart(cmd *cobra.Command, args []string) error {
	if isDaemonRunning() {
		return fmt.Errorf("daemon is already running")
	}

	if daemonWorkDir == "" {
		cwd, _ := os.Getwd()
		daemonWorkDir = cwd
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	daemonArgs := []string{
		"web",
		"--port", strconv.Itoa(daemonPort),
		"--daemon",
	}

	daemonCmd := exec.Command(executable, daemonArgs...)
	daemonCmd.Dir = daemonWorkDir
	daemonCmd.Env = append(os.Environ(), "AI_GIT_DAEMON=1")

	if err := daemonCmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pidFile := getPIDFilePath()
	if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(daemonCmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	info := DaemonInfo{
		PID:       daemonCmd.Process.Pid,
		Port:      daemonPort,
		WorkDir:   daemonWorkDir,
		StartTime: time.Now(),
		Status:    "running",
	}

	infoData, _ := json.Marshal(info)
	infoPath := getDaemonInfoPath()
	if err := ioutil.WriteFile(infoPath, infoData, 0644); err != nil {
		return fmt.Errorf("failed to write daemon info: %w", err)
	}

	time.Sleep(1 * time.Second)

	if !isDaemonRunning() {
		return fmt.Errorf("daemon failed to start")
	}

	fmt.Printf("✓ Daemon started successfully\n")
	fmt.Printf("  PID: %d\n", daemonCmd.Process.Pid)
	fmt.Printf("  Port: %d\n", daemonPort)
	fmt.Printf("  Work Dir: %s\n", daemonWorkDir)
	fmt.Printf("  API: http://localhost:%d/api/v1/\n", daemonPort)

	return nil
}

func runDaemonStop(cmd *cobra.Command, args []string) error {
	if !isDaemonRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	pidFile := getPIDFilePath()
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	os.Remove(pidFile)
	os.Remove(getDaemonInfoPath())

	fmt.Println("✓ Daemon stopped successfully")
	return nil
}

func runDaemonStatus(cmd *cobra.Command, args []string) error {
	if !isDaemonRunning() {
		fmt.Println("Daemon status: not running")
		return nil
	}

	infoPath := getDaemonInfoPath()
	data, err := ioutil.ReadFile(infoPath)
	if err != nil {
		fmt.Println("Daemon status: running (no info available)")
		return nil
	}

	var info DaemonInfo
	if err := json.Unmarshal(data, &info); err != nil {
		fmt.Println("Daemon status: running (invalid info)")
		return nil
	}

	uptime := time.Since(info.StartTime)

	fmt.Println("Daemon status: running")
	fmt.Printf("  PID: %d\n", info.PID)
	fmt.Printf("  Port: %d\n", info.Port)
	fmt.Printf("  Work Dir: %s\n", info.WorkDir)
	fmt.Printf("  Uptime: %v\n", uptime.Round(time.Second))
	fmt.Printf("  API: http://localhost:%d/api/v1/\n", info.Port)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/v1/status", info.Port))
	if err == nil {
		defer resp.Body.Close()
		var status map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&status)
		if stats, ok := status["stats"].(map[string]interface{}); ok {
			if count, ok := stats["index_count"].(float64); ok {
				fmt.Printf("  Index Count: %d\n", int(count))
			}
		}
	}

	return nil
}

func runDaemonRestart(cmd *cobra.Command, args []string) error {
	if isDaemonRunning() {
		if err := runDaemonStop(cmd, args); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return runDaemonStart(cmd, args)
}
