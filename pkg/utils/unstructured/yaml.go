package unstructured

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/pkg/errors"
	byaml "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/yaml"
)

func YamlToUnstructured(yamlObjects []byte) ([]unstructured.Unstructured, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return yaml.ToUnstructured(yamlObjects)
}

// toUnstructured is only used in StripNull, which is only used in pkg/providers/tinkerbell/hardware/csv.go:BuildHardwareYaml
// 
func toUnstructured(yamlObjects []byte) ([]unstructured.Unstructured, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return ToUnstructured(yamlObjects)
}

func UnstructuredToYaml(yamlObjects []unstructured.Unstructured) ([]byte, error) {
	// Using this CAPI util for now, not sure if we want to depend on it but it's well written
	return FromUnstructured(yamlObjects)
}

// StripNull removes all null fields from the provided yaml.
func StripNull(resources []byte) ([]byte, error) {
	uList, err := toUnstructured(resources)
	if err != nil {
		return nil, fmt.Errorf("converting yaml to unstructured: %v", err)
	}
	for _, u := range uList {
		stripNull(u.Object)
	}
	return UnstructuredToYaml(uList)
}

func stripNull(m map[string]interface{}) {
	val := reflect.ValueOf(m)
	for _, key := range val.MapKeys() {
		v := val.MapIndex(key)
		if v.IsNil() {
			delete(m, key.String())
			continue
		}
		if t, ok := v.Interface().(map[string]interface{}); ok {
			stripNull(t)
		}
	}
}

// JoinYaml takes a list of YAML files and join them ensuring
// each YAML that the yaml separator goes on a new line by adding \n where necessary.
func JoinYaml(yamls ...[]byte) []byte {
	var yamlSeparator = []byte("---")

	var cr = []byte("\n")
	var b [][]byte //nolint:prealloc
	for _, y := range yamls {
		if !bytes.HasPrefix(y, cr) {
			y = append(cr, y...)
		}
		if !bytes.HasSuffix(y, cr) {
			y = append(y, cr...)
		}
		b = append(b, y)
	}

	r := bytes.Join(b, yamlSeparator)
	r = bytes.TrimPrefix(r, cr)
	r = bytes.TrimSuffix(r, cr)

	return r
}

// FromUnstructured takes a list of Unstructured objects and converts it into a YAML.
func FromUnstructured(objs []unstructured.Unstructured) ([]byte, error) {
	var ret [][]byte //nolint:prealloc
	for _, o := range objs {
		b := bytes.NewBuffer(nil)
		ec := byaml.NewEncoder(b)
		ec.SetIndent(2)
		err := ec.Encode(o.UnstructuredContent())
		// content, err := byaml.Marshal(o.UnstructuredContent())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to marshal yaml for %s, %s/%s", o.GroupVersionKind(), o.GetNamespace(), o.GetName())
		}
		ret = append(ret, b.Bytes())
	}

	return JoinYaml(ret...), nil
}

// ToUnstructured takes a YAML and converts it to a list of Unstructured objects.
func ToUnstructured(rawyaml []byte) ([]unstructured.Unstructured, error) {
	var ret []unstructured.Unstructured

	reader := apiyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(rawyaml)))
	count := 1
	for {
		// Read one YAML document at a time, until io.EOF is returned
		b, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.Wrapf(err, "failed to read yaml")
		}
		if len(b) == 0 {
			break
		}

		var m map[string]interface{}
		if err := byaml.Unmarshal(b, &m); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal the %s yaml document: %q", util.Ordinalize(count), string(b))
		}

		var u unstructured.Unstructured
		u.SetUnstructuredContent(m)

		// Ignore empty objects.
		// Empty objects are generated if there are weird things in manifest files like e.g. two --- in a row without a yaml doc in the middle
		if u.Object == nil {
			continue
		}

		ret = append(ret, u)
		count++
	}

	return ret, nil
}
