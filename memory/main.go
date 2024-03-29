package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	//"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

var (
	OutputFile string
	Interval   int64
)

var (
	curveCmd = &cobra.Command{
		Use:  "curve",
		RunE: PrintCurve2,
	}
)

func main() {
	rootCmd := &cobra.Command{
		Use: "mem",
		Run: rootRun,
	}

	rootCmd.PersistentFlags().StringVarP(&OutputFile, "output", "o", "mem.csv", "result output")
	rootCmd.PersistentFlags().Int64VarP(&Interval, "interval", "n", 1, "sample interval")

	rootCmd.AddCommand(curveCmd)

	rootCmd.Execute()
}

func rootRun(cmd *cobra.Command, args []string) {
	fi, _ := os.Stat(OutputFile)
	if fi != nil && fi.IsDir() {
		log.Fatalf("output(%s) is dir", OutputFile)
	}

	needHead := fi == nil
	f, err := os.Create(OutputFile)
	CheckErr(err)
	defer f.Close()
	f.Seek(0, io.SeekEnd)

	w := csv.NewWriter(f)
	w.Comma = ','

	if needHead {
		w.Write([]string{
			"date",
			"userTotal",
			"python",
			"chrome",
			"total",
			"used",
			"free",
			"shared",
			"buffers",
			"cache",
			"bufferscache",
			"available",
			"swapTotal",
			"swapUsed",
			"swapFree",
			"tmp",
		})
	}
	w.Flush()

	var now time.Time
	for {
		now = time.Now()
		var ls []string
		ls = append(ls, now.Format("2006-01-02 15:04:05"))

		GetPsInfo(&ls)
		GetMemInfo(&ls)
		GetTmp(&ls)

		w.Write(ls)
		w.Flush()

		time.Sleep(time.Duration(Interval) * time.Second)
	}
}

type PsInfo struct {
	Total  uint64
	Python uint64
	Chrome uint64
}

func GetPsInfo(ls *[]string) {
	i := &PsInfo{}
	defer func() {
		*ls = append(*ls,
			fmt.Sprintf("%d", i.Total),
			fmt.Sprintf("%d", i.Python),
			fmt.Sprintf("%d", i.Chrome),
		)
	}()

	ps, err := process.Processes()
	if err != nil {
		log.Println(err)

		return
	}

	for _, p := range ps {
		mstat, err := p.MemoryInfo()
		if mstat == nil { // maybe get nil
			fmt.Println(err)
			fmt.Println(p.Cmdline())
			continue
		}
		if mstat.RSS == 0 {
			continue
		}
		CheckErr(err)
		cmdline, err := p.Cmdline()
		CheckErr(err)

		// if p.Pid == 2 {
		// 	// kernel task = empty cmdline and mstat is all 0
		// 	fmt.Println(p)
		// 	fmt.Println(p.MemoryInfo())
		// 	fmt.Println(p.Cmdline())
		// }

		// mstat.RSS include so, 未均摊
		i.Total += Uint64ToMB(mstat.RSS)
		if strings.HasPrefix(cmdline, "python") {
			i.Python += Uint64ToMB(mstat.RSS)
		}
		if strings.HasPrefix(cmdline, "/opt/google/chrome/chrome") {
			i.Chrome += Uint64ToMB(mstat.RSS)
		}
	}
}

type MemInfo struct {
	Total        uint64
	Used         uint64
	Free         uint64
	Shared       uint64
	Buffers      uint64
	Cache        uint64
	BuffersCache uint64
	Available    uint64
	SwapTotal    uint64
	SwapUsed     uint64
	SwapFree     uint64
}

func GetMemInfo(ls *[]string) {
	i := &MemInfo{}
	defer func() {
		*ls = append(*ls,
			fmt.Sprintf("%d", i.Total),
			fmt.Sprintf("%d", i.Used),
			fmt.Sprintf("%d", i.Free),
			fmt.Sprintf("%d", i.Shared),
			fmt.Sprintf("%d", i.Buffers),
			fmt.Sprintf("%d", i.Cache),
			fmt.Sprintf("%d", i.BuffersCache),
			fmt.Sprintf("%d", i.Available),
			fmt.Sprintf("%d", i.SwapTotal),
			fmt.Sprintf("%d", i.SwapUsed),
			fmt.Sprintf("%d", i.SwapFree),
		)
	}()

	out, err := CmdExec("free", []string{"-m", "-w"})
	if err != nil {
		log.Println("free", err)

		return
	}

	var tmp string
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		tmp = sc.Text()

		if strings.HasPrefix(tmp, "Mem:") {
			tmp = strings.TrimPrefix(tmp, "Mem:")
			ls := strings.Fields(tmp)

			i.Total = ParseStrToUint64(ls[0])
			i.Used = ParseStrToUint64(ls[1])
			i.Free = ParseStrToUint64(ls[2])
			i.Shared = ParseStrToUint64(ls[3])
			i.Buffers = ParseStrToUint64(ls[4])
			i.Cache = ParseStrToUint64(ls[5])
			i.Available = ParseStrToUint64(ls[6])
		}
		if strings.HasPrefix(tmp, "Swap:") {
			tmp = strings.TrimPrefix(tmp, "Swap:")
			ls := strings.Fields(tmp)

			i.SwapTotal = ParseStrToUint64(ls[0])
			i.SwapUsed = ParseStrToUint64(ls[1])
			i.SwapFree = ParseStrToUint64(ls[2])
		}
	}

	i.BuffersCache = i.Buffers + i.Cache
}

// func GetMemInfo2(ls *[]string) {
// 	i := &MemInfo{}
// 	defer func() {
// 		*ls = append(*ls,
// 			fmt.Sprintf("%d", i.Total),
// 			fmt.Sprintf("%d", i.Used),
// 			fmt.Sprintf("%d", i.Free),
// 			fmt.Sprintf("%d", i.Shared),
// 			fmt.Sprintf("%d", i.Buffers),
// 			fmt.Sprintf("%d", i.Cache),
// 			fmt.Sprintf("%d", i.BuffersCache),
// 			fmt.Sprintf("%d", i.Available),
// 			fmt.Sprintf("%d", i.SwapTotal),
// 			fmt.Sprintf("%d", i.SwapUsed),
// 			fmt.Sprintf("%d", i.SwapFree),
// 		)
// 	}()

// 	v, err := mem.VirtualMemory()
// 	if err != nil {
// 		log.Println(err)

// 		return
// 	}

// 	i.BuffersCache = i.Buffers + i.Cache
// }

func GetTmp(ls *[]string) {
	var n uint64
	defer func() {
		*ls = append(*ls,
			fmt.Sprintf("%d", n),
		)
	}()

	out, err := CmdExec("du", []string{"-s", "-m", "/tmp"}) // tmpfs , need root
	if err != nil {
		log.Println("du", err)

		return
	}

	out = strings.TrimSuffix(out, "/tmp")
	n = ParseStrToUint64(strings.TrimSpace(out))
}

func ParseStrToUint64(s string) uint64 {
	n, err := strconv.ParseUint(s, 10, 64)
	CheckErr(err)

	return n
}

func Uint64ToMB(n uint64) uint64 {
	return n / 1024 / 1024
}

func CmdExec(name string, args []string) (output string, err error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "LANG=C")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(stdoutStderr)), nil
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
