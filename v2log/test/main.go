package main

import (
	"fmt"
	"github.com/Rehtt/Kit/util"
)

func main() {
	info, err := util.GetGitHubFileInfo("lionsoul2014", "ip2region", "data/ip2region.xdb")
	fmt.Println(info, err)
}
