# krun - Kubernetes Pod Runner
A command-line tool to easily run one-off commands in Kubernetes pods interactively. It handles pod creation, command execution, and cleanup automatically while providing fine-grained control over pod configuration. 

Think of it as `kubectl run` on steroids

```
Usage: krun --image IMAGE [--namespace NAMESPACE]  [--service-account SERVICE-ACCOUNT]
[--timeout TIMEOUT]  [--verbose] [--labels KEY=VALUE] [--toleration KEY:VALUE:OPERATOR:EFFECT]
[--node-selector KEY=VALUE]  [--requests-cpu REQUESTS-CPU] [--requests-memory REQUESTS-MEMORY]
[--limits-cpu LIMITS-CPU] [--limits-memory LIMITS-MEMORY] [--keeppod] [--env KEY=VALUE]
[COMMAND [ARGS [ARGS ...]]]

Positional arguments:
  COMMAND                Command to run in the container
  ARGS                   Arguments to pass to the command

Options:
  --image IMAGE, -i IMAGE
                         Container image to use
  --namespace NAMESPACE, -n NAMESPACE
                         Kubernetes namespace to launch pod in
  --service-account SERVICE-ACCOUNT, -s SERVICE-ACCOUNT
                         Service account to use
  --timeout TIMEOUT, -t TIMEOUT
                         Timeout to wait for pod to start [default: 300]
  --verbose, -v          Verbose output
  --labels KEY=VALUE, -l KEY=VALUE
                         Labels to add to the pod
  --toleration KEY:VALUE:OPERATOR:EFFECT, -T KEY:VALUE:OPERATOR:EFFECT
                         Tolerations to add to the pod
  --node-selector KEY=VALUE, -N KEY=VALUE
                         Node selector to add to the pod
  --requests-cpu REQUESTS-CPU, -c REQUESTS-CPU
                         CPU request for the container
  --requests-memory REQUESTS-MEMORY, -m REQUESTS-MEMORY
                         Memory request for the container
  --limits-cpu LIMITS-CPU, -C LIMITS-CPU
                         CPU limit for the container
  --limits-memory LIMITS-MEMORY, -M LIMITS-MEMORY
                         Memory limit for the container
  --keeppod, -k          Keep the pod after the command completes
  --env KEY=VALUE, -e KEY=VALUE
                         Environment variables to set in 
the container
  --help, -h             display this help and exit
