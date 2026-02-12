// 生成与 casbin-pgx-adapter 一致的策略 id（用于手动 INSERT 到 casbin_rule 时）。
// 用法：go run ./app/cmd/genpolicyid p admin '*' '*'  → 策略 p, admin, *, *
//       go run ./app/cmd/genpolicyid g alice admin   → 策略 g, alice, admin
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mmcloughlin/meow"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: genpolicyid <ptype> <v0> [v1] [v2] ...\n例: genpolicyid p admin '*' '*'\n")
		os.Exit(1)
	}
	ptype := os.Args[1]
	rule := os.Args[2:]
	data := strings.Join(append([]string{ptype}, rule...), ",")
	sum := meow.Checksum(0, []byte(data))
	id := fmt.Sprintf("%x", sum)
	fmt.Println(id)
}
