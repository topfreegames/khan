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
