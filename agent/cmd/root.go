package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	agentConfig "agent/internal/config"
	"agent/internal/scan"
	agentWS "agent/internal/ws"
)

var (
	cfgFile string
	agentID string
	wsURL   string
)

var rootCmd = &cobra.Command{
	Use:   "go-agent",
	Short: "Unified E2E POC — Agent CLI",
	Long:  "Agent connects to the control plane via WebSocket and executes scan commands.",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agent and connect to the control plane",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().StringVar(&cfgFile, "config", "config.yaml", "path to config file")
	startCmd.Flags().StringVar(&agentID, "id", "", "agent ID (overrides config)")
	startCmd.Flags().StringVar(&wsURL, "cp-url", "", "control plane WebSocket URL (overrides config)")
	rootCmd.AddCommand(startCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := agentConfig.Load(cfgFile)
	if err != nil {
		return err
	}

	// CLI flags override config file
	if agentID != "" {
		cfg.Agent.ID = agentID
	}
	if wsURL != "" {
		cfg.ControlPlane.WSURL = wsURL
	}

	if cfg.Agent.ID == "" {
		return fmt.Errorf("agent ID is required (use --id or set agent.id in config.yaml)")
	}

	log.Printf("[agent] starting with id=%s cp=%s", cfg.Agent.ID, cfg.ControlPlane.WSURL)

	scanner := scan.New(cfg.Rclone.BinPath)

	// onRunScan is called by the WS client whenever control plane sends RUN_SCAN
	var wsClient *agentWS.Client
	onRunScan := func(scanID, sourcePath string) {
		log.Printf("[agent] starting scan %s on path %s", scanID, sourcePath)

		result, err := scanner.Scan(sourcePath, func(filesScanned int) {
			wsClient.SendScanProgress(scanID, filesScanned)
		})

		if err != nil {
			log.Printf("[agent] scan %s failed: %v", scanID, err)
			wsClient.SendScanFailed(scanID, err.Error())
			return
		}

		// Convert to WS tree payload
		tree := make([]map[string]any, 0, len(result.Entries))
		for _, e := range result.Entries {
			tree = append(tree, map[string]any{
				"path":    e.Path,
				"isDir":   e.IsDir,
				"size":    e.Size,
				"modTime": e.ModTime,
			})
		}

		stats := map[string]int{
			"totalFiles":   result.TotalFiles,
			"totalFolders": result.TotalFolders,
		}

		log.Printf("[agent] scan %s complete: %d files, %d folders",
			scanID, result.TotalFiles, result.TotalFolders)
		wsClient.SendScanComplete(scanID, stats, tree)
	}

	wsClient = agentWS.NewClient(cfg.Agent.ID, cfg.ControlPlane.WSURL, onRunScan)
	wsClient.Connect() // blocks with reconnect loop
	return nil
}
