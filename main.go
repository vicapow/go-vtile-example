package main

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"

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

func paramEnc(value uint32) uint32 {
	return (value << 1) ^ (value >> 31)
}

func buildVTile() ([]byte, error) {
	tile := &vector_tile.Tile{}
	var featureID uint64
	var layerVersion uint32 = 1
	layerName := "points"
	featureType := vector_tile.Tile_POINT
	var layerExtent = vector_tile.Default_Tile_Layer_Extent
  // Put a point in the center of the tile.
	var featureGeometry = []uint32{
		moveTo(1),
		paramEnc(layerExtent / 2), // x
		paramEnc(layerExtent / 2), // y
	}
	tile.Layers = []*vector_tile.Tile_Layer{
		&vector_tile.Tile_Layer{
			Version: &layerVersion,
			Name:    &layerName,
			Extent:  &layerExtent,
			Features: []*vector_tile.Tile_Feature{
				&vector_tile.Tile_Feature{
					Id:       &featureID,
					Tags:     []uint32{},
					Type:     &featureType,
					Geometry: featureGeometry[:],
				},
			},
		},
	}
	return proto.Marshal(tile)
}

// Takes a string of the form `<z>/<x>/<y>` (for example, 1/2/3) and returns
// the individual uint32 values for x, y, and z if there was no error.
// Otherwise, err is set to a non `nil` value and x, y, z are set to 0.
func tileToXYZ(path string) (uint32, uint32, uint32, error) {
	xyzReg := regexp.MustCompile("(?P<z>[0-9]+)/(?P<x>[0-9]+)/(?P<y>[0-9]+)")
	matches := xyzReg.FindStringSubmatch(path)
	if len(matches) == 0 {
		return 0, 0, 0, errors.New("Unable to parse path as tile")
	}
	x, err := strconv.ParseUint(matches[2], 10, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	y, err := strconv.ParseUint(matches[3], 10, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	z, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return 0, 0, 0, err
	}
	return uint32(x), uint32(y), uint32(z), nil
}

func main() {
	mux := http.NewServeMux()

  // Handle requests for urls of the form `/tiles/{z}/{x}/{y}` and returns
  // the vector tile for the even tile x, y, and z coordinates.
	tileBase := "/tiles/"
	mux.HandleFunc(tileBase, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("url: %s", r.URL.Path)
		tilePart := r.URL.Path[len(tileBase):]
		_, _, _, err := tileToXYZ(tilePart)
		if err != nil {
			http.Error(w, "Invalid tile url", 400)
			return
		}
		data, err := buildVTile()
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
