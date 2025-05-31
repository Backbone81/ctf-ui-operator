package ctfdapi_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/ctf-ui-operator/internal/ctfdapi"
)

var _ = Describe("DateOnly", func() {
	It("should correctly marshal to json", func(ctx SpecContext) {
		date := ctfdapi.NewDateOnly(time.Date(2025, 5, 31, 21, 59, 30, 200, time.UTC))
		data, err := json.Marshal(date)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal(`"2025-05-31"`))
	})

	It("should correctly unmarshal from json", func(ctx SpecContext) {
		var date ctfdapi.DateOnly
		Expect(json.Unmarshal([]byte(`"2025-05-31"`), &date)).To(Succeed())
		Expect(date).To(Equal(ctfdapi.NewDateOnly(time.Date(2025, 5, 31, 0, 0, 0, 0, time.UTC))))
	})

	It("should correctly marshal the zero value", func(ctx SpecContext) {
		date := ctfdapi.NewDateOnly(time.Time{})
		data, err := json.Marshal(date)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(Equal(`null`))
	})

	It("should correctly unmarshal the empty string", func(ctx SpecContext) {
		var date ctfdapi.DateOnly
		Expect(json.Unmarshal([]byte(`""`), &date)).To(Succeed())
		Expect(date).To(Equal(ctfdapi.NewDateOnly(time.Time{})))
	})
})
