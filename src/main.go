package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type iface struct {
	name string
	vrf  string
	addr string
	bw   string
	desc string
	shut string
}

func (i *iface) show() {
	fmt.Printf("%25s,%15s,%15s,%10s,%8s,\"%s\"\n", i.name, i.vrf, i.addr, i.bw, i.shut, i.desc)
}

type scanner struct {
	lineCount   int
	table       map[string]*iface
	currIface   *iface
	multilink   int
	loopback    int
	portChannel int
}

func parseLine(ctx *scanner, rawLine string) {

	line := strings.TrimRight(rawLine, "\r\n")

	if line == "" {
		return
	}

	if strings.HasPrefix(line, "interface ") {

		// get interface name
		tail := line[10:]
		fields := strings.Fields(tail)
		if len(fields) < 1 {
			fmt.Printf("parseLine: line=%d bad interface name: [%s]\n", ctx.lineCount, line)
			return
		}
		name := fields[0]

		// find interface
		i, found := ctx.table[name]
		if !found {
			// create interface
			i = &iface{name: name}
			ctx.table[name] = i
		}

		ctx.currIface = i

		if strings.HasPrefix(name, "Multi") {
			ctx.multilink++
		}
		if strings.HasPrefix(name, "Loop") {
			ctx.loopback++
		}
		if strings.HasPrefix(name, "Port") {
			ctx.portChannel++
		}

		return
	}

	if ctx.currIface == nil {
		return
	}

	if line[0] == '!' {
		ctx.currIface = nil
		return
	}

	if strings.HasPrefix(line, " ip vrf forwarding ") {
		vrf := strings.TrimSpace(line[19:])

		if ctx.currIface.vrf != "" {
			fmt.Printf("parseLine: line=%d vrf redefinition old=%s new=%s: [%s]\n", ctx.lineCount, ctx.currIface.vrf, vrf, line)
		}

		ctx.currIface.vrf = vrf

		return
	}

	if strings.HasPrefix(line, " ip address ") {

		if strings.HasSuffix(line, "secondary") {
			return
		}

		tail := line[12:]
		fields := strings.Fields(tail)
		addr := fields[0]

		if ctx.currIface.addr != "" {
			fmt.Printf("parseLine: line=%d addr redefinition old=%s new=%s: [%s]\n", ctx.lineCount, ctx.currIface.addr, addr, line)
		}

		ctx.currIface.addr = addr

		return
	}

	if strings.HasPrefix(line, " description ") {
		desc := strings.TrimSpace(line[13:])

		if ctx.currIface.desc != "" {
			fmt.Printf("parseLine: line=%d desc redefinition old=%s new=%s: [%s]\n", ctx.lineCount, ctx.currIface.desc, desc, line)
		}

		ctx.currIface.desc = desc

		return
	}

	if strings.HasPrefix(line, " bandwidth ") {
		bw := strings.TrimSpace(line[11:])

		if ctx.currIface.bw != "" {
			fmt.Printf("parseLine: line=%d bw redefinition old=%s new=%s: [%s]\n", ctx.lineCount, ctx.currIface.bw, bw, line)
		}

		ctx.currIface.bw = bw

		return
	}

	if strings.HasPrefix(line, " shutdown") {

		if ctx.currIface.shut != "" {
			fmt.Printf("parseLine: line=%d shutdown redefinition: [%s]\n", ctx.lineCount, line)
		}

		ctx.currIface.shut = "shutdown"

		return
	}

}

func scan() (map[string]*iface, error) {
	ctx := scanner{table: map[string]*iface{}}

	input := os.Stdin
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("scan: error reading lines: %v\n", err)
		}
		ctx.lineCount++
		parseLine(&ctx, line)
	}

	fmt.Printf("scan: %d lines, %d interfaces\n", ctx.lineCount, len(ctx.table))
	fmt.Printf("scan: %d multilink, %d loopback, %d port-channel\n", ctx.multilink, ctx.loopback, ctx.portChannel)

	return ctx.table, nil
}

func show(table map[string]*iface) {
	for _, i := range table {
		i.show()
	}
}

func main() {
	fmt.Printf("main: reading input from stdin\n")
	table, err := scan()
	fmt.Printf("main: reading input from stdin -- done\n")
	if err != nil {
		panic(fmt.Sprintf("main: error: %v", err))
	}
	show(table)
	fmt.Printf("main: end\n")
}
