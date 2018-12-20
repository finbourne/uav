package pipeline

import (
	"fmt"
	"log"
	"reflect"
)

func merge(p1 Pipeline, p2 Pipeline) (Pipeline, error) {
	out := Pipeline{}
	var resourceTypesOK, resourcesOK bool
	out.Groups = mergeGroups(p1.Groups, p2.Groups)
	out.Jobs = appendArrayInterfaceNoCheck(p1.Jobs, p2.Jobs)
	out.Merge = appendArrayInterfaceNoCheck(p1.Merge, p2.Merge)
	out.ResourceTypes, resourceTypesOK = mergeArrayInterfaceCheckSame(p1.ResourceTypes, p2.ResourceTypes)
	out.Resources, resourcesOK = mergeArrayInterfaceCheckSame(p1.Resources, p2.Resources)
	out.extraTemplates = append(p1.extraTemplates, p2.extraTemplates...)

	if !resourceTypesOK && !resourcesOK {
		return Pipeline{}, fmt.Errorf("resourceTypes and resource merge error;  two or more items that are not identical")
	}

	if !resourceTypesOK {
		return Pipeline{}, fmt.Errorf("resourceTypes merge error; two or more resourceTypes are not identical")
	}
	if !resourcesOK {
		return Pipeline{}, fmt.Errorf("resource merge error; two or more resources are not identical")
	}

	return out, nil
}

func mergeGroups(a []interface{}, b []interface{}) []interface{} {
	out := make([]interface{}, 0)

	out = append(out, a...)

	for _, v := range b {
		name := getName(v)
		value, exists := findValue(name, out)
		if exists {
			index := findIndex(name, out)
			ngroup := interfaceToMapStringInterface(v)
			njobs := ngroup["jobs"].([]interface{})
			egroup := interfaceToMapStringInterface(value)
			for _, j := range njobs {
				egroup["jobs"] = append(egroup["jobs"].([]interface{}), j)
			}
			out[index] = mapStringInterfaceToMapInterfaceInterface(egroup)
		}
		if !exists {
			out = append(out, v)
		}
	}

	return out
}

func mapStringInterfaceToMapInterfaceInterface(data map[string]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{})
	for k, v := range data {
		out[k] = v
	}
	return out
}

func appendArrayInterfaceNoCheck(a []interface{}, b []interface{}) []interface{} {
	out := make([]interface{}, 0)

	out = append(out, a...)
	out = append(out, b...)

	return out
}

func mergeArrayInterfaceCheckSame(a []interface{}, b []interface{}) ([]interface{}, bool) {
	out := make([]interface{}, 0)

	out = append(out, b...)

	for _, v := range a {
		name := getName(v)
		value, exists := findValue(name, out)
		valuesEqual := valuesSame(value, v)
		if exists && !valuesEqual {
			return out, false
		}
		if !exists {
			out = append(out, v)
		}
	}

	return out, true
}

func getName(data interface{}) string {
	return interfaceToMapStringInterface(data)["name"].(string)
}

func interfaceToMapStringInterface(data interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	interim := data.(map[interface{}]interface{})
	for key, value := range interim {
		switch key := key.(type) {
		case string:
			m[key] = value
		default:
			log.Fatal("key should be string")
		}
	}
	return m
}

func findValue(name string, a []interface{}) (interface{}, bool) {
	for _, data := range a {
		c := interfaceToMapStringInterface(data)
		if c["name"].(string) == name {
			return data, true
		}
	}

	return nil, false
}

func findIndex(name string, a []interface{}) int {
	for i, data := range a {
		c := interfaceToMapStringInterface(data)
		if c["name"].(string) == name {
			return i
		}
	}

	return -1
}

func valuesSame(v1 interface{}, v2 interface{}) bool {
	return reflect.DeepEqual(v1, v2)
}
