package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type Location struct {
	DisplayName string            `json:"display_name"`
	Address     map[string]string `json:"address"`
}

func main() {
	var path = flag.String("path", ".", "path to directory")
	var flatten = flag.Bool("flatten", false, "flatten nested directories")
	flag.Parse()

	files, err := ioutil.ReadDir(*path)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if *flatten {
		fmt.Println("flatten!")
		Flatten(path, path, files)
	} else {
		fmt.Println("organize!")
		Organize(path, files)
	}
}

func Flatten(currPath *string, finalPath *string, files []os.FileInfo) {
	// must be singlethreaded per openstreetmap rate limit
	for _, f := range files {
		dir := fmt.Sprint(*currPath, "/", f.Name())
		fi, err := os.Stat(dir)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			fmt.Println("-+-+-+-+-")
			fmt.Println(dir)
			subFiles, err := ioutil.ReadDir(dir)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			//flattenDirs(path, f.Name())
			Flatten(&dir, finalPath, subFiles)
		default:
			flattenDirs(&dir, finalPath, f.Name())
		}
	}

}
func Organize(path *string, files []os.FileInfo) {
	// must be singlethreaded per openstreetmap rate limit
	for _, f := range files {
		pic := fmt.Sprint(*path, "/", f.Name())
		fi, err := os.Stat(pic)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		switch mode := fi.Mode(); {
		case mode.IsRegular():
			fmt.Println("-+-+-+-+-")
			fmt.Println(pic)
			lat, lng, err := getLatLon(pic)
			if err != nil {
				fmt.Println(err.Error())
				makeDirs(path, f.Name(), map[string]string{})
				continue
			}
			location, err := getLocation(lat, lng)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(location.Address)
			makeDirs(path, f.Name(), location.Address)
		}

		// to prevent exceeding openstreetmap rate limit :-|
		time.Sleep(time.Second)
	}
}

func flattenDirs(currPath *string, destinationPath *string, fName string) {
	err := os.Rename(*currPath, fmt.Sprint(*destinationPath, "/", fName))
	if err != nil {
		fmt.Println(err.Error())
	}
}

// big and ugly
func makeDirs(path *string, fName string, address map[string]string) {
	if country, ok := address["country"]; ok {
		cpath := fmt.Sprint(*path, "/", country)
		if _, err := os.Stat(cpath); os.IsNotExist(err) {
			os.Mkdir(cpath, 0755)
		}
		if state, ok := address["state"]; ok {
			spath := fmt.Sprint(cpath, "/", state)
			if _, err := os.Stat(spath); os.IsNotExist(err) {
				os.Mkdir(spath, 0755)
			}
			if village, ok := address["village"]; ok {
				vpath := fmt.Sprint(spath, "/", village)
				if _, err := os.Stat(vpath); os.IsNotExist(err) {
					os.Mkdir(vpath, 0755)
				}
				err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(vpath, "/", fName))
				if err != nil {
					fmt.Println(err.Error())
				}
			} else if town, ok := address["town"]; ok {
				tpath := fmt.Sprint(spath, "/", town)
				if _, err := os.Stat(tpath); os.IsNotExist(err) {
					os.Mkdir(tpath, 0755)
				}
				err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(tpath, "/", fName))
				if err != nil {
					fmt.Println(err.Error())
				}
			} else if city, ok := address["city"]; ok {
				cityPath := fmt.Sprint(spath, "/", city)
				if _, err := os.Stat(cityPath); os.IsNotExist(err) {
					os.Mkdir(cityPath, 0755)
				}

				if suburb, ok := address["suburb"]; ok {
					subPath := fmt.Sprint(cityPath, "/", suburb)
					if _, err := os.Stat(subPath); os.IsNotExist(err) {
						os.Mkdir(subPath, 0755)
					}

					if neighborhood, ok := address["neighbourhood"]; ok {
						npath := fmt.Sprint(subPath, "/", neighborhood)
						if _, err := os.Stat(npath); os.IsNotExist(err) {
							os.Mkdir(npath, 0755)
						}
						err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(npath, "/", fName))
						if err != nil {
							fmt.Println(err.Error())
						}
					} else {
						err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(subPath, "/", fName))
						if err != nil {
							fmt.Println(err.Error())
						}
					}
				} else if neighborhood, ok := address["neighbourhood"]; ok {
					npath := fmt.Sprint(cityPath, "/", neighborhood)
					if _, err := os.Stat(npath); os.IsNotExist(err) {
						os.Mkdir(npath, 0755)
					}
					err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(npath, "/", fName))
					if err != nil {
						fmt.Println(err.Error())
					}
				} else {
					err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(cityPath, "/", fName))
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			} else {
				err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(spath, "/", fName))
				if err != nil {
					fmt.Println(err.Error())
				}

			}

		}
	} else {

		upath := fmt.Sprint(*path, "/unknown")
		if _, err := os.Stat(upath); os.IsNotExist(err) {
			os.Mkdir(upath, 0755)
		}
		err := os.Rename(fmt.Sprint(*path, "/", fName), fmt.Sprint(upath, "/", fName))
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func getLocation(lat, lng float64) (Location, error) {
	query := fmt.Sprint("http://nominatim.openstreetmap.org/reverse?format=json&lat=", lat, "&lon=", lng)
	resp, err := http.Get(query)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var location Location
	err = json.Unmarshal(body, &location)

	return location, err
}

func getLatLon(path string) (float64, float64, error) {
	pic, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		return 0, 0, err
	}

	x, err := exif.Decode(pic)
	if err != nil {
		fmt.Println(err.Error())
		return 0, 0, err
	}
	return x.LatLong()
}
