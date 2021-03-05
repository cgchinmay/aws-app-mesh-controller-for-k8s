package appmesh

import (
	"context"
	"reflect"
	"strings"

	appmesh "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/aws/aws-app-mesh-controller-for-k8s/pkg/webhook"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const apiPathValidateAppMeshGatewayRoute = "/validate-appmesh-k8s-aws-v1beta2-gatewayroute"

// NewGatewayRouteValidator returns a validator for GatewayRoute.
func NewGatewayRouteValidator() *gatewayRouteValidator {
	return &gatewayRouteValidator{}
}

var _ webhook.Validator = &gatewayRouteValidator{}

type gatewayRouteValidator struct {
}

func (v *gatewayRouteValidator) Prototype(req admission.Request) (runtime.Object, error) {
	return &appmesh.GatewayRoute{}, nil
}

func (v *gatewayRouteValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	currGR := obj.(*appmesh.GatewayRoute)
	if err := v.checkIfHostnameOrPrefixFieldExists(currGR); err != nil {
		return err
	}
	return nil
}

func (v *gatewayRouteValidator) ValidateUpdate(ctx context.Context, obj runtime.Object, oldObj runtime.Object) error {
	newGR := obj.(*appmesh.GatewayRoute)
	oldGR := oldObj.(*appmesh.GatewayRoute)
	if err := v.enforceFieldsImmutability(newGR, oldGR); err != nil {
		return err
	}
	return nil
}

func (v *gatewayRouteValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (v *gatewayRouteValidator) checkIfHostnameOrPrefixFieldExists(currGR *appmesh.GatewayRoute) error {
	spec := currGR.Spec
	if spec.GRPCRoute == nil && spec.HTTP2Route == nil && spec.HTTPRoute == nil {
		return errors.Errorf("No matching route specified")
	}
	if spec.GRPCRoute != nil {
		servicename := spec.GRPCRoute.Match.ServiceName
		hostname := spec.GRPCRoute.Match.Hostname
		if servicename == nil && hostname == (appmesh.Hostname{}) {
			return errors.Errorf("Either servicename or hostname must be specified")
		}
	}
	if spec.HTTP2Route != nil {
		prefix := spec.HTTP2Route.Match.Prefix
		hostname := spec.HTTP2Route.Match.Hostname
		return validatePrefixAndHostName(prefix, hostname)
	}
	prefix := spec.HTTPRoute.Match.Prefix
	hostname := spec.HTTPRoute.Match.Hostname
	return validatePrefixAndHostName(prefix, hostname)
}

func validatePrefixAndHostName(prefix *string, hostname appmesh.Hostname) error {
	if prefix == nil && hostname == (appmesh.Hostname{}) {
		return errors.Errorf("Either prefix or hostname must be specified")
	}
	// Validate Hostname
	if prefix == nil {
		exact := hostname.Exact
		suffix := hostname.Suffix
		if exact == nil && suffix == nil {
			return errors.Errorf("Either exact or suffix match for hostname must be specified")
		}
		if exact != nil && suffix != nil {
			return errors.Errorf("Both exact and suffix match for hostname are not allowed. Only one must be specified")
		}
	}
	// TODO: Validation checks for prefix
	return nil
}

// enforceFieldsImmutability will enforce immutable fields are not changed.
func (v *gatewayRouteValidator) enforceFieldsImmutability(newGR *appmesh.GatewayRoute, oldGR *appmesh.GatewayRoute) error {
	var changedImmutableFields []string
	if !reflect.DeepEqual(newGR.Spec.AWSName, oldGR.Spec.AWSName) {
		changedImmutableFields = append(changedImmutableFields, "spec.awsName")
	}
	if !reflect.DeepEqual(newGR.Spec.MeshRef, oldGR.Spec.MeshRef) {
		changedImmutableFields = append(changedImmutableFields, "spec.meshRef")
	}
	if !reflect.DeepEqual(newGR.Spec.VirtualGatewayRef, oldGR.Spec.VirtualGatewayRef) {
		changedImmutableFields = append(changedImmutableFields, "spec.virtualGatewayRef")
	}
	if len(changedImmutableFields) != 0 {
		return errors.Errorf("%s update may not change these fields: %s", "GatewayRoute", strings.Join(changedImmutableFields, ","))
	}
	return nil
}

// +kubebuilder:webhook:path=/validate-appmesh-k8s-aws-v1beta2-gatewayroute,mutating=false,failurePolicy=fail,groups=appmesh.k8s.aws,resources=gatewayroutes,verbs=create;update,versions=v1beta2,name=vgatewayroute.appmesh.k8s.aws,sideEffects=None,webhookVersions=v1beta1

func (v *gatewayRouteValidator) SetupWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(apiPathValidateAppMeshGatewayRoute, webhook.ValidatingWebhookForValidator(v))
}
