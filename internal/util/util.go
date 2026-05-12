package util

import (
	"fmt"
	"log/slog"

	"bebop831.com/filo/internal/config"
	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v4/disk"
)

var CreateColor func(a ...interface{}) string = color.New(color.FgGreen, color.Bold).SprintFunc()
var RenameColor func(a ...interface{}) string = color.New(color.FgHiYellow, color.Bold).SprintFunc()
var RemoveColor func(a ...interface{}) string = color.New(color.FgRed, color.Bold).SprintFunc()

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
	EiB                           // 1 EiB = 1024 PiB
)

func BytesToString(bytes uint64) string {
	switch {
	case bytes >= EiB:
		return fmt.Sprintf("%.2f EiB", float64(bytes)/float64(EiB))
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

func PrintIntro(cfg *config.Config) {
	PrintBanner()

	targetUsage, err := disk.Usage(cfg.TargetDir)
	if err != nil {
		slog.Error(err.Error() + " " + cfg.TargetDir)
		return
	}

	srcUsage, err := disk.Usage(cfg.SourceDir)
	if err != nil {
		slog.Error(err.Error() + " " + cfg.SourceDir)
		return
	}

	PrintConfig(cfg, srcUsage, targetUsage)
	slog.Info(fmt.Sprintf("Starting FILO watch on '%s'...", cfg.SourceDir))

}
