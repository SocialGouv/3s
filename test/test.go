package main

import (
	"fmt"
	"sort"
)

func main() {
	svc := []string{}
	svc = append(svc, "svc-2")
	svc = append(svc, "svc-1")
	svc = append(svc, "svc-0")
	sort.Slice(svc, func(i, j int) bool {
		return svc[i] < svc[j]
	})
	podCount := 3
	for i, s := range svc {
		targetPodIndex := i % podCount
		targetPodName := fmt.Sprintf("test-%d", targetPodIndex)
		fmt.Printf("%s %s\n", s, targetPodName)
	}
}
