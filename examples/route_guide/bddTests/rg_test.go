package bddTests

import (
	"context"
	"io"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/route_guide/routeguide"
)

var _ = Describe("Route Guide Client", func() {
	var (
		clt    pb.RouteGuideClient
		ctx    context.Context
		cancel context.CancelFunc
		conn   *grpc.ClientConn
		err    error
	)

	BeforeEach(func() {
		conn, err = grpc.Dial("localhost:10000", grpc.WithInsecure(), grpc.WithBlock())
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		clt = pb.NewRouteGuideClient(conn)
	})

	AfterEach(func() {
		conn.Close()
		cancel()
	})

	var _ = Describe("Get feature(s)", func() {
		Context("At one location", func() {
			It("should return a feature name", func() {
				point := &pb.Point{Latitude: 409146138, Longitude: -746188906}
				feature, err := clt.GetFeature(ctx, point)
				Expect(err).NotTo(HaveOccurred())

				Expect(feature.Name).To(Equal("Berkshire Valley Management Area Trail, Jefferson, NJ, USA"))
				Expect(feature.Location.Latitude).To(Equal(point.Latitude))
				Expect(feature.Location.Longitude).To(Equal(point.Longitude))
			})
		})

		Context("At one location with feature missing", func() {
			It("should return a feature name", func() {
				point := &pb.Point{Latitude: 0, Longitude: 0}
				feature, err := clt.GetFeature(ctx, point)
				Expect(err).NotTo(HaveOccurred())

				Expect(feature.Name).To(Equal(""))
				Expect(feature.Location.Latitude).To(Equal(point.Latitude))
				Expect(feature.Location.Longitude).To(Equal(point.Longitude))
			})
		})

		Context("Inside a rectangular area", func() {
			It("should return features inside the area", func() {
				rect := &pb.Rectangle{
					Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
					Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
				}

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

	var _ = Describe("Record route", func() {
		Context("With multiple locations", func() {
			It("should return a route summary", func() {
				points := getRoute()

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

	var _ = Describe("Test the Route Chat feature", func() {
		Context("A client sends and receives notes", func() {
			It("should recieve notes while they are being sent", func() {
				wg := sync.WaitGroup{}
				wg.Add(2)

				notesToSend := getNotes()

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

func getExpectedFeatureList() map[point]string {
	return map[point]string{
		point{latitude: 407838351, longitude: -746143763}: "Patriots Path, Mendham, NJ 07945, USA",
		point{latitude: 408122808, longitude: -743999179}: "101 New Jersey 10, Whippany, NJ 07981, USA",
		point{latitude: 413628156, longitude: -749015468}: "U.S. 6, Shohola, PA 18458, USA",
		point{latitude: 419999544, longitude: -740371136}: "5 Conners Road, Kingston, NY 12401, USA",
		point{latitude: 414008389, longitude: -743951297}: "Mid Hudson Psychiatric Center, New Hampton, NY 10958, USA",
		point{latitude: 419611318, longitude: -746524769}: "287 Flugertown Road, Livingston Manor, NY 12758, USA",
		point{latitude: 406109563, longitude: -742186778}: "4001 Tremley Point Road, Linden, NJ 07036, USA",
		point{latitude: 416802456, longitude: -742370183}: "352 South Mountain Road, Wallkill, NY 12589, USA",
		point{latitude: 412950425, longitude: -741077389}: "Bailey Turn Road, Harriman, NY 10926, USA",
		point{latitude: 412144655, longitude: -743949739}: "193-199 Wawayanda Road, Hewitt, NJ 07421, USA",
		point{latitude: 415736605, longitude: -742847522}: "406-496 Ward Avenue, Pine Bush, NY 12566, USA",
		point{latitude: 413843930, longitude: -740501726}: "162 Merrill Road, Highland Mills, NY 10930, USA",
		point{latitude: 410873075, longitude: -744459023}: "Clinton Road, West Milford, NJ 07480, USA",
		point{latitude: 412346009, longitude: -744026814}: "16 Old Brook Lane, Warwick, NY 10990, USA",
		point{latitude: 402948455, longitude: -747903913}: "3 Drake Lane, Pennington, NJ 08534, USA",
		point{latitude: 406337092, longitude: -740122226}: "6324 8th Avenue, Brooklyn, NY 11220, USA",
		point{latitude: 406421967, longitude: -747727624}: "1 Merck Access Road, Whitehouse Station, NJ 08889, USA",
		point{latitude: 416318082, longitude: -749677716}: "78-98 Schalck Road, Narrowsburg, NY 12764, USA",
		point{latitude: 415301720, longitude: -748416257}: "282 Lakeview Drive Road, Highland Lake, NY 12743, USA",
		point{latitude: 402647019, longitude: -747071791}: "330 Evelyn Avenue, Hamilton Township, NJ 08619, USA",
		point{latitude: 412567807, longitude: -741058078}: "New York State Reference Route 987E, Southfields, NY 10975, USA",
		point{latitude: 416855156, longitude: -744420597}: "103-271 Tempaloni Road, Ellenville, NY 12428, USA",
		point{latitude: 404663628, longitude: -744820157}: "1300 Airport Road, North Brunswick Township, NJ 08902, USA",
		point{latitude: 407113723, longitude: -749746483}: "",
		point{latitude: 402133926, longitude: -743613249}: "",
		point{latitude: 400273442, longitude: -741220915}: "",
		point{latitude: 411236786, longitude: -744070769}: "",
		point{latitude: 411633782, longitude: -746784970}: "211-225 Plains Road, Augusta, NJ 07822, USA",
		point{latitude: 415830701, longitude: -742952812}: "",
		point{latitude: 413447164, longitude: -748712898}: "165 Pedersen Ridge Road, Milford, PA 18337, USA",
		point{latitude: 405047245, longitude: -749800722}: "100-122 Locktown Road, Frenchtown, NJ 08825, USA",
		point{latitude: 418858923, longitude: -746156790}: "",
		point{latitude: 417951888, longitude: -748484944}: "650-652 Willi Hill Road, Swan Lake, NY 12783, USA",
		point{latitude: 407033786, longitude: -743977337}: "26 East 3rd Street, New Providence, NJ 07974, USA",
		point{latitude: 417548014, longitude: -740075041}: "",
		point{latitude: 410395868, longitude: -744972325}: "",
		point{latitude: 404615353, longitude: -745129803}: "",
		point{latitude: 406589790, longitude: -743560121}: "611 Lawrence Avenue, Westfield, NJ 07090, USA",
		point{latitude: 414653148, longitude: -740477477}: "18 Lannis Avenue, New Windsor, NY 12553, USA",
		point{latitude: 405957808, longitude: -743255336}: "82-104 Amherst Avenue, Colonia, NJ 07067, USA",
		point{latitude: 411733589, longitude: -741648093}: "170 Seven Lakes Drive, Sloatsburg, NY 10974, USA",
		point{latitude: 412676291, longitude: -742606606}: "1270 Lakes Road, Monroe, NY 10950, USA",
		point{latitude: 409224445, longitude: -748286738}: "509-535 Alphano Road, Great Meadows, NJ 07838, USA",
		point{latitude: 406523420, longitude: -742135517}: "652 Garden Street, Elizabeth, NJ 07202, USA",
		point{latitude: 401827388, longitude: -740294537}: "349 Sea Spray Court, Neptune City, NJ 07753, USA",
		point{latitude: 410564152, longitude: -743685054}: "13-17 Stanley Street, West Milford, NJ 07480, USA",
		point{latitude: 408472324, longitude: -740726046}: "47 Industrial Avenue, Teterboro, NJ 07608, USA",
		point{latitude: 412452168, longitude: -740214052}: "5 White Oak Lane, Stony Point, NY 10980, USA",
		point{latitude: 409146138, longitude: -746188906}: "Berkshire Valley Management Area Trail, Jefferson, NJ, USA",
		point{latitude: 404701380, longitude: -744781745}: "1007 Jersey Avenue, New Brunswick, NJ 08901, USA",
		point{latitude: 409642566, longitude: -746017679}: "6 East Emerald Isle Drive, Lake Hopatcong, NJ 07849, USA",
		point{latitude: 408031728, longitude: -748645385}: "1358-1474 New Jersey 57, Port Murray, NJ 07865, USA",
		point{latitude: 413700272, longitude: -742135189}: "367 Prospect Road, Chester, NY 10918, USA",
		point{latitude: 404310607, longitude: -740282632}: "10 Simon Lake Drive, Atlantic Highlands, NJ 07716, USA",
		point{latitude: 409319800, longitude: -746201391}: "11 Ward Street, Mount Arlington, NJ 07856, USA",
		point{latitude: 406685311, longitude: -742108603}: "300-398 Jefferson Avenue, Elizabeth, NJ 07201, USA",
		point{latitude: 419018117, longitude: -749142781}: "43 Dreher Road, Roscoe, NY 12776, USA",
		point{latitude: 412856162, longitude: -745148837}: "Swan Street, Pine Island, NY 10969, USA",
		point{latitude: 416560744, longitude: -746721964}: "66 Pleasantview Avenue, Monticello, NY 12701, USA",
		point{latitude: 405314270, longitude: -749836354}: "",
		point{latitude: 414219548, longitude: -743327440}: "",
		point{latitude: 415534177, longitude: -742900616}: "565 Winding Hills Road, Montgomery, NY 12549, USA",
		point{latitude: 406898530, longitude: -749127080}: "231 Rocky Run Road, Glen Gardner, NJ 08826, USA",
		point{latitude: 407586880, longitude: -741670168}: "100 Mount Pleasant Avenue, Newark, NJ 07104, USA",
		point{latitude: 400106455, longitude: -742870190}: "517-521 Huntington Drive, Manchester Township, NJ 08759, USA",
		point{latitude: 400066188, longitude: -746793294}: "",
		point{latitude: 418803880, longitude: -744102673}: "40 Mountain Road, Napanoch, NY 12458, USA",
		point{latitude: 414204288, longitude: -747895140}: "",
		point{latitude: 414777405, longitude: -740615601}: "",
		point{latitude: 415464475, longitude: -747175374}: "48 North Road, Forestburgh, NY 12777, USA",
		point{latitude: 404062378, longitude: -746376177}: "",
		point{latitude: 405688272, longitude: -749285130}: "",
		point{latitude: 400342070, longitude: -748788996}: "",
		point{latitude: 401809022, longitude: -744157964}: "",
		point{latitude: 404226644, longitude: -740517141}: "9 Thompson Avenue, Leonardo, NJ 07737, USA",
		point{latitude: 410322033, longitude: -747871659}: "",
		point{latitude: 407100674, longitude: -747742727}: "",
		point{latitude: 418811433, longitude: -741718005}: "213 Bush Road, Stone Ridge, NY 12484, USA",
		point{latitude: 415034302, longitude: -743850945}: "",
		point{latitude: 411349992, longitude: -743694161}: "",
		point{latitude: 404839914, longitude: -744759616}: "1-17 Bergen Court, New Brunswick, NJ 08901, USA",
		point{latitude: 414638017, longitude: -745957854}: "35 Oakland Valley Road, Cuddebackville, NY 12729, USA",
		point{latitude: 412127800, longitude: -740173578}: "",
		point{latitude: 401263460, longitude: -747964303}: "",
		point{latitude: 412843391, longitude: -749086026}: "",
		point{latitude: 418512773, longitude: -743067823}: "",
		point{latitude: 404318328, longitude: -740835638}: "42-102 Main Street, Belford, NJ 07718, USA",
		point{latitude: 419020746, longitude: -741172328}: "",
		point{latitude: 404080723, longitude: -746119569}: "",
		point{latitude: 401012643, longitude: -744035134}: "",
		point{latitude: 404306372, longitude: -741079661}: "",
		point{latitude: 403966326, longitude: -748519297}: "",
		point{latitude: 405002031, longitude: -748407866}: "",
		point{latitude: 409532885, longitude: -742200683}: "",
		point{latitude: 416851321, longitude: -742674555}: "",
		point{latitude: 406411633, longitude: -741722051}: "3387 Richmond Terrace, Staten Island, NY 10303, USA",
		point{latitude: 413069058, longitude: -744597778}: "261 Van Sickle Road, Goshen, NY 10924, USA",
		point{latitude: 418465462, longitude: -746859398}: "",
		point{latitude: 411733222, longitude: -744228360}: "",
		point{latitude: 410248224, longitude: -747127767}: "3 Hasta Way, Newton, NJ 07860, USA",
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

func getNotes() []*pb.RouteNote {
	return []*pb.RouteNote{
		{Location: &pb.Point{Latitude: 421960920, Longitude: -1227150000}, Message: "1. Ashland, OR, USA"},
		{Location: &pb.Point{Latitude: 436818360, Longitude: -1241778645}, Message: "2. Wincester Bay, OR, USA"},
		{Location: &pb.Point{Latitude: 435381599, Longitude: -1232931860}, Message: "3. Rice Hill, OR, USA"},
		{Location: &pb.Point{Latitude: 458278421, Longitude: -1236026905}, Message: "4. Lukarilla, OR, USA"},
		{Location: &pb.Point{Latitude: 451941630, Longitude: -1206922850}, Message: "5. Kent, OR, USA"},
		{Location: &pb.Point{Latitude: 447086560, Longitude: -1184969630}, Message: "6. Greenhorn, OR, USA"},
		{Location: &pb.Point{Latitude: 432916610, Longitude: -1178952480}, Message: "7. Crowley, OR, USA"},
		{Location: &pb.Point{Latitude: 421984550, Longitude: -1214058770}, Message: "8. Bonanza, OR, USA"},
		{Location: &pb.Point{Latitude: 406523420, Longitude: -742135517}, Message: "9. Elizabeth, NJ, USA"},
		{Location: &pb.Point{Latitude: 421960920, Longitude: -1227150000}, Message: "1. Shakespeare Festival, Ashland, OR, USA"},
		{Location: &pb.Point{Latitude: 436818360, Longitude: -1241778645}, Message: "2. Salmon Harbor Marina, Wincester Bay, OR, USA"},
		{Location: &pb.Point{Latitude: 435381599, Longitude: -1232931860}, Message: "3. Flippin' Chicken Bento, Rice Hill, OR, USA"},
		{Location: &pb.Point{Latitude: 458278421, Longitude: -1236026905}, Message: "4. Nehalem River, Lukarilla, OR, USA"},
		{Location: &pb.Point{Latitude: 451941630, Longitude: -1206922850}, Message: "5. United States Postal Service, Kent, OR, USA"},
		{Location: &pb.Point{Latitude: 447086560, Longitude: -1184969630}, Message: "6. Main St, Greenhorn, OR, USA"},
		{Location: &pb.Point{Latitude: 432916610, Longitude: -1178952480}, Message: "7. Crowley Rd, Crowley, OR, USA"},
		{Location: &pb.Point{Latitude: 421984550, Longitude: -1214058770}, Message: "8. Bonanza RV Park, Bonanza, OR, USA"},
		{Location: &pb.Point{Latitude: 406523420, Longitude: -742135517}, Message: "9. Consulate General of El Salvador, Elizabeth, NJ, USA"},
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
		"5. Kent, OR, USA":                                        2,
		"5. United States Postal Service, Kent, OR, USA":          1,
		"6. Greenhorn, OR, USA":                                   2,
		"6. Main St, Greenhorn, OR, USA":                          1,
		"7. Crowley Rd, Crowley, OR, USA":                         1,
		"7. Crowley, OR, USA":                                     2,
		"8. Bonanza RV Park, Bonanza, OR, USA":                    1,
		"8. Bonanza, OR, USA":                                     2,
		"9. Consulate General of El Salvador, Elizabeth, NJ, USA": 1,
		"9. Elizabeth, NJ, USA":                                   2,
	}
}
