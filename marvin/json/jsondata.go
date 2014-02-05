package jsondata

import (
    "json"
    "fmt"
    "reflect"
    "strconv"
    "strings"
)

type JSONString []byte

type JSONMap struct {
    mapp reflect.Value
}

type JSONSlice struct {
    slice reflect.Value
}

func NewJSONSlice(val reflect.Value) *JSONSlice {
    var slice *JSONSlice = nil
    
    if val.IsValid() && !val.IsNil() {
        if val.Kind() == reflect.Interface {
            val = val.Elem()
        }
        
        if val.Kind() == reflect.Slice {
            //fmt.Printf("slice len: %d\n", val.Len())
            slice = &JSONSlice{val}
        }
    }
    return slice
}

func NewJSONMap(val reflect.Value) *JSONMap {
    var newMap *JSONMap = nil
    
    if val.IsValid() && !val.IsNil() {
        if val.Kind() == reflect.Interface {
            val = val.Elem()
        }
        
        if val.Kind() == reflect.Map {
            newMap = &JSONMap{val}
        }
    }
    return newMap
}


func (jsonObj *JSONMap) GetString(name string) (s string, r bool) {
    /*defer func() {
        if x := recover(); x != nil {
            fmt.Printf("recovered from panic in GetString : %v\n", x)
            s = ""
            r = false
        }
	}()*/
	
	s = ""
	r = false
	
	val := jsonObj.mapp.MapIndex(reflect.ValueOf(name))
    if val.IsValid() && !val.IsNil() {
        if val.Kind() == reflect.Interface {
            s = val.Elem().String()
            r = true
        }
    }
    return strings.TrimSpace(s), r
}

func (jsonObj *JSONMap) GetUInt(name string) (i int, r bool) {
	i = 0
    r = false
    val := jsonObj.mapp.MapIndex(reflect.ValueOf(name));
    
    if val.IsValid() && !val.IsNil() {
        if val.Kind() == reflect.Interface {
            if v, err := strconv.Atoui(val.Elem().String()); err == nil {
                i = int(v);
                r = true
            } else if val.Elem().Kind() == reflect.Uint {
                i = int(val.Elem().Uint())
                r = true
            } else if val.Elem().Kind() == reflect.Int {
                i = int(val.Elem().Int())
                r = true
            } else if val.Elem().Kind() == reflect.Float64 || val.Elem().Kind() == reflect.Float32 {
                i = int(val.Elem().Float())
                r = true
            }
        }
    }
    
    return i, r
}

func (jsonObj *JSONMap) GetSlice(name string) (o *JSONSlice, r bool) {
	o = nil
	r = false
	
    val := jsonObj.mapp.MapIndex(reflect.ValueOf(name));
    o = NewJSONSlice(val)
    
    return o, o != nil
}

func (jsonObj *JSONMap) GetMap(name string) (o *JSONMap, r bool) {
	o = nil
	r = false
	
    val := jsonObj.mapp.MapIndex(reflect.ValueOf(name));
    o = NewJSONMap(val)
    
    return o, o != nil
}

func (slice *JSONSlice) Len() int { return slice.slice.Len() }

func (slice *JSONSlice) GetString(i int) string {
    return slice.slice.Index(i).Elem().String()
}

func UnmarshalJSON(data []byte) *JSONMap {
    var rep interface{}
    var res *JSONMap = nil
    
    err := json.Unmarshal(data, &rep)
    if err == nil {
        t := reflect.TypeOf(rep)
        switch t.Kind() {
        case reflect.Map :
            res = NewJSONMap(reflect.ValueOf(rep))
        }
    } else {
        fmt.Printf("\n[UnmarshalJSON] err:--> %v\n", err)
        fmt.Printf("[UnmarshalJSON] data:-->%v\n", rep)
    }
    
    return res
}

