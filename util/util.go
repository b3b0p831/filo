package util

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"bebop831.com/filo/config"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v4/disk"
)

var Mu sync.Mutex
var re regexp.Regexp = *regexp.MustCompile(`^\d+[smh]$`)

// / ===== AI Generated =======
func PrintBanner() {
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	white := color.New(color.FgWhite, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(blue(" ███████╗██╗██╗      ██████╗ "))
	fmt.Println(blue(" ██╔════╝██║██║     ██╔═══██╗"))
	fmt.Println(blue(" ███████╗██║██║     ██║   ██║"))
	fmt.Println(blue(" ██╔════╝██║██║     ██║   ██║"))
	fmt.Println(blue(" ██║     ██║███████╗╚██████╔╝"))
	fmt.Println(blue(" ╚═╝     ╚═╝╚══════╝ ╚═════╝ "))
	fmt.Println()
	fmt.Println(white("         ") + yellow("先入後出 同期"))
	fmt.Println()
}

// ============================

const (
	_          = iota             // iota generates simple counter starting 0 that ++ for each each declaration in const
	KiB uint64 = 1 << (10 * iota) // 1 KiB = 1024 B
	MiB                           // 1 MiB = 1024 KiB
	GiB                           // 1 GiB = 1024 MiB
	TiB                           // 1 TiB = 1024 GiB
	PiB                           // 1 PiB = 1024 TiB
)

func BytesToString(bytes uint64) string {
	switch {
	case bytes >= PiB:
		return fmt.Sprintf("%.2f PiB", float64(bytes)/float64(PiB))
	case bytes >= TiB:
		return fmt.Sprintf("%.2f TiB", float64(bytes)/float64(TiB))
	case bytes >= GiB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/float64(MiB))
	case bytes >= KiB:
		return fmt.Sprintf("%.2f KiB", float64(bytes)/float64(KiB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func PrintConfig(cfg *config.Config, srcUsage *disk.UsageStat, targetUsage *disk.UsageStat) {
	// Define some reusable colors
	header := color.New(color.FgCyan, color.Bold).SprintFunc()
	label := color.New(color.FgWhite, color.Bold).SprintFunc()
	value := color.New(color.FgGreen).SprintFunc()
	warn := color.New(color.FgYellow).SprintFunc()

	fmt.Println(header("========== FILO Configuration =========="))
	fmt.Printf("%s %s\n", label(" Target Dir :"), value(cfg.TargetDir))
	fmt.Printf("%s %s\n", label(" Used Space :"), warn(BytesToString(targetUsage.Used)))
	fmt.Printf("%s %s\n", label(" Free Space :"), value(BytesToString(targetUsage.Free)))
	fmt.Printf("%s %s\n", label(" Total Size :"), value(BytesToString(targetUsage.Free+targetUsage.Used)))
	fmt.Println(header("---------------------------------------------"))
	fmt.Printf("%s %s\n", label(" Source Dir :"), value(cfg.SourceDir))
	fmt.Printf("%s %s\n", label(" Used Space :"), warn(BytesToString(srcUsage.Used)))
	fmt.Printf("%s %s\n", label(" Free Space :"), value(BytesToString(srcUsage.Free)))
	fmt.Printf("%s %s\n", label(" Total Size :"), value(BytesToString(srcUsage.Free+srcUsage.Used)))
	fmt.Println(header("---------------------------------------------"))
	fmt.Printf("%s %.2f\n", label(" Max Fill   :"), cfg.MaxFill)
	fmt.Printf("%s %s\n", label(" Sync Delay :"), cfg.SyncDelay)
	fmt.Printf("%s %s\n", label(" Log Level  :"), value(cfg.LogLevel))
	fmt.Println(header("============================================="))
}

func GetTimeInterval(interval string) (time.Duration, error) {
	if !re.Match([]byte(interval)) {
		return 0, fmt.Errorf("interval string does not match format (i.e 1s, 3m, 5h): %v", interval)
	}

	timeValStr, lastChar := interval[:len(interval)-1], interval[len(interval)-1]
	timeVal, err := strconv.ParseInt(timeValStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("util/util.go: unable to ParseInt(timevalStr)")
	}

	switch lastChar {
	case 's':
		return time.Duration(timeVal) * time.Second, nil
	case 'm':
		return time.Duration(timeVal) * time.Minute, nil
	case 'h':
		return time.Duration(timeVal) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid time unit: %v", lastChar)
	}

}
