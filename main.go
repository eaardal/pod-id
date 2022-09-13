package main

import (
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	ctx := context.Background()

	appName := readAppNameArg()
	podNumber := readPodNumberArg()
	shouldCopy := readCopyArg()

	var namespace string
	namespace, appName = findNamespace(appName)

	home := homedir.HomeDir()
	kubeConfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	var matchingPods []v1.Pod
	for _, item := range pods.Items {
		if strings.Contains(item.Name, appName) {
			matchingPods = append(matchingPods, item)
		}
	}

	if len(matchingPods) == 0 {
		fmt.Printf("Found no pods containing app name %s\n", appName)
		return
	}

	if len(matchingPods) < podNumber {
		fmt.Printf("Out of range: Found %d pods containing app name %s but you asked for pod number %d", len(matchingPods), appName, podNumber)
		return
	}

	printStdout(matchingPods[podNumber-1].Name)

	if shouldCopy {
		if err := clipboard.WriteAll(matchingPods[podNumber-1].Name); err != nil {
			log.Fatal(err)
		}
	}
}

func readAppNameArg() string {
	if len(os.Args) < 2 {
		log.Fatal("App name is expected to be first parameter but no params were given")
	}
	return os.Args[1]
}

func readPodNumberArg() int {
	for _, arg := range os.Args {
		num, err := strconv.Atoi(arg)
		if err != nil {
			continue
		}
		return num
	}
	return 1
}

func readCopyArg() bool {
	for _, arg := range os.Args {
		if arg == "copy" {
			return true
		}
	}
	return false
}

func findNamespace(appName string) (namespace string, cleanAppName string) {
	// First check if namespace is part of appName
	if strings.Contains(appName, "/") {
		parts := strings.Split(appName, "/")
		if len(parts) == 2 {
			namespace = parts[0]
			cleanAppName = parts[1]
			return
		} else {
			log.Fatalf("Unexpected appName format for %s. Expected either '<namespace>/<appName>' or '<appName>'", appName)
		}
	}

	// If namespace is not part of appName, we assume appName is just the clean app name
	cleanAppName = appName

	// Then check if namespace is set in environment
	env, found := os.LookupEnv("PODID_NAMESPACE")
	if found {
		namespace = env
		return
	}

	// Then check the current namespace set by kubens
	kubensNamespace, err := exec.Command("kubens", "--current").Output()
	if err != nil {
		log.Fatal(err)
	}

	namespace = strings.TrimSuffix(string(kubensNamespace), "\n")
	return
}

func printStdout(a ...interface{}) {
	_, _ = fmt.Fprintln(os.Stdout, a...)
}
