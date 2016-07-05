# A simple example of serving MapboxGL vector tiles in Go

## Installation

To install, ensure github.com/golang/protobuf/proto is installed and available on your $GOPATH.

## To run the project

`cd` into the project directory, then run:

    go run main.go

## To view the tiles

To view the tiles, you'll need to modify your MapboxGL style to add the additional vector tile layer. Here's an example:

```
var map = new mapboxgl.Map({
  container: 'map',
  zoom: 12.5,
  center: [-122.45, 37.79],
  style: 'mapbox://styles/mapbox/basic-v8',
  hash: false
});

map.on('load', function loaded() {
  map.addSource('custom-go-vector-tile-source', {
      type: 'vector',
      tiles: ['http://localhost:8080/tiles/{z}/{x}/{y}']
  });
  map.addLayer({
      "id": "custom-go-vector-tile-layer",
      "type": "circle",
      "source": "custom-go-vector-tile-source",
      "source-layer": "points",
      paint: {
        'circle-radius': 10,
        'circle-color': 'red',
        'circle-opacity': 1
      }
  });
});
```
