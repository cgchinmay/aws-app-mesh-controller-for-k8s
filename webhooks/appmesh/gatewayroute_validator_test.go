package appmesh

import (
	"testing"

	appmesh "github.com/aws/aws-app-mesh-controller-for-k8s/apis/appmesh/v1beta2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_gatewayRouteValidator_checkIfHostnameOrPrefixFieldExists(t *testing.T) {
	tests := []struct {
		name    string
		currGR  *appmesh.GatewayRoute
		wantErr error
	}{
		{
			name: "GRPCGateway Route SevericName specified",
			currGR: &appmesh.GatewayRoute{
				Spec: appmesh.GatewayRouteSpec{
					GRPCRoute: &appmesh.GRPCGatewayRoute{
						Match: appmesh.GRPCGatewayRouteMatch{
							ServiceName: aws.String("my-service"),
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "GRPCGateway Route with Valid Hostname specified",
			currGR: &appmesh.GatewayRoute{
				Spec: appmesh.GatewayRouteSpec{
					GRPCRoute: &appmesh.GRPCGatewayRoute{
						Match: appmesh.GRPCGatewayRouteMatch{
							Hostname: appmesh.Hostname{
								Exact: aws.String("example.com"),
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "GRPCGateway Route with Missing Servicename and Hostname",
			currGR: &appmesh.GatewayRoute{
				Spec: appmesh.GatewayRouteSpec{
					GRPCRoute: &appmesh.GRPCGatewayRoute{
						Match: appmesh.GRPCGatewayRouteMatch{},
					},
				},
			},
			wantErr: errors.New("Either servicename or hostname must be specified"),
		},
		{
			name: "HTTPGateway Route with Missing Prefix and Hostname",
			currGR: &appmesh.GatewayRoute{
				Spec: appmesh.GatewayRouteSpec{
					HTTPRoute: &appmesh.HTTPGatewayRoute{
						Match: appmesh.HTTPGatewayRouteMatch{},
					},
				},
			},
			wantErr: errors.New("Either prefix or hostname must be specified"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &gatewayRouteValidator{}
			err := v.checkIfHostnameOrPrefixFieldExists(tt.currGR)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_gatewayRouteValidator_enforceFieldsImmutability(t *testing.T) {
	type args struct {
		newGR *appmesh.GatewayRoute
		oldGR *appmesh.GatewayRoute
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "GatewayRoute immutable fields didn't change",
			args: args{
				newGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
				oldGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "GatewayRoute field awsName changed",
			args: args{
				newGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns_my-cluster"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
				oldGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
			},
			wantErr: errors.New("GatewayRoute update may not change these fields: spec.awsName"),
		},
		{
			name: "GatewayRoute field meshRef changed",
			args: args{
				newGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "another-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
				oldGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
			},
			wantErr: errors.New("GatewayRoute update may not change these fields: spec.meshRef"),
		},
		{
			name: "GatewayRoute field virtualGatewayRef changed",
			args: args{
				newGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "another-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
				oldGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
			},
			wantErr: errors.New("GatewayRoute update may not change these fields: spec.virtualGatewayRef"),
		},
		{
			name: "GatewayRoute fields awsName, meshRef and virtualGatewayRef changed",
			args: args{
				newGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns-my-cluster"),
						MeshRef: &appmesh.MeshReference{
							Name: "another-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "another-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
				oldGR: &appmesh.GatewayRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "awesome-ns",
						Name:      "my-gr",
					},
					Spec: appmesh.GatewayRouteSpec{
						AWSName: aws.String("my-gr_awesome-ns"),
						MeshRef: &appmesh.MeshReference{
							Name: "my-mesh",
							UID:  "408d3036-7dec-11ea-b156-0e30aabe1ca8",
						},
						VirtualGatewayRef: &appmesh.VirtualGatewayReference{
							Name:      "my-vg",
							Namespace: aws.String("gateway-ns"),
							UID:       "346d3036-7dec-11ea-b678-0e30aabe1dg2",
						},
					},
				},
			},
			wantErr: errors.New("GatewayRoute update may not change these fields: spec.awsName,spec.meshRef,spec.virtualGatewayRef"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &gatewayRouteValidator{}
			err := v.enforceFieldsImmutability(tt.args.newGR, tt.args.oldGR)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
