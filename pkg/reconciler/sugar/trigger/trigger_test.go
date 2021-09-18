/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trigger

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/eventing/pkg/apis/config"
	"knative.dev/eventing/pkg/apis/eventing"
	v1 "knative.dev/eventing/pkg/apis/eventing/v1"
	fakeeventingclient "knative.dev/eventing/pkg/client/injection/client/fake"
	"knative.dev/eventing/pkg/client/injection/reconciler/eventing/v1/trigger"
	"knative.dev/eventing/pkg/reconciler/sugar"
	"knative.dev/eventing/pkg/reconciler/sugar/resources"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	logtesting "knative.dev/pkg/logging/testing"

	. "knative.dev/eventing/pkg/reconciler/testing/v1"
	. "knative.dev/pkg/reconciler/testing"
)

const (
	testNS      = "test-namespace"
	triggerName = "test-trigger"
	brokerName  = "default"
)

func TestEnabledByDefault(t *testing.T) {
	// Events
	brokerEvent := Eventf(corev1.EventTypeNormal, "BrokerCreated", "Default eventing.knative.dev Broker %q created.", "default")

	// Objects
	broker := resources.MakeBroker(testNS, resources.DefaultBrokerName)
	broker.Annotations = map[string]string{eventing.BrokerClassKey: eventing.MTChannelBrokerClassValue}

	configuredCtx, _ := SetupFakeContext(t)
	configuredCtx = config.ToContext(configuredCtx, &config.Config{
		Defaults: &config.Defaults{
			ClusterDefault: &config.ClassAndBrokerConfig{
				BrokerClass: eventing.MTChannelBrokerClassValue,
			},
		},
	})

	table := TableTest{{
		Name: "bad workqueue key",
		// Make sure Reconcile handles bad keys.
		Key: "too/many/parts",
	}, {
		Name: "key not found",
		// Make sure Reconcile handles good keys that don't exist.
		Key: "foo/not-found",
	}, {
		Name: "Trigger is not labeled",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
		WantEvents: []string{
			brokerEvent,
		},
		WantCreates: []runtime.Object{
			broker,
		},
	}, {
		Name: "Trigger is labeled disabled",
		Objects: []runtime.Object{
			NewNamespace(testNS,
				WithNamespaceLabeled(sugar.InjectionDisabledLabels())),
		},
		Ctx: configuredCtx,
		Key: testNS + "/" + triggerName,
	}, {
		Name: "Trigger is deleted no resources",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue),
				WithTriggerDeleted),
		},
		Ctx: configuredCtx,
		Key: testNS + "/" + triggerName,
	}, {
		Name: "Trigger enabled",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue)),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
		WantEvents: []string{
			brokerEvent,
		},
		WantCreates: []runtime.Object{
			broker,
		},
	}, {
		Name: "Trigger enabled, broker exists",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue),
			),
			resources.MakeBroker(testNS, resources.DefaultBrokerName),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
	}, {
		Name: "Trigger enabled, broker exists with no label",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionDisabledLabelValue)),
			&v1.Broker{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNS,
					Name:      resources.DefaultBrokerName,
				},
			},
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
	}}

	logger := logtesting.TestLogger(t)
	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			eventingClientSet: fakeeventingclient.Get(ctx),
			brokerLister:      listers.GetBrokerLister(),
			isEnabled:         sugar.OnByDefault,
		}
		return trigger.NewReconciler(ctx, logger,
			fakeeventingclient.Get(ctx), listers.GetTriggerLister(),
			controller.GetEventRecorder(ctx), r, controller.Options{SkipStatusUpdates: true})
	}, false, logger))
}

func TestDisabledByDefault(t *testing.T) {
	// Events
	brokerEvent := Eventf(corev1.EventTypeNormal, "BrokerCreated", "Default eventing.knative.dev Broker %q created.", "default")

	broker := resources.MakeBroker(testNS, resources.DefaultBrokerName)
	broker.Annotations = map[string]string{eventing.BrokerClassKey: eventing.MTChannelBrokerClassValue}

	configuredCtx, _ := SetupFakeContext(t)
	configuredCtx = config.ToContext(configuredCtx, &config.Config{
		Defaults: &config.Defaults{
			ClusterDefault: &config.ClassAndBrokerConfig{
				BrokerClass: eventing.MTChannelBrokerClassValue,
			},
		},
	})

	table := TableTest{{
		Name: "bad workqueue key",
		// Make sure Reconcile handles bad keys.
		Key: "too/many/parts",
	}, {
		Name: "key not found",
		// Make sure Reconcile handles good keys that don't exist.
		Key: "foo/not-found",
	}, {
		Name: "Trigger is not labeled",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
		// When we're off by default, nothing happens when the label is missing.
	}, {
		Name: "Trigger is labeled disabled",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionDisabledLabelValue)),
		},
		Ctx: configuredCtx,
		Key: testNS + "/" + triggerName,
	}, {
		Name: "Trigger is deleted no resources",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue),
				WithTriggerDeleted),
		},
		Ctx: configuredCtx,
		Key: testNS + "/" + triggerName,
	}, {
		Name: "Trigger enabled",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue)),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
		WantEvents: []string{
			brokerEvent,
		},
		WantCreates: []runtime.Object{
			broker,
		},
	}, {
		Name: "Trigger enabled, broker exists",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionEnabledLabelValue)),
			resources.MakeBroker(testNS, resources.DefaultBrokerName),
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
	}, {
		Name: "Trigger enabled, broker exists with no label",
		Objects: []runtime.Object{
			NewTrigger(triggerName, testNS, brokerName,
				WithAnnotation(sugar.InjectionLabelKey, sugar.InjectionDisabledLabelValue)),
			&v1.Broker{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testNS,
					Name:      resources.DefaultBrokerName,
				},
			},
		},
		Ctx:                     configuredCtx,
		Key:                     testNS + "/" + triggerName,
		SkipNamespaceValidation: true,
		WantErr:                 false,
	}}

	logger := logtesting.TestLogger(t)
	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher) controller.Reconciler {
		r := &Reconciler{
			eventingClientSet: fakeeventingclient.Get(ctx),
			brokerLister:      listers.GetBrokerLister(),
			isEnabled:         sugar.OffByDefault,
		}
		return trigger.NewReconciler(ctx, logger,
			fakeeventingclient.Get(ctx), listers.GetTriggerLister(),
			controller.GetEventRecorder(ctx), r, controller.Options{SkipStatusUpdates: true})
	}, false, logger))
}
