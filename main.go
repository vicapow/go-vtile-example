package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/vicapow/go-vtile-example/gen/third-party/vector-tile-spec/1.0.1"
)

func cmdEnc(id uint32, count uint32) uint32 {
	return (id & 0x7) | (count << 3)
}

func moveTo(count uint32) uint32 {
	return cmdEnc(1, count)
}

func lineTo(count uint32) uint32 {
	return cmdEnc(2, count)
}

func closePath(count uint32) uint32 {
	return cmdEnc(7, count)
}

func paramEnc(value int32) int32 {
	return (value << 1) ^ (value >> 31)
}

func createTileWithPoints(xyz TileID, points []LngLat) ([]byte, error) {
	tile := &vector_tile.Tile{}
	var layerVersion = vector_tile.Default_Tile_Layer_Version
	layerName := "points"
	featureType := vector_tile.Tile_POINT
	var extent = vector_tile.Default_Tile_Layer_Extent
	// Put a point in the center of the tile.
	var geometry []uint32
	var filtered [][2]float64
	for _, point := range points {
		x, y := lngLatToTileXY(point, xyz)
		if x >= 0 && x < 1 && y >= 0 && y < 1 {
			filtered = append(filtered, [2]float64{x, y})
		}
	}
	if len(filtered) > 0 {
		cmd := moveTo(uint32(len(filtered)))
		geometry = append(geometry, cmd)
		var pX int32
		var pY int32
		for _, point := range filtered {
			deltaX := int32(float64(extent)*point[0]+0.5) - pX
			deltaY := int32(float64(extent)*point[1]+0.5) - pY
			geometry = append(geometry, uint32(paramEnc(deltaX)))
			geometry = append(geometry, uint32(paramEnc(deltaY)))
			pX = pX + deltaX
			pY = pY + deltaY
		}
	} else {
		// Return an empty tile if we have no points
		return nil, nil
	}
	tile.Layers = []*vector_tile.Tile_Layer{
		&vector_tile.Tile_Layer{
			Version: &layerVersion,
			Name:    &layerName,
			Extent:  &extent,
			Features: []*vector_tile.Tile_Feature{
				&vector_tile.Tile_Feature{
					Tags:     []uint32{},
					Type:     &featureType,
					Geometry: geometry,
				},
			},
		},
	}
	return proto.Marshal(tile)
}

func xyzToLngLat(tileX float64, tileY float64, tileZ float64) (float64, float64) {
	totalTilesX := math.Pow(2, tileZ)
	totalTilesY := math.Pow(2, tileZ)
	x := float64(tileX) / float64(totalTilesX)
	y := float64(tileY) / float64(totalTilesY)
	// lambda can go from [-pi/2, pi/2]
	lambda := x*math.Pi*2 - math.Pi
	// phi can go from [-1.4844, 1.4844]
	phi := 2*math.Atan(math.Exp((2*y-1)*math.Pi)) - (math.Pi / 2)
	lng := lambda * 180 / math.Pi
	lat := (math.Pi - phi) * 180 / math.Pi
	return lng, lat
}

func lngLatToTileXY(ll LngLat, tile TileID) (float64, float64) {
	totalTilesX := math.Pow(2, float64(tile.z))
	totalTilesY := math.Pow(2, float64(tile.z))
	lambda := (ll.lng + 180) / 180 * math.Pi
	// phi: [-pi/2, pi/2]
	phi := ll.lat / 180 * math.Pi
	tileX := lambda / (2 * math.Pi) * totalTilesX
	// [-1.4844, 1.4844] -> [1, 0]  * totalTilesY
	tileY := (math.Log(math.Tan(math.Pi/4-phi/2))/math.Pi/2 + 0.5) * totalTilesY
	return tileX - float64(tile.x), tileY - float64(tile.y)
}

// Takes a string of the form `<z>/<x>/<y>` (for example, 1/2/3) and returns
// the individual uint32 values for x, y, and z if there was no error.
// Otherwise, err is set to a non `nil` value and x, y, z are set to 0.
func tilePathToXYZ(path string) (TileID, error) {
	xyzReg := regexp.MustCompile("(?P<z>[0-9]+)/(?P<x>[0-9]+)/(?P<y>[0-9]+)")
	matches := xyzReg.FindStringSubmatch(path)
	if len(matches) == 0 {
		return TileID{}, errors.New("Unable to parse path as tile")
	}
	x, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	y, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	z, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return TileID{}, err
	}
	return TileID{x: uint32(x), y: uint32(y), z: uint32(z)}, nil
}

// A LngLat is a struct that holds a longitude and latitude vale.
type LngLat struct {
	lng float64
	lat float64
}

// TileID represents the id of the tile.
type TileID struct {
	x uint32
	y uint32
	z uint32
}

// Tree a struct holder for tree information.
type Tree struct {
	lng     float64
	lat     float64
	species string
}

func loadTrees() []Tree {
	content, err := ioutil.ReadFile("./trees.csv")
	if err != nil {
		log.Fatal(err)
	}
	r := csv.NewReader(strings.NewReader(string(content[:])))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	// TreeID,qLegalStatus,qSpecies,qAddress,SiteOrder,qSiteInfo,PlantType,qCaretaker,qCareAssistant,PlantDate,DBH,PlotSize,PermitNotes,XCoord,YCoord,Latitude,Longitude,Location
	var trees []Tree
	for _, record := range records[1:] {
		lat, _ := strconv.ParseFloat(record[15], 64)
		lng, _ := strconv.ParseFloat(record[16], 64)
		species := record[2]
		trees = append(trees, Tree{lng: lng, lat: lat, species: species})
	}
	return trees
}

func main() {
	trees := loadTrees()
	points := make([]LngLat, len(trees), len(trees))
	for i, tree := range trees {
		points[i] = LngLat{lng: tree.lng, lat: tree.lat}
	}
	// points = points[0:1000]
	fmt.Println("number of points", len(points))
	mux := http.NewServeMux()

	// Handle requests for urls of the form `/tiles/{z}/{x}/{y}` and returns
	// the vector tile for the even tile x, y, and z coordinates.
	tileBase := "/tiles/"
	mux.HandleFunc(tileBase, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("url: %s", r.URL.Path)
		tilePart := r.URL.Path[len(tileBase):]
		xyz, err := tilePathToXYZ(tilePart)
		if err != nil {
			http.Error(w, "Invalid tile url", 400)
			return
		}
		data, err := createTileWithPoints(xyz, points)
		if err != nil {
			log.Fatal("error generating tile", err)
		}
		// All this APi to be requests from other domains.
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Write(data)
	})
	log.Fatal(http.ListenAndServe(":8080", mux))
}
