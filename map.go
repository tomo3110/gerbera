package gerbera

type ConvertToMap interface {
	ToMap() Map
}

type Map map[string]interface{}

func (m Map) Get(key string) interface{} {
	result, ok := m[key]
	if ok {
		return result
	}
	return ""
}
