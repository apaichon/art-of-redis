package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ", os.Args[0], " {start|stop} {session-management|leaderboard|sales-analytics|luckydraw}")
	}

	action := os.Args[1]
	appName := os.Args[2]

	switch action {
	case "start":
		startApp(appName)
	case "stop":
		stopApp(appName)
	default:
		log.Fatal("Invalid action. Please use 'start' or 'stop'.")
	}
}

func startApp(appName string) {
	switch appName {
	case "session-management":
		startSessionManagement()
	case "leaderboard":
		startLeaderboard()
	case "sales-analytics":
		startSalesAnalytics()
	case "luckydraw":
		startLuckyDraw()
	default:
		log.Fatal("Invalid app name. Please use 'session-management', 'leaderboard', 'sales-analytics', or 'luckydraw'.")
	}
}

func stopApp(appName string) {
	switch appName {
	case "session-management":
		stopSessionManagement()
	case "leaderboard":
		stopLeaderboard()
	case "sales-analytics":
		stopSalesAnalytics()
	case "luckydraw":
		stopLuckyDraw()
	default:
		log.Fatal("Invalid app name. Please use 'session-management', 'leaderboard', 'sales-analytics', or 'luckydraw'.")
	}
}

func startSessionManagement() {
	cmd := "cd ../session-management/cmd/server && go run main.go --port 9001 &"
	runCommand(cmd)
}

func startLeaderboard() {
	cmd := "cd ../leaderboard/cmd/server && go run main.go --port 9002 &"
	runCommand(cmd)
}

func startSalesAnalytics() {
	cmd := "cd ../sales-analytics/cmd/server && go run main.go --port 9003 &"
	runCommand(cmd)
}

func startLuckyDraw() {
	cmd := "cd ../lucky-draw/cmd/server && go run main.go --port 9004 &"
	runCommand(cmd)
}

func stopSessionManagement() {
	cmd := "pkill -f \"session-management.*9001\""
	runCommand(cmd)
}

func stopLeaderboard() {
	cmd := "pkill -f \"leaderboard.*9002\""
	runCommand(cmd)
}

func stopSalesAnalytics() {
	cmd := "pkill -f \"sales-analytics.*9003\""
	runCommand(cmd)
}

func stopLuckyDraw() {
	cmd := "pkill -f \"lucky-draw.*9004\""
	runCommand(cmd)
}

func runCommand(cmd string) {
	fmt.Println("Running command:", cmd)
	output, err := strconv.Unquote(strconv.Quote(cmd))
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("PATH", "/bin:/usr/bin:/usr/local/bin")
	if err != nil {
		log.Fatal(err)
	}
	cmdParts := []string{"/bin/sh", "-c", output}
	_, err = os.StartProcess(cmdParts[0], cmdParts, &os.ProcAttr{
		Dir:   "",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Sys:   nil,
	})
	if err != nil {
		log.Fatal(err)
	}
}

