package healthcheck

import (
	"context"
	"fmt"

	l5dcharts "github.com/linkerd/linkerd2/pkg/charts/linkerd2"
	"github.com/linkerd/linkerd2/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	// LinkerdAddOnChecks adds checks to validate the add-on components
	LinkerdAddOnChecks CategoryID = "linkerd-addons"

	// LinkerdGrafanaAddOnChecks adds checks related to grafana add-on components
	LinkerdGrafanaAddOnChecks CategoryID = "linkerd-grafana"

	// LinkerdTracingAddOnChecks adds checks related to tracing add-on components
	LinkerdTracingAddOnChecks CategoryID = "linkerd-tracing"
)

var (
	// AddOnCategories is the list of add-on category checks
	AddOnCategories = []CategoryID{LinkerdAddOnChecks, LinkerdGrafanaAddOnChecks, LinkerdTracingAddOnChecks}
)

// addOnCategories contain all the checks w.r.t add-ons. It is strongly advised to
// have warning as true, to not make the check fail for add-on failures as most of them are
// not hard requirements unless otherwise.
func (hc *HealthChecker) addOnCategories() []category {
	return []category{
		{
			id: LinkerdAddOnChecks,
			checkers: []checker{
				{
					description: fmt.Sprintf("%s config map exists", k8s.AddOnsConfigMapName),
					warning:     true,
					check: func(context.Context) error {
						return hc.checkIfAddOnsConfigMapExists()
					},
				},
			},
		},
		{
			id: LinkerdGrafanaAddOnChecks,
			checkers: []checker{
				{
					description: "grafana add-on service account exists",
					warning:     true,
					check: func(context.Context) error {
						if grafana, ok := hc.addOns[l5dcharts.GrafanaAddOn]; ok {
							name, ok := grafana.(map[string]interface{})["name"].(string)
							if !ok {
								return fmt.Errorf("couldn't find %s in grafana", "name")
							}
							return hc.checkServiceAccounts([]string{name}, hc.ControlPlaneNamespace, "")
						}
						return &SkipError{Reason: "grafana add-on not enabled"}
					},
				},
				{
					description: "grafana add-on config map exists",
					warning:     true,
					check: func(context.Context) error {
						if grafana, ok := hc.addOns[l5dcharts.GrafanaAddOn]; ok {
							name, ok := grafana.(map[string]interface{})["name"].(string)
							if !ok {
								return fmt.Errorf("couldn't find %s in grafana", "name")
							}
							_, err := hc.kubeAPI.CoreV1().ConfigMaps(hc.ControlPlaneNamespace).Get(fmt.Sprintf("%s-config", name), metav1.GetOptions{})
							if err != nil {
								return err
							}
							return nil
						}
						return &SkipError{Reason: "grafana add-on not enabled"}
					},
				},
				{
					description:         "grafana pod is running",
					warning:             true,
					retryDeadline:       hc.RetryDeadline,
					surfaceErrorOnRetry: true,
					check: func(context.Context) error {
						if _, ok := hc.addOns[l5dcharts.GrafanaAddOn]; ok {
							// populate controlPlanePods to get the latest status, during retries
							var err error
							hc.controlPlanePods, err = hc.kubeAPI.GetPodsByNamespace(hc.ControlPlaneNamespace)
							if err != nil {
								return err
							}

							return checkContainerRunning(hc.controlPlanePods, "grafana")
						}
						return &SkipError{Reason: "grafana add-on not enabled"}
					},
				},
			},
		},
		{
			id: LinkerdTracingAddOnChecks,
			checkers: []checker{
				{
					description: "collector service account exists",
					warning:     true,
					check: func(context.Context) error {
						if tracing, ok := hc.addOns[l5dcharts.TracingAddOn]; ok {
							collectorName, ok := tracing.(map[string]interface{})["collector"].(map[string]interface{})["name"].(string)
							if !ok {
								return fmt.Errorf("couldn't find %s in tracing", "collector.name")
							}
							return hc.checkServiceAccounts([]string{collectorName}, hc.ControlPlaneNamespace, "")
						}
						return &SkipError{Reason: "tracing add-on not enabled"}
					},
				},
				{
					description: "jaeger service account exists",
					warning:     true,
					check: func(context.Context) error {
						if tracing, ok := hc.addOns[l5dcharts.TracingAddOn]; ok {
							jaegerName, ok := tracing.(map[string]interface{})["jaeger"].(map[string]interface{})["name"].(string)
							if !ok {
								return fmt.Errorf("couldn't find %s in tracing", "jaeger.name")
							}
							return hc.checkServiceAccounts([]string{jaegerName}, hc.ControlPlaneNamespace, "")
						}
						return &SkipError{Reason: "tracing add-on not enabled"}
					},
				},
				{
					description: "collector config map exists",
					warning:     true,
					check: func(context.Context) error {
						if tracing, ok := hc.addOns[l5dcharts.TracingAddOn]; ok {
							collectorName, ok := tracing.(map[string]interface{})["collector"].(map[string]interface{})["name"].(string)
							if !ok {
								return fmt.Errorf("couldn't find %s in tracing", "collector.name")
							}
							_, err := hc.kubeAPI.CoreV1().ConfigMaps(hc.ControlPlaneNamespace).Get(fmt.Sprintf("%s-config", collectorName), metav1.GetOptions{})
							if err != nil {
								return err
							}
							return nil
						}
						return &SkipError{Reason: "tracing add-on not enabled"}
					},
				},
				{
					description:         "collector pod is running",
					warning:             true,
					retryDeadline:       hc.RetryDeadline,
					surfaceErrorOnRetry: true,
					check: func(context.Context) error {
						if _, ok := hc.addOns[l5dcharts.TracingAddOn]; ok {
							// populate controlPlanePods to get the latest status, during retries
							var err error
							hc.controlPlanePods, err = hc.kubeAPI.GetPodsByNamespace(hc.ControlPlaneNamespace)
							if err != nil {
								return err
							}

							return checkContainerRunning(hc.controlPlanePods, "collector")
						}
						return &SkipError{Reason: "tracing add-on not enabled"}
					},
				},
				{
					description:         "jaeger pod is running",
					warning:             true,
					retryDeadline:       hc.RetryDeadline,
					surfaceErrorOnRetry: true,
					check: func(context.Context) error {
						if _, ok := hc.addOns[l5dcharts.TracingAddOn]; ok {
							// populate controlPlanePods to get the latest status, during retries
							var err error
							hc.controlPlanePods, err = hc.kubeAPI.GetPodsByNamespace(hc.ControlPlaneNamespace)
							if err != nil {
								return err
							}

							return checkContainerRunning(hc.controlPlanePods, "jaeger")
						}
						return &SkipError{Reason: "tracing add-on not enabled"}
					},
				},
			},
		},
	}
}

func (hc *HealthChecker) checkIfAddOnsConfigMapExists() error {

	// Check if linkerd-config-addons ConfigMap present, If not skip the next checks
	cm, err := hc.checkForAddOnCM()
	if err != nil {
		return err
	}

	// linkerd-config-addons cm is present,now update hc to include those add-ons
	// so that further add-on specific checks can be ran
	var values l5dcharts.Values
	err = yaml.Unmarshal([]byte(cm), &values)
	if err != nil {
		return fmt.Errorf("could not unmarshal %s config-map: %s", k8s.AddOnsConfigMapName, err)
	}

	addOns, err := l5dcharts.ParseAddOnValues(&values)
	if err != nil {
		return fmt.Errorf("could not read %s config-map: %s", k8s.AddOnsConfigMapName, err)
	}

	hc.addOns = make(map[string]interface{})

	for _, addOn := range addOns {
		values := map[string]interface{}{}
		err = yaml.Unmarshal(addOn.Values(), &values)
		if err != nil {
			return err
		}

		hc.addOns[addOn.Name()] = values
	}

	return nil
}

func (hc *HealthChecker) checkForAddOnCM() (string, error) {
	cm, err := k8s.GetAddOnsConfigMap(hc.kubeAPI, hc.ControlPlaneNamespace)
	if err != nil {
		return "", err
	}

	values, ok := cm["values"]
	if !ok {
		return "", fmt.Errorf("values subpath not found in %s configmap", k8s.AddOnsConfigMapName)
	}

	return values, nil
}
