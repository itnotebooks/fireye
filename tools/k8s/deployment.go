package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Client *kubernetes.Clientset
}

func NewK8SClient() *Client {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return &Client{
		Client: client,
	}
}

func (c *Client) GetDeploymentList(namespace string) (list []string) {
	deploymentList, err := c.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, v := range deploymentList.Items {
		list = append(list, v.Name)
	}
	return list
}

func (c *Client) GetPodList(namespace, app string) (list []string) {
	labelSelect := fmt.Sprintf("app=%s", app)
	podList, err := c.Client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: labelSelect})
	if err != nil {
		panic(err)
	}

	for _, v := range podList.Items {
		list = append(list, v.Name)
	}
	return list
}
