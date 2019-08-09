package loadtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Load Test Application", func() {
	Describe("normalizeSampleSpace()", func() {
		It("Should change slice contents", func() {
			const pSum float64 = 0.5
			var sampleSpace []operation
			sampleSpace = append(sampleSpace, operation{
				probability: pSum,
			})
			normalizeSampleSpace(sampleSpace, pSum)
			Expect(sampleSpace[0].probability).To(Equal(1.0))
		})
	})

	Describe("getRandomOperationFromSampleSpace()", func() {
		It("Should return first operation", func() {
			var sampleSpace []operation
			sampleSpace = append(sampleSpace, operation{
				probability: 0.2,
			})
			sampleSpace = append(sampleSpace, operation{
				probability: 0.3,
			})
			operation, err := getRandomOperationFromSampleSpace(sampleSpace, 0.15)
			Expect(operation.probability).To(Equal(sampleSpace[0].probability))
			Expect(err).To(BeNil())
		})

		It("Should return second operation", func() {
			var sampleSpace []operation
			sampleSpace = append(sampleSpace, operation{
				probability: 0.2,
			})
			sampleSpace = append(sampleSpace, operation{
				probability: 0.3,
			})
			operation, err := getRandomOperationFromSampleSpace(sampleSpace, 0.25)
			Expect(operation.probability).To(Equal(sampleSpace[1].probability))
			Expect(err).To(BeNil())
		})

		It("Should return last operation", func() {
			var sampleSpace []operation
			sampleSpace = append(sampleSpace, operation{
				probability: 0.2,
			})
			sampleSpace = append(sampleSpace, operation{
				probability: 0.3,
			})
			operation, err := getRandomOperationFromSampleSpace(sampleSpace, 0.5)
			Expect(operation.probability).To(Equal(sampleSpace[1].probability))
			Expect(err).To(BeNil())
		})

		It("Should return an error because dice is larger than one", func() {
			var sampleSpace []operation
			sampleSpace = append(sampleSpace, operation{
				probability: 0.2,
			})
			sampleSpace = append(sampleSpace, operation{
				probability: 0.3,
			})
			_, err := getRandomOperationFromSampleSpace(sampleSpace, 1.1)
			expectedError := &GenericError{"SampleSpaceSumBelowOneError", "Sum of all probabilities is less than one."}
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(Equal(expectedError.Error()))
		})
	})
})
