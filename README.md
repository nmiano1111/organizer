# organizer
small utility I hacked out to organize my photos by location

`go install .` from root of project

`organizer -path=/path/to/photos` to run


### deps

- [goexif](https://github.com/rwcarlsen/goexif) for handling photo exif metadata
- [nominatim](http://wiki.openstreetmap.org/wiki/Nominatim) from [OpenStreetMap](https://www.openstreetmap.org/#map=5/51.500/-0.100) for reverse geocoding (finding addresses via lat/lon coordinates)
