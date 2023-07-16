package main

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//"strings"
)

type Linestring struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type ScatteredLine struct {
	ID         string    `json:"id"`
	StartPoint []float64 `json:"startPoint"`
	EndPoint   []float64 `json:"endPoint"`
}

type Intersection struct {
	LineID        string    `json:"lineID"`
	Intersect []float64 `json:"intersection"`
}

func main() {
	http.HandleFunc("/intersections", handleIntersections)
	log.Fatal(http.ListenAndServe(":8092", nil))
}

func handleIntersections(w http.ResponseWriter, r *http.Request) {
	// Validate the request method
	if r.Method != http.MethodPost {
		http.Error(w,"Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate the authentication header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "Authorization" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Read and parse the long linestring from the request body
	var linestring Linestring
	err := json.NewDecoder(r.Body).Decode(&linestring)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Load the scattered lines from the file
	scatteredLines, err := loadScatteredLines("scattered_lines.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Find intersecting lines
	intersections := findIntersections(linestring, scatteredLines)

	// Prepare the response
	responseJSON, err := json.Marshal(intersections)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the response headers and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func loadScatteredLines(filename string) ([]ScatteredLine, error) {
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var scatteredLines []ScatteredLine
	err = json.Unmarshal(fileData, &scatteredLines)
	if err != nil {
		return nil, err
	}

	return scatteredLines, nil
}

func findIntersections(linestring Linestring, scatteredLines []ScatteredLine) []Intersection {
	var intersections []Intersection

	for _, line := range scatteredLines {
		for i := 1; i < len(linestring.Coordinates); i++ {
			if doLinesIntersect(line.StartPoint, line.EndPoint, linestring.Coordinates[i-1], linestring.Coordinates[i]) {
				intersection := Intersection{
					LineID:    line.ID,
					Intersect: findIntersectionPoint(line.StartPoint, line.EndPoint, linestring.Coordinates[i-1], linestring.Coordinates[i]),
				}
				intersections = append(intersections, intersection)
			}
		}
	}

	return intersections
}

func doLinesIntersect(p1, p2, q1, q2 []float64) bool {
	o1 := getOrientation(p1, p2, q1)
	o2 := getOrientation(p1, p2, q2)
	o3 := getOrientation(q1, q2, p1)
	o4 := getOrientation(q1, q2, p2)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && isOnSegment(p1, q1, p2) {
		return true
	}

	if o2 == 0 && isOnSegment(p1, q2, p2) {
		return true
	}

	if o3 == 0 && isOnSegment(q1, p1, q2) {
		return true
	}

	if o4 == 0 && isOnSegment(q1, p2, q2) {
		return true
	}

	return false
}

func getOrientation(p, q, r []float64) int {
	val := (q[1]-p[1])*(r[0]-q[0]) - (q[0]-p[0])*(r[1]-q[1])

	if val == 0 {
		return 0 // Collinear
	} else if val > 0 {
		return 1 // Clockwise
	} else {
		return 2 // Counterclockwise
	}
}

func isOnSegment(p, q, r []float64) bool {
	if q[0] <= max(p[0], r[0]) && q[0] >= min(p[0], r[0]) &&
		q[1] <= max(p[1], r[1]) && q[1] >= min(p[1], r[1]) {
		return true
	}
	return false
}

func findIntersectionPoint(p1, p2, q1, q2 []float64) []float64 {
	x1, y1 := p1[0], p1[1]
	x2, y2 := p2[0], p2[1]
	x3, y3 := q1[0], q1[1]
	x4, y4 := q2[0], q2[1]

	denominator := (x1-x2)*(y3-y4) - (y1-y2)*(x3-x4)
	if denominator == 0 {
		return nil // Parallel lines, no intersection
	}

	intersectX := ((x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)) / denominator
	intersectY := ((x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)) / denominator

	return []float64{intersectX, intersectY}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
