package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/alexflint/go-arg"
)

type KrunOptions struct {
	Image          string   `arg:"-i,required,help:Container image to use"`
	Namespace      string   `arg:"-n,help:Kubernetes namespace to launch pod in"`
	ServiceAccount string   `arg:"-s,--service-account,help:Service account to use"`
	Timeout        int64    `arg:"-t,help:Timeout to wait for pod to start" default:"300"`
	Verbose        bool     `arg:"-v,help:Verbose output"`
	Labels         []string `arg:"-l,help:Labels to add to the pod" placeholder:"KEY=VALUE"`
	Tolerations    []string `arg:"-T,--toleration" help:"Tolerations to add to the pod" placeholder:"KEY:VALUE:OPERATOR:EFFECT"` //nolint:lll
	NodeSelector   []string `arg:"-N,--node-selector" help:"Node selector to add to the pod" placeholder:"KEY=VALUE"`
	RequestsCPU    string   `arg:"-c,--requests-cpu" help:"CPU request for the container"`
	RequestsMem    string   `arg:"-m,--requests-memory" help:"Memory request for the container"`
	LimitsCPU      string   `arg:"-C,--limits-cpu" help:"CPU limit for the container"`
	LimitsMem      string   `arg:"-M,--limits-memory" help:"Memory limit for the container"`
	KeepPod        bool     `arg:"-k,help:Keep the pod after the command completes"`
	Env            []string `arg:"-e,help:Environment variables to set in the container" placeholder:"KEY=VALUE"`
	Command        string   `arg:"positional" default:"/bin/sh" help:"Command to run in the container"`
	Args           []string `arg:"positional" help:"Arguments to pass to the command"`
}

func main() {
	var opts KrunOptions

	arg.MustParse(&opts)

	logger := &ConsoleLogger{verbose: opts.Verbose}

	config, err := NewKubeConfig()
	if err != nil {
		logger.Errorf("Failed to create kube config: %v", err)
		os.Exit(1)
	}

	if err := run(opts, config, logger); err != nil {
		logger.Errorf("Failed to run: %v", err)
		os.Exit(1)
	}
}

func run(opts KrunOptions, config *KubeConfig, logger Logger) error { //nolint:funlen
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager := NewPodManager(config, logger)

	podOpts := createPodOptions{
		image:          opts.Image,
		command:        opts.Command,
		args:           opts.Args,
		serviceAccount: opts.ServiceAccount,
		labels:         opts.Labels,
		tolerations:    opts.Tolerations,
		nodeSelector:   opts.NodeSelector,
		requestsCPU:    opts.RequestsCPU,
		requestsMem:    opts.RequestsMem,
		limitsCPU:      opts.LimitsCPU,
		limitsMem:      opts.LimitsMem,
		activeDeadline: opts.Timeout,
		env:            opts.Env,
	}

	namespace := opts.Namespace
	if namespace == "" {
		namespace = config.Namespace()
	}

	podName, err := manager.CreatePod(ctx, namespace, podOpts)
	if err != nil {
		return fmt.Errorf("failed to create pod: %w", err)
	}

	sigch := make(chan os.Signal, 1)

	go func() {
		signal.Notify(sigch, os.Interrupt)
		select {
		case <-sigch:
			cancel()
		case <-ctx.Done():
		}
	}()

	defer func() {
		signal.Stop(sigch)

		if opts.KeepPod {
			return
		}

		if err := manager.DeletePod(ctx, namespace, podName); err != nil {
			logger.Errorf("Failed to delete pod: %v", err)
		}
	}()

	err = manager.WaitForPodRunningWithTimeout(ctx, namespace, podName)
	if err != nil {
		return fmt.Errorf("failed to wait for pod to start: %w", err)
	}

	if err = manager.AttachToPod(ctx, namespace, podName); err != nil {
		return fmt.Errorf("failed to attach to pod: %w", err)
	}

	return nil
}
