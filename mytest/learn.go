package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Player struct {
	string
	int
}

func (p *Player) Name() string {
	return p.string
}

func ArrayTest() {
	ar := make([]string, 10)
	for i := 0; i < 10; i++ {
		ar[i] = strconv.Itoa(i)
	}
	ShowArray(ar)

	pos := 3

	arr := append(ar[:pos], ar[pos+1:]...)

	log.Println("--------------------")

	ShowArray(arr)

}

func ShowArray(ar []string) {
	for i, s := range ar {
		log.Printf("ar[%d]=%s", i, s)
	}
}

func TestPointer() {
	name := string("hui")
	ar := []string{name, "an", "lei"}
	log.Printf("1 name=%s \n", name)
	n := &name
	*n = "hhh"
	log.Printf("2 name=%s \n", name)
	ShowArray(ar)
	log.Println("--------------------")
	for i, s := range ar {
		log.Printf("ar[%d]=%s", i, s)
	}
}

var flagMap map[uint32]*bool = make(map[uint32]*bool, 10)

var flagVMap map[uint32]bool = make(map[uint32]bool, 10)

func PutFlag(key uint32, v *bool) {
	flagMap[key] = v
}

func PutVFlag(key uint32, v bool) {
	flagVMap[key] = v
}

func TestMapPoint() {
	key := uint32(1)
	value := true
	k2 := uint32(2)
	v2 := false
	PutFlag(key, &value)
	PutFlag(k2, &v2)
	value = false
	log.Printf("Point transfer k1=%d,v1=%v", 1, *flagMap[1])
}

func TestMapVPoint() {
	key := uint32(1)
	value := true
	k2 := uint32(2)
	v2 := false
	PutVFlag(key, value)
	PutVFlag(k2, v2)
	value = false
	log.Printf("Value transfer k1=%d,v1=%v", 1, flagVMap[1])
}

type ESConditionType int32

const (
	ESConditionType_ConditionType_BuildingLevel ESConditionType = 0
	ESConditionType_ConditionType_BlockPass     ESConditionType = 1
	// 请在MaxId之前添加
	ESConditionType_ConditionType_MaxId ESConditionType = 99
)

func (x ESConditionType) Number() int32 {
	return int32(x)
}

type Obj struct {
	ctype *ESConditionType
}

type Inter interface {
	Name() string
}

type Hal struct {
	n string
}

func (e *Hal) Name() string {
	return e.n
}

// -------------
func TestJsonDecode() {
	var jsonBlob = []byte(` [ 
        { "Name" : "Platypus" , "Order" : "Monotremata" } , 
        { "Name" : "Quoll" ,     "Order" : "Dasyuromorphia" } 
    ] `)
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	err := json.Unmarshal(jsonBlob, &animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animals)
}

func TestJsonEndode() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	b, err := json.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("b len =%d  \n",len(b))
	os.Stdout.Write(b)
}

func main() {
	TestJsonDecode()
	fmt.Println("===============")
	TestJsonEndode()
}
