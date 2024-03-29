package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getNamespace(defaultNamespace string) string {
	namespacePath := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if data, err := ioutil.ReadFile(namespacePath); err == nil {
		if ns := string(data); ns != "" {
			return ns
		}
	}
	return defaultNamespace
}

func updateServiceSelectors(clientset *kubernetes.Clientset, namespace, podSelector string) {
	serviceSelector := getEnv("SSS_SERVICE_SELECTOR", "")
	servicePrefix := getEnv("SSS_SERVICE_PREFIX", "svc-")
	statefulSetNameFormat := getEnv("SSS_STATEFULSET_NAME_FORMAT", "my-statefulset-%d")

	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: serviceSelector,
	})
	if err != nil {
		fmt.Println("Failed to list services:", err)
		return
	}

	filteredServices := []corev1.Service{}
	for _, service := range services.Items {
		if strings.HasPrefix(service.Name, servicePrefix) {
			filteredServices = append(filteredServices, service)
		}
	}

	sort.Slice(filteredServices, func(i, j int) bool {
		return filteredServices[i].Name < filteredServices[j].Name
	})

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: podSelector,
	})
	if err != nil {
		fmt.Println("Failed to list pods:", err)
		return
	}

	podCount := len(pods.Items)
	fmt.Printf("podCount: %d\n", podCount)
	if podCount == 0 {
		fmt.Println("No pods found, skipping service update")
		return
	}

	for i, service := range filteredServices {
		targetPodIndex := i % podCount
		targetPodName := fmt.Sprintf(statefulSetNameFormat, targetPodIndex)

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			updatingService, getErr := clientset.CoreV1().Services(namespace).Get(context.TODO(), service.Name, metav1.GetOptions{})
			if getErr != nil {
				return getErr
			}

			updatingService.Spec.Selector = map[string]string{
				"statefulset.kubernetes.io/pod-name": targetPodName,
			}

			_, updateErr := clientset.CoreV1().Services(namespace).Update(context.TODO(), updatingService, metav1.UpdateOptions{})
			return updateErr
		})

		if retryErr != nil {
			fmt.Printf("Update failed for service %s: %v\n", service.Name, retryErr)
		} else {
			fmt.Printf("Service %s updated to target pod %s\n", service.Name, targetPodName)
		}
	}
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := getEnv("SSS_NAMESPACE", getNamespace(getEnv("SSS_DEFAULT_NAMESPACE", "default")))
	podSelector := getEnv("SSS_POD_SELECTOR", "app=myapp")

	// Perform initial adjustment of service selectors
	updateServiceSelectors(clientset, namespace, podSelector)

	podListWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.LabelSelector = podSelector
			return clientset.CoreV1().Pods(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = podSelector
			return clientset.CoreV1().Pods(namespace).Watch(context.TODO(), options)
		},
	}

	_, controller := cache.NewInformer(
		podListWatch,
		&corev1.Pod{},
		time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				updateServiceSelectors(clientset, namespace, podSelector)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				updateServiceSelectors(clientset, namespace, podSelector)
			},
			DeleteFunc: func(obj interface{}) {
				updateServiceSelectors(clientset, namespace, podSelector)
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	// Wait forever
	select {}
}
