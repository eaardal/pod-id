package main

import (
	"context"
	"fmt"
	"github.com/atotto/clipboard"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	ctx := context.Background()

	appName := readAppNameArg()
	podNumber := readPodNumberArg()
	shouldCopy := readCopyArg()
	selectorMode := readSelectorModeArg()

	var namespace string
	namespace, appName = findNamespace(appName)

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		log.Fatalf("Failed to load Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes clientset: %v", err)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list pods in namespace \"%s\": %v", namespace, err)
	}

	matchingPods := filterPodsByNames(pods.Items, splitAppNames(appName))

	if len(matchingPods) == 0 {
		printStderr("Found no pod name containing \"%s\"", appName)
		return
	}

	if selectorMode {
		selector, err := resolveSelector(matchingPods)
		if err != nil {
			printStderr("%v", err)
			return
		}

		printStdout(selector)

		if shouldCopy {
			if err := clipboard.WriteAll(selector); err != nil {
				log.Fatalf("Failed to write the selector to your clipboard: %v", err)
			}
		}
		return
	}

	if len(matchingPods) < podNumber {
		printStderr("Out of range: Found %d pod name(s) containing \"%s\" but you asked for pod number %d", len(matchingPods), appName, podNumber)
		return
	}

	printStdout(matchingPods[podNumber-1].Name)

	if shouldCopy {
		if err := clipboard.WriteAll(matchingPods[podNumber-1].Name); err != nil {
			log.Fatalf("Failed to write the pod's ID to your clipboard: %v", err)
		}
	}
}

// splitAppNames splits the raw app-name argument into one or more partial
// names. Callers may pass a comma-separated list (e.g. "api-gateway,invoice")
// to match several apps at once; a plain name yields a single-element slice.
// Empty parts and surrounding whitespace are discarded.
func splitAppNames(appName string) []string {
	var names []string
	for _, part := range strings.Split(appName, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			names = append(names, part)
		}
	}
	return names
}

// filterPodsByNames returns every pod whose name contains at least one of the
// given partial names. Each pod appears at most once even when several partial
// names match it, and the original listing order is preserved.
func filterPodsByNames(pods []v1.Pod, names []string) []v1.Pod {
	var matching []v1.Pod
	for _, item := range pods {
		if podNameMatchesAny(item.Name, names) {
			matching = append(matching, item)
		}
	}
	return matching
}

// podNameMatchesAny reports whether the pod name contains any of the partial
// names as a substring.
func podNameMatchesAny(podName string, names []string) bool {
	for _, name := range names {
		if strings.Contains(podName, name) {
			return true
		}
	}
	return false
}

func readAppNameArg() string {
	// The app name is the first positional argument: the first that isn't a flag
	// (e.g. -l), the "copy" keyword, or a pod number.
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") || arg == "copy" {
			continue
		}
		if _, err := strconv.Atoi(arg); err == nil {
			continue
		}
		return arg
	}

	log.Fatal("App name is expected but no app name argument was given")
	return ""
}

func readSelectorModeArg() bool {
	for _, arg := range os.Args {
		if arg == "-l" || arg == "--selector" {
			return true
		}
	}
	return false
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
			log.Fatalf("Unexpected appName format for \"%s\". Expected either '<namespace>/<appName>' or '<appName>'", appName)
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

	if !hasExe("kubens") {
		log.Fatalln("Could not resolve the Kubernetes namespace to use because the first two methods of looking up the namespace didn't give any results and it appears kubens is not available on your PATH. Use one of these three methods to provide your namespace: 1) Specify the namespace along with the name to search for: <namespace>/<appName>. 2) Set the PODID_NAMESPACE environment variable. 3) Use kubens (https://github.com/ahmetb/kubectx) to set the current namespace. This is the recommended method. The namespace is resolved in the order these alternatives are listed.")
	}

	// Then check the current namespace set by kubens
	kubensNamespace, err := exec.Command("kubens", "--current").Output()
	if err != nil {
		log.Fatal(err)
	}

	if len(kubensNamespace) > 0 {
		namespace = strings.TrimSuffix(string(kubensNamespace), "\n")
	} else {
		log.Fatalln("Found no namespace by invoking `kubens --current`. Ensure kubens is set up correctly.")
	}

	if len(namespace) == 0 {
		log.Fatalln("Could not resolve the Kubernetes namespace to use. Use one of the three methods: 1) Specify the namespace along with the name to search for: <namespace>/<appName>. 2) Set the PODID_NAMESPACE environment variable. 3) Use kubens (https://github.com/ahmetb/kubectx) to set the current namespace. This is the recommended method. The namespace is resolved in the order these alternatives are listed.")
	}
	return
}

func hasExe(exe string) bool {
	path, err := exec.LookPath(exe)
	return err == nil && len(path) > 0
}

func printStdout(format string, a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(format, a...))
}

func printStderr(format string, a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}
