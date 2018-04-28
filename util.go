package sdp

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
)

func SplitLines(s *string) *[]Line {
	var lines []Line
	bs := bufio.NewScanner(strings.NewReader(*s))
	for bs.Scan() {
		l := bs.Text()
		l = strings.Trim(l, " \t")
		if len(l) == 0 {
			continue
		}
		lines = append(lines, Line(l))
	}
	return &lines
}

func UInt(s string) uint {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Panicln(err)
	}
	return uint(n)
}

func SdpErrFmt(errCode, errStr string) string {
	return fmt.Sprintf("[%s:%s]", errCode, errStr)
}
