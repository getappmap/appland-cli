package metadata

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch"
	"strings"

	"reflect"

	"fmt"
)

var errorOmitField = fmt.Errorf("omitted field")

func getFieldName(field reflect.StructField) (string, error) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", errorOmitField
	}

	if tag != "" {
		return strings.Split(tag, ",")[0], nil
	}

	return field.Name, nil
}

func BuildPatch(op string, path string, obj interface{}) (jsonpatch.Patch, error) {
	reflection := reflect.ValueOf(obj)
	objType := reflection.Elem().Type()
	numFields := objType.NumField()
	patches := []string{fmt.Sprintf(`{"op": "%s", "path": "%s", "value": {}}`, op, path)}

	for i := 0; i < numFields; i++ {
		field := reflection.Elem().Field(i)
		if field.IsZero() || (field.Kind() == reflect.Ptr && field.IsNil()) {
			continue
		}

		fieldValue, err := json.Marshal(field.Interface())
		if err != nil {
			return nil, err
		}

		name, err := getFieldName(objType.Field(i))
		if err != nil {
			if err == errorOmitField {
				continue
			}

			return nil, err
		}

		fieldPath := fmt.Sprintf("%s/%s", path, name)
		patches = append(patches, fmt.Sprintf(`{"op": "%s", "path": "%s", "value": %s}`, op, fieldPath, string(fieldValue)))
	}

	return jsonpatch.DecodePatch([]byte("[" + strings.Join(patches, ",") + "]"))
}
