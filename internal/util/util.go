package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	git "github.com/go-git/go-git/v5"
)

var errorOmitField = fmt.Errorf("omitted field")

type RepositoryInfo struct {
	Path       string
	Repository *git.Repository
}

func GetRepository(filePath string) (*RepositoryInfo, error) {
	currentPath := path.Clean(filePath)
	for {
		if currentPath == "." || currentPath == "/" {
			break
		}

		file, err := os.Stat(currentPath)
		if err != nil {
			return nil, err
		}

		if file.IsDir() {
			repo, err := git.PlainOpen(currentPath)
			if err != nil && err != git.ErrRepositoryNotExists {
				return nil, err
			}

			if repo != nil {
				return &RepositoryInfo{
					Path:       currentPath,
					Repository: repo,
				}, nil
			}
		}

		currentPath = path.Dir(currentPath)
	}

	return nil, git.ErrRepositoryNotExists
}

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
