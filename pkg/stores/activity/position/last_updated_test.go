package position_test

import (
	"time"

	timeProvider "github.com/backtesting-org/kronos-sdk/pkg/runtime/time"
	"github.com/backtesting-org/kronos-sdk/pkg/stores/activity/position"
	activityTypes "github.com/backtesting-org/kronos-sdk/pkg/types/data/stores/activity"
	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position Store - Last Updated", func() {
	var (
		store    activityTypes.Positions
		provider temporal.TimeProvider
	)

	BeforeEach(func() {
		provider = timeProvider.NewTimeProvider()
		store = position.NewStore(provider)
	})

	Describe("UpdateLastUpdated", func() {
		Context("when updating timestamps", func() {
			It("should store a new timestamp for a key", func() {
				key := activityTypes.UpdateKey("strategy-1-orders")

				store.UpdateLastUpdated(key)

				lastUpdated := store.GetLastUpdated()
				Expect(lastUpdated).To(HaveKey(key))
				Expect(lastUpdated[key]).To(BeTemporally("~", time.Now(), time.Second))
			})

			It("should handle multiple different keys", func() {
				key1 := activityTypes.UpdateKey("strategy-1-orders")
				key2 := activityTypes.UpdateKey("strategy-1-trades")
				key3 := activityTypes.UpdateKey("strategy-2-orders")

				store.UpdateLastUpdated(key1)
				store.UpdateLastUpdated(key2)
				store.UpdateLastUpdated(key3)

				lastUpdated := store.GetLastUpdated()
				Expect(lastUpdated).To(HaveLen(3))
				Expect(lastUpdated).To(HaveKey(key1))
				Expect(lastUpdated).To(HaveKey(key2))
				Expect(lastUpdated).To(HaveKey(key3))
			})

			It("should update timestamp when called again with same key", func() {
				key := activityTypes.UpdateKey("strategy-1-orders")

				store.UpdateLastUpdated(key)
				firstUpdate := store.GetLastUpdated()[key]

				// Wait a tiny bit
				time.Sleep(10 * time.Millisecond)

				store.UpdateLastUpdated(key)
				secondUpdate := store.GetLastUpdated()[key]

				Expect(secondUpdate).To(BeTemporally(">", firstUpdate))
			})
		})
	})

	Describe("GetLastUpdated", func() {
		Context("when no updates have been made", func() {
			It("should return an empty map", func() {
				lastUpdated := store.GetLastUpdated()
				Expect(lastUpdated).To(BeEmpty())
			})
		})

		Context("when updates exist", func() {
			It("should return all tracked timestamps", func() {
				store.UpdateLastUpdated(activityTypes.UpdateKey("key1"))
				store.UpdateLastUpdated(activityTypes.UpdateKey("key2"))

				lastUpdated := store.GetLastUpdated()

				Expect(lastUpdated).To(HaveLen(2))
			})
		})
	})

	Describe("Clear", func() {
		Context("when clearing the store", func() {
			It("should also clear last updated timestamps", func() {
				store.UpdateLastUpdated(activityTypes.UpdateKey("key1"))
				store.UpdateLastUpdated(activityTypes.UpdateKey("key2"))

				store.Clear()

				lastUpdated := store.GetLastUpdated()
				Expect(lastUpdated).To(BeEmpty())
			})
		})
	})

	Describe("Concurrent access", func() {
		Context("when multiple goroutines update timestamps", func() {
			It("should handle concurrent writes safely", func() {
				done := make(chan bool)
				iterations := 50

				// Writer 1
				go func() {
					for i := 0; i < iterations; i++ {
						store.UpdateLastUpdated(activityTypes.UpdateKey("key-a"))
					}
					done <- true
				}()

				// Writer 2
				go func() {
					for i := 0; i < iterations; i++ {
						store.UpdateLastUpdated(activityTypes.UpdateKey("key-b"))
					}
					done <- true
				}()

				// Reader
				go func() {
					for i := 0; i < iterations; i++ {
						_ = store.GetLastUpdated()
					}
					done <- true
				}()

				<-done
				<-done
				<-done

				// Both keys should exist
				lastUpdated := store.GetLastUpdated()
				Expect(lastUpdated).To(HaveLen(2))
			})
		})
	})
})
