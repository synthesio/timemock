package timemock

import (
	"sync"
	"testing"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTimeMock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TimeMock Suite")
}

var _ = Describe("clock", func() {

	var subject *timemockClock
	var stubTime time.Time

	BeforeEach(func() {
		subject = New().(*timemockClock)
		stubTime = time.Unix(1522549800, 0) //Human time (GMT): Sunday, April 1, 2018 2:30:00 AM
	})

	Describe("Freeze", func() {
		Context("when freezing time", func() {
			BeforeEach(func() {
				subject.Freeze(stubTime)
			})

			It("Now should return the frozen time", func() {
				time.Sleep(time.Second)
				Ω(subject.Now()).Should(Equal(stubTime))
			})

			Context("when we unfreeze", func() {
				BeforeEach(func() {
					subject.Return()
				})

				It("should return the current time", func() {
					Ω(subject.Now()).Should(BeTemporally("~", time.Now(), time.Millisecond))
				})
			})

			It("it should return that no time was passed", func() {
				time.Sleep(time.Second)
				Ω(subject.Since(stubTime)).Should(BeZero())
			})

		})
	})

	Describe("Travel", func() {
		Context("when we travel in time", func() {
			var gap time.Duration

			BeforeEach(func() {
				subject.Travel(stubTime)
			})

			It("Now should return the traveled time plus second", func() {
				gap = subject.Now().Sub(stubTime)
				time.Sleep(time.Second)
				Ω(subject.Now()).Should(BeTemporally("~", stubTime.Add(time.Second).Add(gap), 10*time.Millisecond))
			})

			Context("when we un-travel", func() {
				BeforeEach(func() {
					subject.Return()
				})

				It("should return the current time", func() {
					Ω(subject.Now()).Should(BeTemporally("~", time.Now(), time.Millisecond))
				})

				It("it should return that a second passed was passed", func() {
					now := time.Now()
					time.Sleep(time.Second)
					Ω(subject.Since(now)).Should(BeNumerically("~", time.Second, 10*time.Millisecond))
				})
			})

			Context("When scale is 60", func() {
				var scale float64 = 60

				BeforeEach(func() {
					subject.Scale(scale)
				})

				It("it should return that a minute passed", func() {
					time.Sleep(time.Second)
					Ω(subject.Now()).Should(BeTemporally("~", stubTime.Add(time.Minute), time.Second))
				})

				It("it should return that a minute passed was passed", func() {
					time.Sleep(time.Second)
					Ω(subject.Since(stubTime)).Should(BeNumerically("~", time.Minute, time.Second))
				})

			})
		})
	})

	Describe("Scale", func() {
		var scale float64 = 60

		It("should make time run faster", func() {
			beforeScale := subject.Now()
			subject.Scale(scale)
			time.Sleep(time.Second)
			Ω(subject.Now()).Should(BeTemporally("~", beforeScale.Add(time.Minute), time.Second))

		})

	})

	Describe("Since", func() {
		It("should return that a second was passed", func() {
			now := time.Now()
			time.Sleep(time.Second)
			Ω(subject.Since(now)).Should(BeNumerically("~", time.Second, 10*time.Millisecond))
		})
	})

	Describe("Until", func() {
		It("should return that a second will pass", func() {
			t := time.Now().Add(1 * time.Second)
			Ω(subject.Until(t)).Should(BeNumerically("~", time.Second, 10*time.Millisecond))
		})
	})

})

func TestRaces(t *testing.T) {
	clock := New()
	refTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

	n := 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			for j := 0; j < n; j++ {
				clock.Now()
				clock.Freeze(refTime)
				clock.Return()
				clock.Travel(refTime)
				clock.Scale(2)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkClock(b *testing.B) {
	b.Run("Freezed", func(b *testing.B) {
		clock := New()
		refTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
		clock.Freeze(refTime)
		for i := 0; i < b.N; i++ {
			clock.Now()
		}
	})
	b.Run("Unfreezed", func(b *testing.B) {
		clock := New()
		for i := 0; i < b.N; i++ {
			clock.Now()
		}
	})
	b.Run("Stdlib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.Now()
		}
	})
}
