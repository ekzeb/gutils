package util

import (
	"strconv"
	"log"
	"fmt"
	"regexp"
)

func ParseInts(sids ...string) (ints []int, err error) {
	for _,s := range sids {
		int2add, e := strconv.Atoi(s)
		if e != nil {
			log.Println(fmt.Sprintf("Error %v parsing %v ", e, s))
			err = e
			return
		}
		ints = append(ints, int2add)
	}
	return
}