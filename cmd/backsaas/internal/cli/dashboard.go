package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Real-time platform dashboard",
	Long: `Display a real-time dashboard showing platform health, metrics, and status.

Similar to 'top' command but for BackSaas platform monitoring.
Press 'q' or Ctrl+C to exit.`,
	RunE: runDashboard,
}

var (
	dashboardRefresh int
	dashboardCompact bool
)

func init() {
	dashboardCmd.Flags().IntVar(&dashboardRefresh, "refresh", 2, "Refresh interval in seconds")
	dashboardCmd.Flags().BoolVar(&dashboardCompact, "compact", false, "Compact display mode")
}

type DashboardData struct {
	Timestamp    time.Time
	Services     []ServiceHealth
	SystemStats  SystemStats
	TenantStats  TenantStats
	RequestStats RequestStats
}

type SystemStats struct {
	Uptime       string
	Version      string
	Environment  string
	TotalMemory  string
	UsedMemory   string
	CPUUsage     string
	ActiveConns  int
}

type TenantStats struct {
	TotalTenants   int
	ActiveTenants  int
	TotalUsers     int
	ActiveUsers    int
	TotalSchemas   int
	RecentActivity []string
}

type RequestStats struct {
	RequestsPerSecond float64
	AvgResponseTime   time.Duration
	ErrorRate         float64
	Top5Endpoints     []EndpointStat
}

type EndpointStat struct {
	Path     string
	Count    int
	AvgTime  time.Duration
	ErrorPct float64
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Setup signal handling for graceful exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	// Hide cursor and setup terminal
	fmt.Print("\033[?25l") // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor on exit

	color.Cyan("🚀 BackSaas Platform Dashboard")
	color.Cyan("==============================")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	ticker := time.NewTicker(time.Duration(dashboardRefresh) * time.Second)
	defer ticker.Stop()

	// Initial display
	if err := displayDashboard(); err != nil {
		return err
	}

	// Main dashboard loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n👋 Dashboard stopped")
			return nil
		case <-ticker.C:
			if err := displayDashboard(); err != nil {
				color.Red("Error updating dashboard: %v", err)
			}
		}
	}
}

func displayDashboard() error {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	data, err := collectDashboardData()
	if err != nil {
		return err
	}

	// Header with timestamp
	color.Cyan("🚀 BackSaas Platform Dashboard - %s", data.Timestamp.Format("15:04:05"))
	color.Cyan("================================================================")
	fmt.Println()

	// Service Health Section
	displayServiceHealth(data.Services)
	fmt.Println()

	// System Stats Section
	displaySystemStats(data.SystemStats)
	fmt.Println()

	// Tenant Stats Section
	displayTenantStats(data.TenantStats)
	fmt.Println()

	// Request Stats Section
	displayRequestStats(data.RequestStats)
	fmt.Println()

	// Footer
	color.Yellow("Refresh: %ds | Press Ctrl+C to exit", dashboardRefresh)

	return nil
}

func collectDashboardData() (*DashboardData, error) {
	data := &DashboardData{
		Timestamp: time.Now(),
	}

	// Collect service health
	services := []ServiceHealth{
		{Name: "API Gateway", URL: viper.GetString("gateway_url") + "/health"},
		{Name: "Platform API", URL: viper.GetString("platform_url") + "/health"},
		{Name: "Gateway Metrics", URL: viper.GetString("gateway_url") + "/metrics"},
	}

	for i := range services {
		checkServiceDashboard(&services[i])
	}
	data.Services = services

	// Collect system stats (mock data for now)
	data.SystemStats = SystemStats{
		Uptime:      "2h 34m",
		Version:     "v0.7.0-dev",
		Environment: "development",
		TotalMemory: "512MB",
		UsedMemory:  "234MB",
		CPUUsage:    "12.3%",
		ActiveConns: 42,
	}

	// Collect tenant stats (mock data for now)
	data.TenantStats = TenantStats{
		TotalTenants:   2,
		ActiveTenants:  2,
		TotalUsers:     26,
		ActiveUsers:    8,
		TotalSchemas:   4,
		RecentActivity: []string{
			"User login: john@acme-corp.com",
			"Schema deployed: acme-corp/v1.2.0",
			"Tenant created: test-company",
		},
	}

	// Collect request stats (mock data for now)
	data.RequestStats = RequestStats{
		RequestsPerSecond: 23.4,
		AvgResponseTime:   45 * time.Millisecond,
		ErrorRate:         0.02,
		Top5Endpoints: []EndpointStat{
			{Path: "/api/platform/tenants", Count: 156, AvgTime: 23 * time.Millisecond, ErrorPct: 0.0},
			{Path: "/api/tenants/acme-corp/users", Count: 89, AvgTime: 67 * time.Millisecond, ErrorPct: 0.01},
			{Path: "/health", Count: 234, AvgTime: 12 * time.Millisecond, ErrorPct: 0.0},
			{Path: "/metrics", Count: 45, AvgTime: 8 * time.Millisecond, ErrorPct: 0.0},
			{Path: "/api/platform/schemas", Count: 34, AvgTime: 89 * time.Millisecond, ErrorPct: 0.03},
		},
	}

	return data, nil
}

func displayServiceHealth(services []ServiceHealth) {
	color.White("🏥 Service Health")
	color.White("================")

	healthy := 0
	for _, service := range services {
		status := service.Status
		if strings.Contains(status, "Healthy") {
			healthy++
			color.Green("✅ %-15s %s (%dms)", service.Name, "Healthy", service.Duration.Milliseconds())
		} else {
			color.Red("❌ %-15s %s", service.Name, "Unhealthy")
		}
	}

	// Overall health indicator
	if healthy == len(services) {
		color.Green("🎉 All services healthy (%d/%d)", healthy, len(services))
	} else {
		color.Yellow("⚠️  %d/%d services healthy", healthy, len(services))
	}
}

func displaySystemStats(stats SystemStats) {
	color.White("💻 System Stats")
	color.White("===============")

	fmt.Printf("Uptime: %s | Version: %s | Environment: %s\n", 
		color.GreenString(stats.Uptime), 
		color.BlueString(stats.Version), 
		color.YellowString(stats.Environment))
	
	fmt.Printf("Memory: %s / %s | CPU: %s | Connections: %d\n",
		color.CyanString(stats.UsedMemory),
		stats.TotalMemory,
		color.MagentaString(stats.CPUUsage),
		stats.ActiveConns)
}

func displayTenantStats(stats TenantStats) {
	color.White("🏢 Tenant Stats")
	color.White("===============")

	fmt.Printf("Tenants: %s active / %d total | Users: %s active / %d total | Schemas: %d\n",
		color.GreenString("%d", stats.ActiveTenants),
		stats.TotalTenants,
		color.GreenString("%d", stats.ActiveUsers),
		stats.TotalUsers,
		stats.TotalSchemas)

	if !dashboardCompact && len(stats.RecentActivity) > 0 {
		fmt.Println("\nRecent Activity:")
		for i, activity := range stats.RecentActivity {
			if i >= 3 { // Show only last 3 activities
				break
			}
			color.Cyan("  • %s", activity)
		}
	}
}

func displayRequestStats(stats RequestStats) {
	color.White("📊 Request Stats")
	color.White("================")

	fmt.Printf("RPS: %s | Avg Response: %s | Error Rate: %s\n",
		color.GreenString("%.1f", stats.RequestsPerSecond),
		color.CyanString("%dms", stats.AvgResponseTime.Milliseconds()),
		getErrorRateColor(stats.ErrorRate))

	if !dashboardCompact && len(stats.Top5Endpoints) > 0 {
		fmt.Println("\nTop Endpoints:")
		for i, endpoint := range stats.Top5Endpoints {
			if i >= 3 { // Show only top 3 in dashboard
				break
			}
			fmt.Printf("  %s %s (%d req, %dms avg)\n",
				getEndpointStatusIcon(endpoint.ErrorPct),
				color.BlueString("%-30s", endpoint.Path),
				endpoint.Count,
				endpoint.AvgTime.Milliseconds())
		}
	}
}

func getErrorRateColor(rate float64) string {
	if rate < 0.01 {
		return color.GreenString("%.2f%%", rate*100)
	} else if rate < 0.05 {
		return color.YellowString("%.2f%%", rate*100)
	} else {
		return color.RedString("%.2f%%", rate*100)
	}
}

func getEndpointStatusIcon(errorPct float64) string {
	if errorPct == 0.0 {
		return "✅"
	} else if errorPct < 0.02 {
		return "⚠️ "
	} else {
		return "❌"
	}
}

// Enhanced service check with shorter timeout for dashboard
func checkServiceDashboard(service *ServiceHealth) {
	start := time.Now()
	
	client := &http.Client{
		Timeout: 3 * time.Second, // Shorter timeout for dashboard
	}

	resp, err := client.Get(service.URL)
	service.Duration = time.Since(start)

	if err != nil {
		service.Status = "❌ Error"
		service.Response = truncateError(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		service.Status = "✅ Healthy"
		service.Response = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		service.Status = "⚠️  Unhealthy"
		service.Response = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
}

func truncateError(err string) string {
	if len(err) > 50 {
		return err[:47] + "..."
	}
	return err
}
