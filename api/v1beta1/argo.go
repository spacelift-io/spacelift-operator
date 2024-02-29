package v1beta1

// ArgoHealth is a string type to represent the argo health of a resource
// More info on the argo doc here https://argo-cd.readthedocs.io/en/stable/operator-manual/health/
type ArgoHealth string

const (
	// ArgoHealthHealthy the resource is healthy
	ArgoHealthHealthy ArgoHealth = "Healthy"
	// ArgoHealthProgressing the resource is not healthy yet but still making progress and might be healthy soon
	ArgoHealthProgressing ArgoHealth = "Progressing"
	// ArgoHealthSuspended the resource is suspended and waiting for some external event to resume (e.g. suspended CronJob or paused Deployment)
	ArgoHealthSuspended ArgoHealth = "Suspended"
	// ArgoHealthDegraded the resource is degraded
	ArgoHealthDegraded ArgoHealth = "Degraded"
)

type ArgoStatus struct {
	Health ArgoHealth `json:"health"`
}
