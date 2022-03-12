// Copyright (c) 2021 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filter

import (
	"context"
	"strings"

	netv1 "k8s.io/api/networking/v1"
	netv1beta1 "k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/bfenetworks/ingress-bfe/internal/bfeConfig/annotations"
	"github.com/bfenetworks/ingress-bfe/internal/option"
)

func IngressClassFilter(ctx context.Context, r client.Reader, annots map[string]string, ingressClassName *string) bool {
	if annots[annotations.IngressClassKey] == option.Opts.Ingress.IngressClass {
		return true
	}

	classListV1 := &netv1.IngressClassList{}
	err := r.List(ctx, classListV1)
	if err == nil {
		for _, class := range classListV1.Items {
			if class.Spec.Controller != option.Opts.Ingress.ControllerName {
				continue
			}
			if (ingressClassName != nil && *ingressClassName == class.Name) ||
				(ingressClassName == nil && strings.EqualFold(class.Annotations[annotations.IsDefaultIngressClass], "true")) {
				return true
			}
		}
	}

	classListV1Beta1 := &netv1beta1.IngressClassList{}
	err = r.List(ctx, classListV1Beta1)
	if err != nil {
		return false
	}
	for _, classV1Beta1 := range classListV1Beta1.Items {
		if classV1Beta1.Spec.Controller != option.Opts.Ingress.ControllerName {
			continue
		}
		if (ingressClassName != nil && *ingressClassName == classV1Beta1.Name) ||
			(ingressClassName == nil && strings.EqualFold(classV1Beta1.Annotations[annotations.IsDefaultIngressClass], "true")) {
			return true
		}
	}

	return false
}
