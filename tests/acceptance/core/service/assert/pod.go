package assert

import (
	"fmt"
	"strings"

	"github.com/k3s-io/k3s/tests/acceptance/shared"
	. "github.com/onsi/gomega"
)

// PodAssertFunc is a function type used to create pod assertions
type PodAssertFunc func(g Gomega, pod shared.Pod)

var (
	completed = "Completed"
	running   = "Running"
)

// PodAssertRestart custom assertion func that asserts that pods are not restarting with no reason
// controller, scheduler, helm-install pods can be restarted occasionally when cluster started if only once
func PodAssertRestart() PodAssertFunc {
	return func(g Gomega, pod shared.Pod) {
		if strings.Contains(pod.NameSpace, "kube-system") &&
			strings.Contains(pod.Name, "controller") &&
			strings.Contains(pod.Name, "scheduler") {
			g.Expect(pod.Restarts).Should(SatisfyAny(Equal("0"),
				Equal("1")),
				"could be restarted occasionally when cluster started", pod.Name)
		}
	}
}

// PodAssertStatus custom assertion that asserts that pods status is completed or in some cases
// apply pods can have error status
func PodAssertStatus() PodAssertFunc {
	return func(g Gomega, pod shared.Pod) {
		if strings.Contains(pod.Name, "helm-install") {
			g.Expect(pod.Status).Should(Equal(completed), pod.Name)
		} else if strings.Contains(pod.NameSpace, "system-upgrade") {
			g.Expect(pod.Status).Should(SatisfyAny(
				ContainSubstring("Error"),
				Equal(completed),
			), pod.Name)
		} else {
			g.Expect(pod.Status).Should(Equal(running), pod.Name)
		}
	}
}

// PodAssertReady custom assertion func that asserts that the pod is
// with correct numbers of ready containers
func PodAssertReady() PodAssertFunc {
	return func(g Gomega, pod shared.Pod) {
		g.ExpectWithOffset(1, pod.Ready).To(checkReadyFields(),
			"should have equal values in n/n format")
	}
}

// checkReadyFields is a custom matcher that checks
// if the input string is in N/N format and the same qty
func checkReadyFields() OmegaMatcher {
	return WithTransform(func(s string) (bool, error) {
		var a, b int
		n, err := fmt.Sscanf(s, "%d/%d", &a, &b)
		if err != nil || n != 2 {
			return false, fmt.Errorf("failed to parse format: %v", err)
		}
		return a == b, nil
	}, BeTrue())
}

// CheckPodStatusRunning asserts that the pod is running with the specified label = app name.
// don't need to send KubeconfigFile
func CheckPodStatusRunning(name, assert string) {
	cmd := "kubectl get pods -l k8s-app=" + name +
		" --field-selector=status.phase=Running --kubeconfig=" + shared.KubeConfigFile
	Eventually(func(g Gomega) (string, error) {
		return shared.RunCommandHost(cmd)
	}, "180s", "5s").Should(ContainSubstring(assert))
}
