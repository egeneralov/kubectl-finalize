package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	namespace = ""
)

// get first argument
// if it's not empty, use it as namespace
func init() {
	if len(os.Args) > 1 {
		namespace = os.Args[1]
	}
	if namespace == "" {
		panic("namespace is empty")
	}
}

func main() {
	go func() {
		out, err := runCommandWithTimeout("kubectl proxy", 10*time.Second)
		if err != nil {
			panic(fmt.Errorf("failed to run kubectl proxy: %w, %v", err, out))
		}
	}()
	time.Sleep(3 * time.Second)
	for _, command := range []string{
		fmt.Sprintf("kubectl get namespace %v -o json |jq '.spec = {\"finalizers\":[]}' > temp.json", namespace),
		fmt.Sprintf("kubectl replace --raw \"/api/v1/namespaces/%v/finalize\" -f temp.json", namespace),
	} {
		_, err := runCommandWithTimeout(command, 3*time.Second)
		if err != nil {
			panic(fmt.Errorf("command %v failed: %v", command, err))
		}
	}
	err := os.Remove("temp.json")
	if err != nil {
		panic(fmt.Errorf("failed to remove temp.json: %v", err))
	}
}

func runCommandWithTimeout(command string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	out, err := cmd.CombinedOutput()

	return string(out), err
}
