package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/util/term"
)

const (
	PodPrefix        = "krun-"
	PodGetTimeout    = 10 * time.Second
	PodCreateTimeout = 60 * time.Second
	PodDeleteTimeout = 300 * time.Second
)

var ErrInvalidPodStatus = fmt.Errorf("invalid pod status")

type PodManager struct {
	logger Logger
	config *KubeConfig
}

type createPodOptions struct {
	command        string
	args           []string
	image          string
	serviceAccount string
	labels         []string
	tolerations    []string
	nodeSelector   []string
	requestsCPU    string
	requestsMem    string
	limitsCPU      string
	limitsMem      string
	activeDeadline int64
	env            []string
}

func NewPodManager(config *KubeConfig, logger Logger) *PodManager {
	return &PodManager{
		config: config,
		logger: logger,
	}
}

func (pm *PodManager) CreatePod( //nolint:funlen,cyclop
	ctx context.Context,
	namespace string,
	opts createPodOptions,
) (string, error) {
	createCtx, cancel := context.WithTimeout(ctx, PodCreateTimeout)
	defer cancel()

	command := append([]string{opts.command}, opts.args...)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: PodPrefix,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "main",
					Image:   opts.image,
					Command: command,
					Stdin:   true,
					TTY:     true,
				},
			},
			RestartPolicy:         v1.RestartPolicyNever,
			ActiveDeadlineSeconds: &opts.activeDeadline,
		},
	}

	labels, err := parseKVPairs(opts.labels)
	if err != nil {
		return "", fmt.Errorf("failed to parse labels: %w", err)
	}

	if len(labels) > 0 {
		pod.ObjectMeta.Labels = labels
	}

	tolerations, err := parseTolerations(opts.tolerations)
	if err != nil {
		return "", fmt.Errorf("failed to parse tolerations: %w", err)
	}

	if len(tolerations) > 0 {
		pod.Spec.Tolerations = tolerations
	}

	nodeSelector, err := parseKVPairs(opts.nodeSelector)
	if err != nil {
		return "", fmt.Errorf("failed to parse node selector: %w", err)
	}

	if len(nodeSelector) > 0 {
		pod.Spec.NodeSelector = nodeSelector
	}

	if opts.serviceAccount != "" {
		pod.Spec.ServiceAccountName = opts.serviceAccount
	}

	requests := v1.ResourceList{}
	requestsSet := false

	limits := v1.ResourceList{}
	limitsSet := false

	if opts.requestsCPU != "" {
		requests[v1.ResourceCPU] = resource.MustParse(opts.requestsCPU)
		requestsSet = true
	}

	if opts.requestsMem != "" {
		requests[v1.ResourceMemory] = resource.MustParse(opts.requestsMem)
		requestsSet = true
	}

	if opts.limitsCPU != "" {
		limits[v1.ResourceCPU] = resource.MustParse(opts.limitsCPU)
		limitsSet = true
	}

	if opts.limitsMem != "" {
		limits[v1.ResourceMemory] = resource.MustParse(opts.limitsMem)
		limitsSet = true
	}

	if requestsSet || limitsSet {
		resources := v1.ResourceRequirements{}

		if requestsSet {
			resources.Requests = requests
		}

		if limitsSet {
			resources.Limits = limits
		}

		pod.Spec.Containers[0].Resources = resources
	}

	if len(opts.env) > 0 {
		envMap, err := parseKVPairs(opts.env)
		if err != nil {
			return "", fmt.Errorf("failed to parse env vars: %w", err)
		}

		envVars := make([]v1.EnvVar, 0, len(envMap))
		for k, v := range envMap {
			envVars = append(envVars, v1.EnvVar{Name: k, Value: v})
		}

		pod.Spec.Containers[0].Env = envVars
	}

	pm.logger.Debugf("Creating pod in namespace '%s' to run command '%s' with image '%s'",
		namespace, strings.Join(command, " "), opts.image)

	createdPod, err := pm.config.Clientset().CoreV1().Pods(namespace).Create(createCtx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create pod: %w", err)
	}

	return createdPod.Name, nil
}

func (pm *PodManager) WaitForPodRunningWithTimeout(
	ctx context.Context,
	namespace, podName string,
) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Wrap the call in a function to use defer for cleanup
		pod, err := func() (*v1.Pod, error) {
			getCtx, cancel := context.WithTimeout(ctx, PodGetTimeout)
			defer cancel()

			return pm.config.Clientset().CoreV1().Pods(namespace).Get(getCtx, podName, metav1.GetOptions{})
		}()
		if err != nil {
			return fmt.Errorf("failed to get pod status: %w", err)
		}

		pm.logger.Debugf("Pod '%s' status: %s", podName, pod.Status.Phase)

		switch pod.Status.Phase {
		case v1.PodRunning:
			return nil
		case v1.PodPending:
			continue
		case v1.PodFailed, v1.PodSucceeded, v1.PodUnknown:
			return fmt.Errorf("%w: pod entered terminal phase: %s", ErrInvalidPodStatus, pod.Status.Phase)
		}
	}

	return nil
}

func (pm *PodManager) DeletePod(ctx context.Context, namespace, podName string) error {
	deleteCtx, cancel := context.WithTimeout(ctx, PodDeleteTimeout)
	defer cancel()

	pm.logger.Debugf("Deleting pod '%s'", podName)

	err := pm.config.Clientset().CoreV1().Pods(namespace).Delete(deleteCtx, podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	return nil
}

func (pm *PodManager) AttachToPod(ctx context.Context, namespace, podName string) error {
	req := pm.config.Clientset().CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("attach").
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "true")

	exec, err := remotecommand.NewSPDYExecutor(pm.config.RESTConfig(), "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to initialize SPDY executor: %w", err)
	}

	pm.logger.Debugf("Attaching to pod '%s'", podName)

	tty := term.TTY{
		In:  os.Stdin,
		Out: os.Stdout,
		Raw: true,
	}

	pm.logger.Infof("If you don't see a command prompt, try pressing enter.")

	err = tty.Safe(func() error {
		return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:  tty.In,
			Stdout: tty.Out,
			Stderr: tty.Out,
			Tty:    true,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to attach to pod: %w", err)
	}

	return nil
}
