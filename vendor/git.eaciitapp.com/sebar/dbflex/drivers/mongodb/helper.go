package mongodb

import (
	"reflect"
	"strings"
	"time"

	"github.com/eaciit/toolkit"
)

// This function is completely different with toolkit.ToM
// In here we use reflect to keep its original value and type
// instead of serde using json like toolkit.M do that will convert everything to text (CMIIW)
func objToM(data interface{}) (toolkit.M, error) {
	rv := reflect.Indirect(reflect.ValueOf(data))
	// Create emapty map as a result
	res := toolkit.M{}

	// Because of the difference behaviour of Struct type and Map type, we need to check the data element type
	if rv.Kind() == reflect.Struct {
		// Iterate through all the available field
		for i := 0; i < rv.NumField(); i++ {
			// Get the field type
			f := rv.Type().Field(i)

			// Initiate field name (lowercase the name similar with https://godoc.org/gopkg.in/mgo.v2/bson#Marshal)
			name := strings.ToLower(f.Name)
			// Because this is mongo driver that obviously use bson
			// So we look for bson tag
			if tag, ok := f.Tag.Lookup("bson"); ok {
				// Split tag components by ","
				tagComponents := strings.Split(tag, ",")
				// Get the fist one as a bson tag name
				bsonName := tagComponents[0]
				// If the bson tag name is "-"
				if strings.TrimSpace(bsonName) == "-" {
					// Then skip it
					continue
				} else if strings.TrimSpace(bsonName) != "" {
					// Else if bson tag name is not empty string
					// Use bson tag name as name
					name = bsonName
				}
			}

			// If the type is struct but not time.Time or is a map
			if (f.Type.Kind() == reflect.Struct && f.Type != reflect.TypeOf(time.Time{})) || f.Type.Kind() == reflect.Map {
				// Then we need to call this function again to fetch the sub value
				subRes, err := objToM(rv.Field(i).Interface())
				if err != nil {
					return nil, err
				}

				// Put the sub result into result
				res[name] = subRes

				// Skip the rest
				continue
			}

			// If the type is time.Time or is not struct and map then put it in the result directly
			res[name] = rv.Field(i).Interface()
		}

		// Return the result
		return res, nil
	} else if rv.Kind() == reflect.Map {
		// If the data element is kind of map
		// Iterate through all avilable keys
		for _, key := range rv.MapKeys() {
			// Get the map value type of the specified key
			t := rv.MapIndex(key).Elem().Type()
			// If the type is struct but not time.Time or is a map
			if (t.Kind() == reflect.Struct && t != reflect.TypeOf(time.Time{})) || t.Kind() == reflect.Map {
				// Then we need to call this function again to fetch the sub value
				subRes, err := objToM(rv.MapIndex(key).Interface())
				if err != nil {
					return nil, err
				}
				res[key.String()] = subRes

				// Skip the rest
				continue
			}

			// If the type is time.Time or is not struct and map then put it in the result directly
			res[key.String()] = rv.MapIndex(key).Interface()
		}

		// Return the result
		return res, nil
	}

	// If the data element is not map or struct then return error
	return nil, toolkit.Errorf("Expecting struct or map object but got: %s", rv.Kind())
}
