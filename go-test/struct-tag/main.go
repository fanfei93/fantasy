package main

import (
	"fmt"
	"reflect"
)

type SelectTerms struct {
	Test string
	WhereMaps
}

type WhereMaps struct {
	Name string `form:"name" fuzzy_search:"1"`
	Age int `form:"age"`
}

func main()  {
	s := &SelectTerms{
		Test: "1",
		WhereMaps: WhereMaps {
		Name: "test",
		Age:  10,
	}}

	v := reflect.ValueOf(s.WhereMaps)

	t := reflect.TypeOf(s.WhereMaps)

	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Int:
			if v.Field(i).Int() == 0 {
				fmt.Println(t.Field(i).Name + " is zero value")
			}
		case reflect.String:
			if v.Field(i).Len() == 0 {
				fmt.Println(t.Field(i).Name + " is zero value")
			}
		}
		str := fmt.Sprintf("like %%%v%%", v.Field(i).Interface())
		fmt.Println(str)
		//fmt.Println(v.Field(i).Interface())

	}


	//fmt.Println(reflect.TypeOf(s.WhereMaps).Field(0).Tag.Get("fuzzy_search"))


	//v := reflect.ValueOf(s)
	//t := v.Type()
	//for i := 0; i < t.NumField(); i++ {
	//	//fieldType := t.Field(i)
	//	fieldValue := v.Field(i)
	//	//fmt.Println(elem.Field(i).Tag.Get("fuzzy_search"))
	//	switch fieldValue.Kind() {
	//	case reflect.Struct:
	//		fmt.Println(fieldValue.Interface())
	//		fmt.Println(reflect.TypeOf(fieldValue.Interface()).Field(0).Tag.Get("fuzzy_search"))
	//	}
	//
	//}
	//fmt.Println(elem.Field(0).Name)




}