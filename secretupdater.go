package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func secretUpdated(clientset *kubernetes.Clientset, password string) {

	// pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	log.Println("Error in Fetching Pods", err.Error())
	// }
	// log.Println("List of pod name")
	// for _, pod := range pods.Items {
	// 	log.Println(pod.Name)
	// }

	// labelSelector := "app=my-app,environment=ci"
	// secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{
	// 	LabelSelector: labelSelector,
	// })

	currentSecret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Println("Error in fetching secrets", err.Error())
	}

	if !bytes.Equal(currentSecret.Data["password"], []byte(password)) {

		log.Println("current password", string(currentSecret.Data["password"]))
		log.Println("New Password", string([]byte(password)))
		currentSecret.Data["password"] = []byte(password)

		updatedSecret, err := clientset.CoreV1().Secrets(namespace).Update(context.Background(), currentSecret, metav1.UpdateOptions{})

		if err != nil {
			log.Println("Error in updating secret", err.Error())
		}
		log.Println("List of secret name")
		log.Printf("Secret %s/%s has been patched.\n", updatedSecret.Namespace, updatedSecret.Name)
		log.Println("Triggering Rolling Restart of Deployment ", deploymentName)

		/// Triggering Deployment / STS

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			deployment, getErr := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
			if getErr != nil {
				return fmt.Errorf("failed to retrieve deployment: %v", getErr)
			}

			// Modify the template spec (for example, change an annotation to trigger the update).
			annotationKey := "trigger-rollout"
			annotationValue := time.Now().String()
			// Check if the annotation exists
			existingAnnotationValue, exists := deployment.Spec.Template.ObjectMeta.Annotations[annotationKey]

			if !exists {
				// Annotation doesn't exist, so add it
				if deployment.Spec.Template.ObjectMeta.Annotations == nil {
					deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
				}
				deployment.Spec.Template.ObjectMeta.Annotations[annotationKey] = annotationValue
			} else if existingAnnotationValue != annotationValue {
				// Annotation exists but with a different value, so modify it
				deployment.Spec.Template.ObjectMeta.Annotations[annotationKey] = annotationValue
			}

			// deployment.Spec.Template.ObjectMeta.Annotations["trigger-rollout"] = time.Now().String()

			_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return updateErr
		})

		if retryErr != nil {
			log.Printf("Error updating Deployment: %v", retryErr)
			os.Exit(1)
		}
		log.Println("Rollout restart triggered for Deployment.")
	}
}
