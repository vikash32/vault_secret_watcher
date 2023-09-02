package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	prevPassword   string
	vault_addr     string = os.Getenv("VAULT_ADDR")
	vault_token    string = os.Getenv("VAULT_TOKEN")
	secret_path    string = os.Getenv("SECRET_PATH")
	namespace      string = os.Getenv("NAMESPACE")
	deploymentName string = os.Getenv("DEPLOYMENT_NAME")
	secretName     string = os.Getenv("SECRET_NAME")
)

func initClusterConnection() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println("Config Error", err.Error())
		log.Println("Seems we are not inside the clusters, trying the kubeconfig")
		kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		kubeconfig := flag.String("kubeconfig", kubeconfigPath, "Kubeconfig file path")
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Println("Config err : ", err.Error())
			return nil, err
		}
	}

	////// Creating Client Set
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error in creating client set", err.Error())
		return nil, err
	}

	return clientset, nil
}

func main() {

	sleepTime := time.Second * 30

	// Initialize Cluster connectivity
	clientset, err := initClusterConnection()
	if err != nil {
		log.Println("Error in Initialization", err.Error())
		panic(err)
	}

	// Calling Vault Watcher Function
	for {
		VaultWatcher(clientset)
		time.Sleep(sleepTime)
	}

}
