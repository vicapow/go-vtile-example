# An example Go app for dynamically serving MapboxGL vector tiles

![](https://cloud.githubusercontent.com/assets/583385/16578797/4cbf4d8a-4251-11e6-9f4c-75820d220405.pn)

## Installation

To install, ensure github.com/golang/protobuf/proto is installed and available on your $GOPATH.

## To run the project

`cd` into the project directory, then run:

    go run main.go

## To view the tiles

To view the tiles, you'll need to modify your MapboxGL style to add an additional vector tile layer. Here's an example:

```
var map = new mapboxgl.Map({
  container: 'map',
  zoom: 12.5,
  center: [-122.45, 37.79],
  style: {
    version: 8,
    sources: {},
    layers: []
  },
  hash: false
});

map.on('load', function loaded() {
  map.addSource('custom-go-vector-tile-source', {
      type: 'vector',
      tiles: ['http://localhost:8080/tiles/{z}/{x}/{y}']
  });
  map.addLayer({
    id: 'background',
    type: 'background',
    paint: {
      'background-color': 'white'
    }
  });
  map.addLayer({
      "id": "custom-go-vector-tile-layer",
      "type": "circle",
      "source": "custom-go-vector-tile-source",
      "source-layer": "points",
      paint: {
        'circle-radius': {
          stops: [[8, 0.1], [11, 0.5], [15, 3], [20, 20]]
        },
        'circle-color': '#e74c3c',
        'circle-opacity': 1
      }
  });
});
```

## Data from SFGov.org

https://data.sfgov.org/City-Infrastructure/Street-Tree-Map/337t-q2b4
