/*
Copyright 2022.

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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	simulationv1alpha1 "sigs.k8s.io/kube-scheduler-simulator/scenario/api/v1alpha1"
	// TODO (user): Add any additional imports if needed
)

var _ = Describe("Scenario Webhook", func() {
	var (
		obj       *simulationv1alpha1.Scenario
		oldObj    *simulationv1alpha1.Scenario
		validator ScenarioCustomValidator
	)

	BeforeEach(func() {
		obj = &simulationv1alpha1.Scenario{}
		oldObj = &simulationv1alpha1.Scenario{}
		validator = ScenarioCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
		// TODO (user): Add any setup logic common to all tests
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating or updating Scenario under Validating Webhook", func() {
		It("Should deny creation if there are some operation fields", func() {
			By("simulating an invalid creation scenario")
			obj.Spec.Operations = []*simulationv1alpha1.ScenarioOperation{
				{
					Create: &simulationv1alpha1.CreateOperation{},
					Delete: &simulationv1alpha1.DeleteOperation{},
				},
			}
			Expect(validator.ValidateCreate(ctx, obj)).Error().To(HaveOccurred())
		})

		It("Should validate create correctly", func() {
			By("simulating a valid creation scenario")
			obj.Spec.Operations = []*simulationv1alpha1.ScenarioOperation{
				{
					Create: &simulationv1alpha1.CreateOperation{},
				},
			}
			Expect(validator.ValidateCreate(ctx, obj)).To(BeNil())
		})

		It("Should deny update if there are some operation fields", func() {
			By("simulating an invalid creation scenario")
			obj.Spec.Operations = []*simulationv1alpha1.ScenarioOperation{
				{
					Create: &simulationv1alpha1.CreateOperation{},
					Delete: &simulationv1alpha1.DeleteOperation{},
				},
			}
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).Error().To(HaveOccurred())
		})

		It("Should validate updates correctly", func() {
			By("simulating a valid update scenario")
			obj.Spec.Operations = []*simulationv1alpha1.ScenarioOperation{
				{
					Create: &simulationv1alpha1.CreateOperation{},
				},
			}
			Expect(validator.ValidateUpdate(ctx, oldObj, obj)).To(BeNil())
		})
	})

})
