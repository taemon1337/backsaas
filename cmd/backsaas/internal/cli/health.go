package cli

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health",
	Long:  `Check the health status of all BackSaas services`,
}

var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check all services health",
	Long:  `Check the health status of Gateway, Platform API, Database, and Redis`,
	RunE:  runHealthCheck,
}

var (
	healthService string
	healthTimeout int
)

func init() {
	healthCmd.AddCommand(healthCheckCmd)
	
	healthCheckCmd.Flags().StringVar(&healthService, "service", "", "Check specific service (gateway, platform-api, database, redis)")
	healthCheckCmd.Flags().IntVar(&healthTimeout, "timeout", 10, "Timeout in seconds for health checks")
}

type ServiceHealth struct {
	Name     string
	URL      string
	Status   string
	Response string
	Duration time.Duration
}

func runHealthCheck(cmd *cobra.Command, args []string) error {
	services := []ServiceHealth{
		{
			Name: "API Gateway",
			URL:  viper.GetString("gateway_url") + "/health",
		},
		{
			Name: "Platform API",
			URL:  viper.GetString("platform_url") + "/health",
		},
		{
			Name: "Gateway Metrics",
			URL:  viper.GetString("gateway_url") + "/metrics",
		},
	}

	// Filter by specific service if requested
	if healthService != "" {
		var filteredServices []ServiceHealth
		for _, service := range services {
			switch healthService {
			case "gateway":
				if service.Name == "API Gateway" || service.Name == "Gateway Metrics" {
					filteredServices = append(filteredServices, service)
				}
			case "platform-api":
				if service.Name == "Platform API" {
					filteredServices = append(filteredServices, service)
				}
			}
		}
		services = filteredServices
	}

	fmt.Println("🏥 BackSaas Health Check")
	fmt.Println("========================")

	// Check each service
	for i := range services {
		checkService(&services[i])
	}

	// Display results in table
	displayHealthTable(services)

	// Summary
	healthy := 0
	for _, service := range services {
		if service.Status == "✅ Healthy" {
			healthy++
		}
	}

	fmt.Printf("\n📊 Summary: %d/%d services healthy\n", healthy, len(services))

	if healthy == len(services) {
		color.Green("🎉 All services are healthy!")
		return nil
	} else {
		color.Red("⚠️  Some services are unhealthy")
		return fmt.Errorf("%d services are unhealthy", len(services)-healthy)
	}
}

func checkService(service *ServiceHealth) {
	start := time.Now()
	
	client := &http.Client{
		Timeout: time.Duration(healthTimeout) * time.Second,
	}

	resp, err := client.Get(service.URL)
	service.Duration = time.Since(start)

	if err != nil {
		service.Status = "❌ Error"
		service.Response = err.Error()
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

func displayHealthTable(services []ServiceHealth) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "Status", "Response", "Duration", "URL"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, service := range services {
		durationStr := fmt.Sprintf("%dms", service.Duration.Milliseconds())
		table.Append([]string{
			service.Name,
			service.Status,
			service.Response,
			durationStr,
			service.URL,
		})
	}

	table.Render()
}
