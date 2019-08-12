package loadtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getNewTestSampleSpace() []operation {
	var sampleSpace []operation
	sampleSpace = append(sampleSpace, operation{
		probability: 0.2,
	})
	sampleSpace = append(sampleSpace, operation{
		probability: 0.3,
	})
	return sampleSpace
}

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
			sampleSpace := getNewTestSampleSpace()
			operation := getRandomOperationFromSampleSpace(sampleSpace, 0.15)
			Expect(operation.probability).To(Equal(sampleSpace[0].probability))
		})

		It("Should return second operation", func() {
			sampleSpace := getNewTestSampleSpace()
			operation := getRandomOperationFromSampleSpace(sampleSpace, 0.25)
			Expect(operation.probability).To(Equal(sampleSpace[1].probability))
		})

		It("Should return last operation", func() {
			sampleSpace := getNewTestSampleSpace()
			operation := getRandomOperationFromSampleSpace(sampleSpace, 0.5)
			Expect(operation.probability).To(Equal(sampleSpace[1].probability))
		})

		It("Should return the first operation because dice is larger than one", func() {
			sampleSpace := getNewTestSampleSpace()
			operation := getRandomOperationFromSampleSpace(sampleSpace, 1.1)
			Expect(operation.probability).To(Equal(sampleSpace[0].probability))
		})
	})
})
