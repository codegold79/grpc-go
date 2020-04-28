package bdd_tests

import (
	"io"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "google.golang.org/grpc/examples/route_guide/routeguide"
)

var _ = Describe("Route Guide Client", func() {
	var _ = Describe("Get feature(s)", func() {
		Context("given a location", func() {
			var (
				feature *pb.Feature
				err     error
				point   *pb.Point
			)

			// Describe, Context, and When are functionally equivalent,
			// but provide semantic nuance.
			When("providing an location with a feature", func() {
				// BeforeEach runs before every spec, i.e. It(),
				// within this When() scope.
				BeforeEach(func() {
					point = &pb.Point{Latitude: 409146138, Longitude: -746188906}
					feature, err = clt.GetFeature(ctx, point)
				})

				It("should return a feature name", func() {
					Expect(feature.Name).To(Equal("Berkshire Valley Management Area Trail, Jefferson, NJ, USA"))
					Expect(feature.Location.Latitude).To(Equal(point.Latitude))
					Expect(feature.Location.Longitude).To(Equal(point.Longitude))
				})

				It("should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("providing a location without a feature", func() {
				BeforeEach(func() {
					point = &pb.Point{Latitude: 0, Longitude: 0}
					feature, err = clt.GetFeature(ctx, point)
				})

				It("should return an empty string", func() {
					Expect(feature.Name).To(BeEmpty())
					Expect(feature.Location.Latitude).To(Equal(point.Latitude))
					Expect(feature.Location.Longitude).To(Equal(point.Longitude))
				})

				It("should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			// These specs were added last.
			When("providing invalid coordinate(s)", func() {
				BeforeEach(func() {
					point = &pb.Point{Latitude: -910000000, Longitude: 810000000}
					feature, err = clt.GetFeature(ctx, point)
				})

				It("should return an invalid coordinate error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("coordinate(s) out of range"))
				})
			})
		})

		Context("given a rectangular area", func() {
			When("stream from server has been established", func() {
				var (
					feature *pb.Feature
					err     error
					rect    *pb.Rectangle
					stream  pb.RouteGuide_ListFeaturesClient
					want    map[point]string
				)

				BeforeEach(func() {
					rect = &pb.Rectangle{
						Hi: &pb.Point{Latitude: 420000000, Longitude: -746000000},
						Lo: &pb.Point{Latitude: 400000000, Longitude: -746500000},
					}
					stream, err = clt.ListFeatures(ctx, rect)

					want = getExpectedFeatureList()
				})

				It("should not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return (empty) features inside the area", func() {
					for {
						feature, err = stream.Recv()
						if err == io.EOF {
							break
						}

						// Convert received feature location to a valid map key
						// for equality comparison.
						location := point{
							latitude:  feature.Location.Latitude,
							longitude: feature.Location.Longitude,
						}

						Expect(feature.Name).To(Equal(want[location]))
					}
				})
			})
		})
	})

	var _ = Describe("Record route", func() {
		Context("When multiple locations are sent", func() {
			points := getRoute()

			It("should return a route summary", func() {
				stream, err := clt.RecordRoute(ctx)
				Expect(err).NotTo(HaveOccurred())

				for _, point := range points {
					err := stream.Send(point)
					Expect(err).NotTo(HaveOccurred())
				}

				reply, err := stream.CloseAndRecv()
				Expect(err).NotTo(HaveOccurred())

				Expect(reply.PointCount).To(Equal(int32(9)))
				Expect(reply.FeatureCount).To(Equal(int32(1)))
				Expect(reply.Distance).To(Equal(int32(5314662)))
			})
		})
	})

	var _ = Describe("Route Chat feature", func() {
		Context("When a client sends notes for a location", func() {
			notesToSend := getNotes()

			It("should recieve all other notes sent for that location", func() {
				wg := sync.WaitGroup{}
				wg.Add(2)

				// Messages received will be stored in a map.
				got := map[string]int{}
				want := getExpectedNotes()

				// Start streaming for both clients.
				stream, err := clt.RouteChat(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Send the messages and then updates.
				go func() {
					for _, note := range notesToSend {
						err = stream.Send(note)
					}
					err = stream.CloseSend()
					wg.Done()
				}()

				// Read the messages while they are being sent.
				go func() {
					for {
						in, err := stream.Recv()
						if err == io.EOF {
							// read complete.
							wg.Done()
							return
						}
						got[in.Message]++
					}
				}()

				wg.Wait()

				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(want))
			})
		})
	})
})

type point struct {
	latitude  int32
	longitude int32
}

// Declutter the body of the tests by returning large amounts of
// expected data from outside.
func getExpectedFeatureList() map[point]string {
	return map[point]string{
		{latitude: 407838351, longitude: -746143763}: "Patriots Path, Mendham, NJ 07945, USA",
		{latitude: 418858923, longitude: -746156790}: "",
		{latitude: 409146138, longitude: -746188906}: "Berkshire Valley Management Area Trail, Jefferson, NJ, USA",
		{latitude: 409642566, longitude: -746017679}: "6 East Emerald Isle Drive, Lake Hopatcong, NJ 07849, USA",
		{latitude: 409319800, longitude: -746201391}: "11 Ward Street, Mount Arlington, NJ 07856, USA",
		{latitude: 416560744, longitude: -746721964}: "66 Pleasantview Avenue, Monticello, NY 12701, USA",
		{latitude: 400066188, longitude: -746793294}: "",
		{latitude: 404062378, longitude: -746376177}: "",
		{latitude: 404080723, longitude: -746119569}: "",
		{latitude: 418465462, longitude: -746859398}: "",
	}
}

func getNotes() []*pb.RouteNote {
	return []*pb.RouteNote{
		{Location: &pb.Point{Latitude: 421960920, Longitude: -1227150000}, Message: "1. Ashland, OR, USA"},
		{Location: &pb.Point{Latitude: 436818360, Longitude: -1241778645}, Message: "2. Wincester Bay, OR, USA"},
		{Location: &pb.Point{Latitude: 435381599, Longitude: -1232931860}, Message: "3. Rice Hill, OR, USA"},
		{Location: &pb.Point{Latitude: 458278421, Longitude: -1236026905}, Message: "4. Lukarilla, OR, USA"},
		{Location: &pb.Point{Latitude: 406523420, Longitude: -742135517}, Message: "5. Elizabeth, NJ, USA"},
		{Location: &pb.Point{Latitude: 421960920, Longitude: -1227150000}, Message: "1. Shakespeare Festival, Ashland, OR, USA"},
		{Location: &pb.Point{Latitude: 436818360, Longitude: -1241778645}, Message: "2. Salmon Harbor Marina, Wincester Bay, OR, USA"},
		{Location: &pb.Point{Latitude: 435381599, Longitude: -1232931860}, Message: "3. Flippin' Chicken Bento, Rice Hill, OR, USA"},
		{Location: &pb.Point{Latitude: 458278421, Longitude: -1236026905}, Message: "4. Nehalem River, Lukarilla, OR, USA"},
		{Location: &pb.Point{Latitude: 406523420, Longitude: -742135517}, Message: "5. Consulate General of El Salvador, Elizabeth, NJ, USA"},
	}
}

func getExpectedNotes() map[string]int {
	return map[string]int{
		"1. Ashland, OR, USA":                                     2,
		"1. Shakespeare Festival, Ashland, OR, USA":               1,
		"2. Salmon Harbor Marina, Wincester Bay, OR, USA":         1,
		"2. Wincester Bay, OR, USA":                               2,
		"3. Flippin' Chicken Bento, Rice Hill, OR, USA":           1,
		"3. Rice Hill, OR, USA":                                   2,
		"4. Lukarilla, OR, USA":                                   2,
		"4. Nehalem River, Lukarilla, OR, USA":                    1,
		"5. Consulate General of El Salvador, Elizabeth, NJ, USA": 1,
		"5. Elizabeth, NJ, USA":                                   2,
	}
}

func getRoute() []*pb.Point {
	return []*pb.Point{
		{Latitude: 421960920, Longitude: -1227150000},
		{Latitude: 436818360, Longitude: -1241778645},
		{Latitude: 435381599, Longitude: -1232931860},
		{Latitude: 458278421, Longitude: -1236026905},
		{Latitude: 451941630, Longitude: -1206922850},
		{Latitude: 447086560, Longitude: -1184969630},
		{Latitude: 432916610, Longitude: -1178952480},
		{Latitude: 421984550, Longitude: -1214058770},
		{Latitude: 406523420, Longitude: -742135517},
	}
}
