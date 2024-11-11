package main

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

var ErrInvalidArg = fmt.Errorf("invalid argument")

func lookup[T any](arr []T, idx int, fallback T) T { //nolint:ireturn
	if idx < 0 || idx >= len(arr) {
		return fallback
	}

	return arr[idx]
}

func parseKVPairs(labels []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, label := range labels {
		parts := strings.SplitN(label, "=", 2) //nolint:mnd
		if len(parts) != 2 {                   //nolint:mnd
			return nil, fmt.Errorf("%w: invalid kv format: %s", ErrInvalidArg, label)
		}

		result[parts[0]] = parts[1]
	}

	return result, nil
}

func parseTolerations(tolerations []string) ([]v1.Toleration, error) {
	result := make([]v1.Toleration, 0)

	for _, toleration := range tolerations {
		parts := strings.SplitN(toleration, ":", 4) //nolint:mnd
		if len(parts) < 1 {
			return nil, fmt.Errorf("%w: invalid toleration format: '%s', expected 'key:operator:value:effect'",
				ErrInvalidArg, toleration)
		}

		toleration := v1.Toleration{
			Key:      lookup(parts, 0, ""),
			Operator: v1.TolerationOperator(lookup(parts, 1, "")),
			Value:    lookup(parts, 2, ""),                 //nolint:mnd
			Effect:   v1.TaintEffect(lookup(parts, 3, "")), //nolint:mnd
		}
		result = append(result, toleration)
	}

	return result, nil
}
