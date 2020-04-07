package bddTests

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	pb "google.golang.org/grpc/examples/route_guide/routeguide"
)

var _ = Describe("Route Guide Client", func() {
	var _ = Describe("Get feature(s)", func() {
		Context("At one location that has a feature", func() {
			point := &pb.Point{Latitude: 409146138, Longitude: -746188906}

			It("should return a feature name", func() {
				feature, err := clt.GetFeature(ctx, point)
				Expect(feature.Name).To(Equal("Berkshire Valley Management Area Trail, Jefferson, NJ, USA"))
				Expect(feature.Location.Latitude).To(Equal(point.Latitude))
				Expect(feature.Location.Longitude).To(Equal(point.Longitude))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("At one location with feature missing", func() {
			point := &pb.Point{Latitude: 0, Longitude: 0}

			It("should return an empty string", func() {
				feature, err := clt.GetFeature(ctx, point)
				Expect(err).NotTo(HaveOccurred())

				Expect(feature.Name).To(Equal(""))
				Expect(feature.Location.Latitude).To(Equal(point.Latitude))
				Expect(feature.Location.Longitude).To(Equal(point.Longitude))
			})
		})

		Context("When requesting features for a given rectangular area", func() {
			rect := &pb.Rectangle{
				Hi: &pb.Point{Latitude: 420000000, Longitude: -746000000},
				Lo: &pb.Point{Latitude: 400000000, Longitude: -746500000},
			}

			It("should return features inside the area", func() {
				stream, err := clt.ListFeatures(ctx, rect)
				Expect(err).NotTo(HaveOccurred())

				for {
					feature, err := stream.Recv()
					if err == io.EOF {
						break
					}
					Expect(err).NotTo(HaveOccurred())

					// Convert to a valid map key
					location := point{
						latitude:  feature.Location.Latitude,
						longitude: feature.Location.Longitude,
					}
					featureList := getExpectedFeatureList()
					Expect(feature.Name).To(Equal(featureList[location]))
				}
			})
		})
	})
	/*
	    * The remaining tests can be done as an exercise for the blog reader
	    *
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
	   	*
	   	*
	*/
})

/*
 * The below functions are no longer needed as respective tests have been commented out.
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
		&pb.Point{Latitude: 421960920, Longitude: -1227150000},
		&pb.Point{Latitude: 436818360, Longitude: -1241778645},
		&pb.Point{Latitude: 435381599, Longitude: -1232931860},
		&pb.Point{Latitude: 458278421, Longitude: -1236026905},
		&pb.Point{Latitude: 451941630, Longitude: -1206922850},
		&pb.Point{Latitude: 447086560, Longitude: -1184969630},
		&pb.Point{Latitude: 432916610, Longitude: -1178952480},
		&pb.Point{Latitude: 421984550, Longitude: -1214058770},
		&pb.Point{Latitude: 406523420, Longitude: -742135517},
	}
}

*
*/

type point struct {
	latitude  int32
	longitude int32
}

func getExpectedFeatureList() map[point]string {
	return map[point]string{
		point{latitude: 407838351, longitude: -746143763}: "Patriots Path, Mendham, NJ 07945, USA",
		point{latitude: 418858923, longitude: -746156790}: "",
		point{latitude: 409146138, longitude: -746188906}: "Berkshire Valley Management Area Trail, Jefferson, NJ, USA",
		point{latitude: 409642566, longitude: -746017679}: "6 East Emerald Isle Drive, Lake Hopatcong, NJ 07849, USA",
		point{latitude: 409319800, longitude: -746201391}: "11 Ward Street, Mount Arlington, NJ 07856, USA",
		point{latitude: 416560744, longitude: -746721964}: "66 Pleasantview Avenue, Monticello, NY 12701, USA",
		point{latitude: 400066188, longitude: -746793294}: "",
		point{latitude: 404062378, longitude: -746376177}: "",
		point{latitude: 404080723, longitude: -746119569}: "",
		point{latitude: 418465462, longitude: -746859398}: "",
	}
}
