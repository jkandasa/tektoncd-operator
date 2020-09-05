/*
Copyright 2019 The Knative Authors

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

package v1alpha1

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	apistest "knative.dev/pkg/apis/testing"
)

func TestTekonPipelineGroupVersionKind(t *testing.T) {
	r := &TekonPipeline{}
	want := schema.GroupVersionKind{
		Group:   GroupName,
		Version: SchemaVersion,
		Kind:    KindTekonPipeline,
	}
	if got := r.GroupVersionKind(); got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestTekonPipelineHappyPath(t *testing.T) {
	ks := &TekonPipelineStatus{}
	ks.InitializeConditions()

	apistest.CheckConditionOngoing(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionOngoing(ks, InstallSucceeded, t)

	ks.MarkVersionMigrationEligible()

	// Install succeeds.
	ks.MarkInstallSucceeded()
	// Dependencies are assumed successful too.
	apistest.CheckConditionSucceeded(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)

	// Deployments are not available at first.
	ks.MarkDeploymentsNotReady()
	apistest.CheckConditionSucceeded(ks, DependenciesInstalled, t)
	apistest.CheckConditionFailed(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
	if ready := ks.IsReady(); ready {
		t.Errorf("ks.IsReady() = %v, want false", ready)
	}

	// Deployments become ready and we're good.
	ks.MarkDeploymentsAvailable()
	apistest.CheckConditionSucceeded(ks, DependenciesInstalled, t)
	apistest.CheckConditionSucceeded(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
	if ready := ks.IsReady(); !ready {
		t.Errorf("ks.IsReady() = %v, want true", ready)
	}
}

func TestTekonPipelineErrorPath(t *testing.T) {
	ks := &TekonPipelineStatus{}
	ks.InitializeConditions()

	apistest.CheckConditionOngoing(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionOngoing(ks, InstallSucceeded, t)

	ks.MarkVersionMigrationEligible()

	// Install fails.
	ks.MarkInstallFailed("test")
	apistest.CheckConditionOngoing(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionFailed(ks, InstallSucceeded, t)

	// Dependencies are installing.
	ks.MarkDependencyInstalling("testing")
	apistest.CheckConditionFailed(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionFailed(ks, InstallSucceeded, t)

	// Install now succeeds.
	ks.MarkInstallSucceeded()
	apistest.CheckConditionFailed(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
	if ready := ks.IsReady(); ready {
		t.Errorf("ks.IsReady() = %v, want false", ready)
	}

	// Deployments become ready
	ks.MarkDeploymentsAvailable()
	apistest.CheckConditionFailed(ks, DependenciesInstalled, t)
	apistest.CheckConditionSucceeded(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
	if ready := ks.IsReady(); ready {
		t.Errorf("ks.IsReady() = %v, want false", ready)
	}

	// Finally, dependencies become available.
	ks.MarkDependenciesInstalled()
	apistest.CheckConditionSucceeded(ks, DependenciesInstalled, t)
	apistest.CheckConditionSucceeded(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
	if ready := ks.IsReady(); !ready {
		t.Errorf("ks.IsReady() = %v, want true", ready)
	}
}

func TestTekonPipelineExternalDependency(t *testing.T) {
	ks := &TekonPipelineStatus{}
	ks.InitializeConditions()

	// External marks dependency as failed.
	ks.MarkDependencyMissing("test")

	// Install succeeds.
	ks.MarkInstallSucceeded()
	apistest.CheckConditionFailed(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)

	// Dependencies are now ready.
	ks.MarkDependenciesInstalled()
	apistest.CheckConditionSucceeded(ks, DependenciesInstalled, t)
	apistest.CheckConditionOngoing(ks, DeploymentsAvailable, t)
	apistest.CheckConditionSucceeded(ks, InstallSucceeded, t)
}

func TestTekonPipelineVersionMigrationNotEligible(t *testing.T) {
	ks := &TekonPipelineStatus{}
	ks.InitializeConditions()

	ks.MarkVersionMigrationNotEligible("Version migration not eligible.")
	apistest.CheckConditionFailed(ks, VersionMigrationEligible, t)
}